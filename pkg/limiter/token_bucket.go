package limiter

import (
	"context"
	"fmt"
	"time"

	limiterRedis "github.com/distributed-systems/rate-limiter/pkg/redis"
)

type TokenBucketLimiter struct {
	client *limiterRedis.Client
}

func NewTokenBucketLimiter(client *limiterRedis.Client) *TokenBucketLimiter {
	return &TokenBucketLimiter{client: client}
}

func (t *TokenBucketLimiter) Allow(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	redisKey := fmt.Sprintf("rl:tb:%s", req.Key)
	nowMs := start.UnixMilli()

	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	res, err := t.client.EvalTokenBucket(ctx, redisKey, req.Capacity, req.RatePerSecond, cost, nowMs)
	if err != nil {
		return nil, fmt.Errorf("token bucket eval error: %w", err)
	}

	if len(res) < 5 {
		return nil, fmt.Errorf("unexpected token bucket response length: %d", len(res))
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
		Algorithm:    AlgorithmTokenBucket,
		CacheTier:    "L2_REDIS",
		LatencyNanos: time.Since(start).Nanoseconds(),
	}, nil
}

func (t *TokenBucketLimiter) GetStatus(ctx context.Context, key string, alg Algorithm) (*Result, error) {
	redisKey := fmt.Sprintf("rl:tb:%s", key)
	vals, err := t.client.Raw().HMGet(ctx, redisKey, "tokens", "last_updated").Result()
	if err != nil {
		return nil, err
	}

	tokens := int64(0)
	if vals[0] != nil {
		fmt.Sscanf(fmt.Sprintf("%v", vals[0]), "%d", &tokens)
	}

	return &Result{
		Allowed:    tokens > 0,
		Remaining:  tokens,
		Algorithm:  AlgorithmTokenBucket,
		CacheTier:  "L2_REDIS",
	}, nil
}

func (t *TokenBucketLimiter) Reset(ctx context.Context, key string, alg Algorithm) error {
	redisKey := fmt.Sprintf("rl:tb:%s", key)
	return t.client.Raw().Del(ctx, redisKey).Err()
}

func (t *TokenBucketLimiter) Name() string {
	return string(AlgorithmTokenBucket)
}
