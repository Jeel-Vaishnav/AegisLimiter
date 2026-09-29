package tests

import (
	"context"
	"testing"
	"time"

	"github.com/distributed-systems/rate-limiter/pkg/limiter"
)

func TestTokenBucket_Memory(t *testing.T) {
	mem := limiter.NewMemoryLimiter()
	ctx := context.Background()

	req := limiter.Request{
		Key:           "test_user_1",
		Algorithm:     limiter.AlgorithmTokenBucket,
		Cost:          1,
		Capacity:      5,
		RatePerSecond: 10,
	}

	// First 5 requests should pass
	for i := 0; i < 5; i++ {
		res, err := mem.Allow(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error at request %d: %v", i, err)
		}
		if !res.Allowed {
			t.Fatalf("expected request %d to be allowed", i)
		}
	}

	// 6th request should be rate-limited
	res, err := mem.Allow(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Allowed {
		t.Fatalf("expected 6th request to be blocked (429 Too Many Requests)")
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("expected retry_after to be > 0, got %v", res.RetryAfter)
	}
}

func TestSlidingWindowCounter_Memory(t *testing.T) {
	mem := limiter.NewMemoryLimiter()
	ctx := context.Background()

	req := limiter.Request{
		Key:        "test_sw_user",
		Algorithm:  limiter.AlgorithmSlidingWindowCounter,
		Cost:       1,
		Capacity:   3,
		WindowSize: 500 * time.Millisecond,
	}

	for i := 0; i < 3; i++ {
		res, err := mem.Allow(ctx, req)
		if err != nil || !res.Allowed {
			t.Fatalf("request %d failed: err=%v, allowed=%v", i, err, res.Allowed)
		}
	}

	// 4th request should be blocked
	res, _ := mem.Allow(ctx, req)
	if res.Allowed {
		t.Fatalf("expected 4th request to be rejected")
	}

	// Wait for window to roll over
	time.Sleep(600 * time.Millisecond)

	res, err := mem.Allow(ctx, req)
	if err != nil || !res.Allowed {
		t.Fatalf("expected request after window expiration to be allowed: %v", err)
	}
}

func BenchmarkMemoryTokenBucket(b *testing.B) {
	mem := limiter.NewMemoryLimiter()
	ctx := context.Background()
	req := limiter.Request{
		Key:           "bench_key",
		Algorithm:     limiter.AlgorithmTokenBucket,
		Cost:          1,
		Capacity:      100000000,
		RatePerSecond: 100000000,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = mem.Allow(ctx, req)
		}
	})
}

func BenchmarkMemorySlidingWindow(b *testing.B) {
	mem := limiter.NewMemoryLimiter()
	ctx := context.Background()
	req := limiter.Request{
		Key:        "bench_sw_key",
		Algorithm:  limiter.AlgorithmSlidingWindowCounter,
		Cost:       1,
		Capacity:   100000000,
		WindowSize: time.Minute,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = mem.Allow(ctx, req)
		}
	})
}
