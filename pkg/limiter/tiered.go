package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type localBatchToken struct {
	remainingTokens int64
	batchExpiresAt  time.Time
}

// TieredLimiter combines L1 In-Memory token batches with L2 Distributed Redis
type TieredLimiter struct {
	mu           sync.Mutex
	localTokens  map[string]*localBatchToken
	redisLimiter RateLimiter
	batchSize    int64
	cacheTTL     time.Duration
}

func NewTieredLimiter(redisLimiter RateLimiter, batchSize int64, cacheTTL time.Duration) *TieredLimiter {
	if batchSize <= 0 {
		batchSize = 10
	}
	if cacheTTL <= 0 {
		cacheTTL = 500 * time.Millisecond
	}

	return &TieredLimiter{
		localTokens:  make(map[string]*localBatchToken),
		redisLimiter: redisLimiter,
		batchSize:    batchSize,
		cacheTTL:     cacheTTL,
	}
}

func (t *TieredLimiter) Allow(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	now := start

	// If algorithm is not token bucket or batch size is 1, bypass directly to Redis L2
	if req.Algorithm != AlgorithmTokenBucket || t.batchSize <= 1 {
		return t.redisLimiter.Allow(ctx, req)
	}

	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	// 1. Check L1 local cache
	t.mu.Lock()
	entry, exists := t.localTokens[req.Key]
	if exists && entry.remainingTokens >= cost && now.Before(entry.batchExpiresAt) {
		entry.remainingTokens -= cost
		rem := entry.remainingTokens
		t.mu.Unlock()

		return &Result{
			Allowed:      true,
			Remaining:    rem,
			Limit:        req.Capacity,
			RetryAfter:   0,
			ResetAfter:   time.Until(entry.batchExpiresAt),
			Algorithm:    AlgorithmTokenBucket,
			CacheTier:    "L1_MEMORY",
			LatencyNanos: time.Since(start).Nanoseconds(),
		}, nil
	}
	t.mu.Unlock()

	// 2. L1 Miss: Request a batch of tokens from L2 Redis
	batchReq := req
	fetchCount := t.batchSize
	if fetchCount < cost {
		fetchCount = cost
	}
	batchReq.Cost = fetchCount

	redisResult, err := t.redisLimiter.Allow(ctx, batchReq)
	if err != nil {
		return nil, fmt.Errorf("tiered redis error: %w", err)
	}

	if !redisResult.Allowed {
		// Quota exhausted at Redis level: reject request
		redisResult.CacheTier = "L2_REDIS"
		redisResult.LatencyNanos = time.Since(start).Nanoseconds()
		return redisResult, nil
	}

	// 3. Tokens successfully reserved from Redis: store remaining in L1
	leftover := fetchCount - cost
	t.mu.Lock()
	t.localTokens[req.Key] = &localBatchToken{
		remainingTokens: leftover,
		batchExpiresAt:  now.Add(t.cacheTTL),
	}
	t.mu.Unlock()

	return &Result{
		Allowed:      true,
		Remaining:    redisResult.Remaining,
		Limit:        redisResult.Limit,
		RetryAfter:   0,
		ResetAfter:   redisResult.ResetAfter,
		Algorithm:    AlgorithmTokenBucket,
		CacheTier:    "L2_REDIS_RESERVED",
		LatencyNanos: time.Since(start).Nanoseconds(),
	}, nil
}

func (t *TieredLimiter) GetStatus(ctx context.Context, key string, alg Algorithm) (*Result, error) {
	return t.redisLimiter.GetStatus(ctx, key, alg)
}

func (t *TieredLimiter) Reset(ctx context.Context, key string, alg Algorithm) error {
	t.mu.Lock()
	delete(t.localTokens, key)
	t.mu.Unlock()
	return t.redisLimiter.Reset(ctx, key, alg)
}

func (t *TieredLimiter) Name() string {
	return "tiered_l1_l2"
}
