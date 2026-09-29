package limiter

import (
	"context"
	"math"
	"sync"
	"time"
)

type memoryTokenBucket struct {
	tokens      float64
	lastUpdated time.Time
}

type memoryWindowCounter struct {
	counts    map[int64]int64
	lastClean time.Time
}

type MemoryLimiter struct {
	mu           sync.RWMutex
	tokenBuckets map[string]*memoryTokenBucket
	swCounters   map[string]*memoryWindowCounter
}

func NewMemoryLimiter() *MemoryLimiter {
	m := &MemoryLimiter{
		tokenBuckets: make(map[string]*memoryTokenBucket),
		swCounters:   make(map[string]*memoryWindowCounter),
	}
	// Start background cleanup
	go m.cleanupLoop()
	return m
}

func (m *MemoryLimiter) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for k, b := range m.tokenBuckets {
			if now.Sub(b.lastUpdated) > 5*time.Minute {
				delete(m.tokenBuckets, k)
			}
		}
		for k, w := range m.swCounters {
			if now.Sub(w.lastClean) > 5*time.Minute {
				delete(m.swCounters, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *MemoryLimiter) Allow(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	capacity := float64(req.Capacity)
	if capacity <= 0 {
		capacity = 100
	}
	rate := req.RatePerSecond
	if rate <= 0 {
		rate = 100
	}

	if req.Algorithm == AlgorithmSlidingWindowCounter || req.Algorithm == AlgorithmSlidingWindowLog {
		return m.allowSlidingWindow(req, start)
	}

	// Default to Token Bucket
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, exists := m.tokenBuckets[req.Key]
	if !exists {
		bucket = &memoryTokenBucket{
			tokens:      capacity,
			lastUpdated: start,
		}
		m.tokenBuckets[req.Key] = bucket
	}

	elapsed := start.Sub(bucket.lastUpdated).Seconds()
	bucket.tokens = math.Min(capacity, bucket.tokens+elapsed*rate)
	bucket.lastUpdated = start

	if bucket.tokens >= float64(cost) {
		bucket.tokens -= float64(cost)
		return &Result{
			Allowed:      true,
			Remaining:    int64(bucket.tokens),
			Limit:        int64(capacity),
			RetryAfter:   0,
			ResetAfter:   time.Duration(math.Ceil((capacity-bucket.tokens)/rate)) * time.Second,
			Algorithm:    AlgorithmTokenBucket,
			CacheTier:    "L1_MEMORY",
			LatencyNanos: time.Since(start).Nanoseconds(),
		}, nil
	}

	needed := float64(cost) - bucket.tokens
	retryAfter := time.Duration((needed/rate)*1000) * time.Millisecond

	return &Result{
		Allowed:      false,
		Remaining:    0,
		Limit:        int64(capacity),
		RetryAfter:   retryAfter,
		ResetAfter:   time.Duration(math.Ceil((capacity-bucket.tokens)/rate)) * time.Second,
		Algorithm:    AlgorithmTokenBucket,
		CacheTier:    "L1_MEMORY",
		LatencyNanos: time.Since(start).Nanoseconds(),
	}, nil
}

func (m *MemoryLimiter) allowSlidingWindow(req Request, start time.Time) (*Result, error) {
	windowSize := req.WindowSize
	if windowSize <= 0 {
		windowSize = time.Second
	}
	windowSizeMs := windowSize.Milliseconds()
	limit := req.Capacity
	if limit <= 0 {
		limit = 100
	}

	nowMs := start.UnixMilli()
	currentBucket := nowMs / windowSizeMs
	prevBucket := currentBucket - 1

	m.mu.Lock()
	defer m.mu.Unlock()

	w, exists := m.swCounters[req.Key]
	if !exists {
		w = &memoryWindowCounter{
			counts:    make(map[int64]int64),
			lastClean: start,
		}
		m.swCounters[req.Key] = w
	}

	currCount := w.counts[currentBucket]
	prevCount := w.counts[prevBucket]

	elapsedInWindow := nowMs % windowSizeMs
	weightPrev := float64(windowSizeMs-elapsedInWindow) / float64(windowSizeMs)
	estimatedRequests := int64(float64(prevCount)*weightPrev) + currCount

	cost := req.Cost
	if cost <= 0 {
		cost = 1
	}

	if estimatedRequests+cost <= limit {
		w.counts[currentBucket] = currCount + cost
		return &Result{
			Allowed:      true,
			Remaining:    limit - (estimatedRequests + cost),
			Limit:        limit,
			RetryAfter:   0,
			ResetAfter:   time.Duration(windowSizeMs-elapsedInWindow) * time.Millisecond,
			Algorithm:    AlgorithmSlidingWindowCounter,
			CacheTier:    "L1_MEMORY",
			LatencyNanos: time.Since(start).Nanoseconds(),
		}, nil
	}

	return &Result{
		Allowed:      false,
		Remaining:    0,
		Limit:        limit,
		RetryAfter:   time.Duration(windowSizeMs-elapsedInWindow) * time.Millisecond,
		ResetAfter:   time.Duration(windowSizeMs-elapsedInWindow) * time.Millisecond,
		Algorithm:    AlgorithmSlidingWindowCounter,
		CacheTier:    "L1_MEMORY",
		LatencyNanos: time.Since(start).Nanoseconds(),
	}, nil
}

func (m *MemoryLimiter) GetStatus(ctx context.Context, key string, alg Algorithm) (*Result, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if b, ok := m.tokenBuckets[key]; ok {
		return &Result{
			Allowed:   b.tokens >= 1,
			Remaining: int64(b.tokens),
			Algorithm: AlgorithmTokenBucket,
			CacheTier: "L1_MEMORY",
		}, nil
	}

	return &Result{
		Allowed:   true,
		Remaining: 100,
		Algorithm: alg,
		CacheTier: "L1_MEMORY",
	}, nil
}

func (m *MemoryLimiter) Reset(ctx context.Context, key string, alg Algorithm) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokenBuckets, key)
	delete(m.swCounters, key)
	return nil
}

func (m *MemoryLimiter) Name() string {
	return "in_memory"
}
