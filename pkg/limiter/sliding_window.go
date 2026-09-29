package limiter

import (
	"context"
	"fmt"
	"time"

	limiterRedis "github.com/distributed-systems/rate-limiter/pkg/redis"
)

type SlidingWindowLimiter struct {
	client *limiterRedis.Client
}

func NewSlidingWindowLimiter(client *limiterRedis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{client: client}
}

func (s *SlidingWindowLimiter) Allow(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	nowMs := start.UnixMilli()

	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	windowSizeMs := req.WindowSize.Milliseconds()
	if windowSizeMs <= 0 {
		windowSizeMs = 1000 // default 1 second
	}

	limit := req.Capacity
	if limit <= 0 {
		limit = int64(req.RatePerSecond)
		if limit <= 0 {
			limit = 100
		}
	}

	if req.Algorithm == AlgorithmSlidingWindowLog {
		// Exact Sliding Window using Redis Sorted Set
		redisKey := fmt.Sprintf("rl:swl:%s", req.Key)
		nonce := fmt.Sprintf("%d_%d", nowMs, time.Now().UnixNano()%1000000)

		res, err := s.client.EvalSlidingWindowLog(ctx, redisKey, limit, windowSizeMs, cost, nowMs, nonce)
		if err != nil {
			return nil, fmt.Errorf("sliding window log eval error: %w", err)
		}

		allowedInt := res[0].(int64)
		remaining := res[1].(int64)
		retryAfterMs := res[2].(int64)
		lim := res[3].(int64)
		resetMs := res[4].(int64)

		return &Result{
			Allowed:      allowedInt == 1,
			Remaining:    remaining,
			Limit:        lim,
			RetryAfter:   time.Duration(retryAfterMs) * time.Millisecond,
			ResetAfter:   time.Duration(resetMs) * time.Millisecond,
			Algorithm:    AlgorithmSlidingWindowLog,
			CacheTier:    "L2_REDIS",
			LatencyNanos: time.Since(start).Nanoseconds(),
		}, nil
	}

	// Hybrid Sliding Window Counter (O(1) memory & time)
	redisKey := fmt.Sprintf("rl:swc:%s", req.Key)
	res, err := s.client.EvalSlidingWindowCounter(ctx, redisKey, limit, windowSizeMs, cost, nowMs)
	if err != nil {
		return nil, fmt.Errorf("sliding window counter eval error: %w", err)
	}

	allowedInt := res[0].(int64)
	remaining := res[1].(int64)
	retryAfterMs := res[2].(int64)
	lim := res[3].(int64)
	resetMs := res[4].(int64)

	return &Result{
		Allowed:      allowedInt == 1,
		Remaining:    remaining,
		Limit:        lim,
		RetryAfter:   time.Duration(retryAfterMs) * time.Millisecond,
		ResetAfter:   time.Duration(resetMs) * time.Millisecond,
		Algorithm:    AlgorithmSlidingWindowCounter,
		CacheTier:    "L2_REDIS",
		LatencyNanos: time.Since(start).Nanoseconds(),
	}, nil
}

func (s *SlidingWindowLimiter) GetStatus(ctx context.Context, key string, alg Algorithm) (*Result, error) {
	return &Result{
		Allowed:   true,
		Algorithm: alg,
		CacheTier: "L2_REDIS",
	}, nil
}

func (s *SlidingWindowLimiter) Reset(ctx context.Context, key string, alg Algorithm) error {
	var prefix string
	if alg == AlgorithmSlidingWindowLog {
		prefix = fmt.Sprintf("rl:swl:%s", key)
	} else {
		prefix = fmt.Sprintf("rl:swc:%s*", key)
	}

	if alg == AlgorithmSlidingWindowLog {
		return s.client.Raw().Del(ctx, prefix).Err()
	}

	// Delete all matching buckets for sliding window counter
	iter := s.client.Raw().Scan(ctx, 0, prefix, 0).Iterator()
	for iter.Next(ctx) {
		_ = s.client.Raw().Del(ctx, iter.Val()).Err()
	}
	return iter.Err()
}

func (s *SlidingWindowLimiter) Name() string {
	return "sliding_window"
}
