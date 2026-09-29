package limiter

import (
	"context"
	"fmt"
	"time"

	limiterRedis "github.com/distributed-systems/rate-limiter/pkg/redis"
)

type LeakyBucketLimiter struct {
	client *limiterRedis.Client
}

func NewLeakyBucketLimiter(client *limiterRedis.Client) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{client: client}
}

func (l *LeakyBucketLimiter) Allow(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	redisKey := fmt.Sprintf("rl:lb:%s", req.Key)
	nowMs := start.UnixMilli()

	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	res, err := l.client.EvalLeakyBucket(ctx, redisKey, req.Capacity, req.RatePerSecond, cost, nowMs)
	if err != nil {
		return nil, fmt.Errorf("leaky bucket eval error: %w", err)
	}

	allowedInt := res[0].(int64)
	remaining := res[1].(int64)
	retryAfterMs := res[2].(int64)
	limit := res[3].(int64)
	resetMs := res[4].(int64)

	return &Result{
		Allowed:      allowedInt == 1,
		Remaining:    remaining,
		Limit:        limit,
		RetryAfter:   time.Duration(retryAfterMs) * time.Millisecond,
		ResetAfter:   time.Duration(resetMs) * time.Millisecond,
		Algorithm:    AlgorithmLeakyBucket,
		CacheTier:    "L2_REDIS",
		LatencyNanos: time.Since(start).Nanoseconds(),
	}, nil
}

func (l *LeakyBucketLimiter) GetStatus(ctx context.Context, key string, alg Algorithm) (*Result, error) {
	redisKey := fmt.Sprintf("rl:lb:%s", key)
	vals, err := l.client.Raw().HMGet(ctx, redisKey, "water", "last_leak").Result()
	if err != nil {
		return nil, err
	}

	water := int64(0)
	if vals[0] != nil {
		fmt.Sscanf(fmt.Sprintf("%v", vals[0]), "%d", &water)
	}

	return &Result{
		Allowed:    true,
		Remaining:  water,
		Algorithm:  AlgorithmLeakyBucket,
		CacheTier:  "L2_REDIS",
	}, nil
}

func (l *LeakyBucketLimiter) Reset(ctx context.Context, key string, alg Algorithm) error {
	redisKey := fmt.Sprintf("rl:lb:%s", key)
	return l.client.Raw().Del(ctx, redisKey).Err()
}

func (l *LeakyBucketLimiter) Name() string {
	return string(AlgorithmLeakyBucket)
}
