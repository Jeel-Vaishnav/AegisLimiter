package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Sub-millisecond fine-grained latency buckets: 50µs to 50ms
	latencyBuckets = []float64{
		0.00005, 0.0001, 0.00025, 0.0005, 0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1,
	}

	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiter",
			Name:      "requests_total",
			Help:      "Total number of rate limiter requests handled",
		},
		[]string{"algorithm", "status", "tier"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ratelimiter",
			Name:      "request_duration_seconds",
			Help:      "Latency of rate limit decisions in seconds",
			Buckets:   latencyBuckets,
		},
		[]string{"handler", "algorithm"},
	)

	RedisDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ratelimiter",
			Name:      "redis_latency_seconds",
			Help:      "Latency of Redis Lua script executions in seconds",
			Buckets:   latencyBuckets,
		},
		[]string{"algorithm"},
	)

	L1CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ratelimiter",
			Name:      "l1_cache_hits_total",
			Help:      "Total count of requests satisfied directly by L1 local memory cache",
		},
	)

	L1CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ratelimiter",
			Name:      "l1_cache_misses_total",
			Help:      "Total count of L1 cache misses requiring L2 Redis synchronization",
		},
	)

	ActiveKeys = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ratelimiter",
			Name:      "active_keys",
			Help:      "Number of currently tracked active rate limit keys",
		},
	)
)

func RecordDecision(handler, algorithm, status, tier string, durationSeconds float64, cacheTier string) {
	RequestsTotal.WithLabelValues(algorithm, status, tier).Inc()
	RequestDuration.WithLabelValues(handler, algorithm).Observe(durationSeconds)

	if cacheTier == "L1_MEMORY" {
		L1CacheHits.Inc()
	} else if cacheTier == "L2_REDIS_RESERVED" {
		L1CacheMisses.Inc()
	}
}
