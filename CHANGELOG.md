# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-09-29

### Added
- **Multi-Algorithm Defense Engine**:
  - Token Bucket rate limiter with continuous mathematical refill and burst protection.
  - Sliding Window Counter rate limiter with $O(1)$ memory approximation.
  - Sliding Window Log rate limiter with millisecond precision using Redis Sorted Sets (`ZSET`).
  - Leaky Bucket rate limiter for smooth egress traffic shaping.
- **Atomic Redis Lua Architecture**:
  - Embedded Lua scripts executed with SHA preloading (`EVALSHA`) to eliminate race conditions.
  - Dynamic TTL auto-pruning to prevent memory leaks from inactive keys.
- **Dual Ingress Interfaces**:
  - High-performance gRPC service (`ratelimiter.v1.RateLimiterService`) on port `50051`.
  - HTTP REST API with standard rate-limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`) on port `8080`.
- **High-Throughput Tiered Caching**:
  - L1 In-Memory token batch reservations resolving rate decisions in $< 10\,\mu\text{s}$.
  - Benchmarked to sustain $10{,}000+$ requests/second with sub-millisecond p50 latency ($0.06\,\text{ms}$).
- **Interactive Holographic Web Dashboard**:
  - 2D Canvas physics token canister with gravitational bounce and nozzle injection.
  - Network flow pipeline topology visualizer.
  - Oscilloscope throughput & latency charts (60 FPS).
  - Web Audio API synthesizer for acoustic feedback on traffic decisions.
- **Telemetry & Cloud Orchestration**:
  - Prometheus metrics exported at `/metrics`.
  - Pre-provisioned Grafana dashboard configuration.
  - Multi-stage production `Dockerfile` and `docker-compose.yml`.
  - High-concurrency load testing suite with `k6` and Python asyncio runners.
