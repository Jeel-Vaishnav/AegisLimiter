package limiter

import (
	"context"
	"time"
)

type Algorithm string

const (
	AlgorithmTokenBucket          Algorithm = "token_bucket"
	AlgorithmSlidingWindowCounter Algorithm = "sliding_window_counter"
	AlgorithmSlidingWindowLog     Algorithm = "sliding_window_log"
	AlgorithmLeakyBucket          Algorithm = "leaky_bucket"
)

// Request defines the input parameters for evaluating a rate limit
type Request struct {
	Key           string    `json:"key"`             // Client or route key, e.g. "api:client_123"
	Algorithm     Algorithm `json:"algorithm"`       // Algorithm to use
	Cost          int64     `json:"cost"`            // Token / request cost (default 1)
	Capacity      int64     `json:"capacity"`        // Burst capacity or window limit
	RatePerSecond float64   `json:"rate_per_second"` // Tokens refilled per sec
	WindowSize    time.Duration `json:"window_size"` // Time window for sliding algorithms
}

// Result holds the rate limiting decision and metadata
type Result struct {
	Allowed      bool          `json:"allowed"`
	Remaining    int64         `json:"remaining"`
	Limit        int64         `json:"limit"`
	RetryAfter   time.Duration `json:"retry_after"`
	ResetAfter   time.Duration `json:"reset_after"`
	Algorithm    Algorithm     `json:"algorithm"`
	CacheTier    string        `json:"cache_tier"` // "L1_MEMORY" or "L2_REDIS"
	LatencyNanos int64         `json:"latency_nanos"`
}

// RateLimiter is the common interface implemented by all algorithms and drivers
type RateLimiter interface {
	// Allow checks if the request is permitted and updates quota atomically
	Allow(ctx context.Context, req Request) (*Result, error)

	// GetStatus queries current quota state without consuming quota
	GetStatus(ctx context.Context, key string, alg Algorithm) (*Result, error)

	// Reset clears rate limit state for the given key
	Reset(ctx context.Context, key string, alg Algorithm) error

	// Name returns the identifier of the rate limiter implementation
	Name() string
}

// Rule defines static or dynamic policy per client tier
type Rule struct {
	ClientTier    string        `json:"client_tier"`
	Algorithm     Algorithm     `json:"algorithm"`
	Capacity      int64         `json:"capacity"`
	RatePerSecond float64       `json:"rate_per_second"`
	WindowSize    time.Duration `json:"window_size"`
}
