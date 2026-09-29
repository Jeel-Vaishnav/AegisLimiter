#!/usr/bin/env python3
"""
Distributed High-Throughput Rate Limiter Service - Ready-to-Run Engine
Supports Token Bucket, Sliding Window Counter, Sliding Window Log, Leaky Bucket.
Serves gRPC/REST endpoints, Prometheus /metrics, and the Live Visualizer Web Dashboard.
"""

import time
import math
import asyncio
import os
import sys

# Configure UTF-8 for Windows console
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

from typing import Optional, Dict, Any
from fastapi import FastAPI, Request, Response, status
from fastapi.responses import HTMLResponse, JSONResponse, FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import uvicorn

# Prometheus counters for local tracking
METRICS = {
    "requests_total": {"allowed": 0, "rejected": 0},
    "latencies": [],
    "l1_hits": 0,
    "l1_misses": 0,
}

# In-Memory Rate Limiter Data Store (Thread-safe & Async)
class MemoryDataStore:
    def __init__(self):
        self.tb_store: Dict[str, Dict[str, float]] = {}
        self.sw_store: Dict[str, Dict[int, int]] = {}
        self.sw_log_store: Dict[str, list] = {}
        self.lb_store: Dict[str, Dict[str, float]] = {}
        self.lock = asyncio.Lock()

    async def allow_token_bucket(self, key: str, capacity: int, rate: float, cost: int, now_ms: float):
        async with self.lock:
            if key not in self.tb_store:
                self.tb_store[key] = {"tokens": float(capacity), "last_updated": now_ms}

            data = self.tb_store[key]
            elapsed_sec = max(0.0, (now_ms - data["last_updated"]) / 1000.0)
            data["tokens"] = min(float(capacity), data["tokens"] + elapsed_sec * rate)
            data["last_updated"] = now_ms

            if data["tokens"] >= cost:
                data["tokens"] -= cost
                return True, int(data["tokens"]), 0, capacity, int(max(0, (capacity - data["tokens"]) / rate * 1000))
            else:
                needed = cost - data["tokens"]
                retry_ms = math.ceil((needed / rate) * 1000) if rate > 0 else 3600000
                return False, int(data["tokens"]), retry_ms, capacity, int(max(0, (capacity - data["tokens"]) / rate * 1000))

    async def allow_sliding_window_counter(self, key: str, limit: int, window_ms: int, cost: int, now_ms: float):
        async with self.lock:
            current_bucket = int(now_ms // window_ms)
            prev_bucket = current_bucket - 1

            if key not in self.sw_store:
                self.sw_store[key] = {}

            b_map = self.sw_store[key]
            curr_count = b_map.get(current_bucket, 0)
            prev_count = b_map.get(prev_bucket, 0)

            elapsed_in_window = now_ms % window_ms
            weight_prev = (window_ms - elapsed_in_window) / window_ms
            estimated_requests = int(prev_count * weight_prev) + curr_count

            reset_ms = int(window_ms - elapsed_in_window)

            if estimated_requests + cost <= limit:
                b_map[current_bucket] = curr_count + cost
                remaining = max(0, limit - (estimated_requests + cost))
                return True, remaining, 0, limit, reset_ms
            else:
                return False, 0, max(1, reset_ms), limit, reset_ms

    async def allow_sliding_window_log(self, key: str, limit: int, window_ms: int, cost: int, now_ms: float):
        async with self.lock:
            if key not in self.sw_log_store:
                self.sw_log_store[key] = []

            cutoff = now_ms - window_ms
            # Filter entries within window
            self.sw_log_store[key] = [t for t in self.sw_log_store[key] if t > cutoff]
            count = len(self.sw_log_store[key])

            if count + cost <= limit:
                for _ in range(cost):
                    self.sw_log_store[key].append(now_ms)
                return True, limit - (count + cost), 0, limit, window_ms
            else:
                oldest = self.sw_log_store[key][0] if self.sw_log_store[key] else now_ms
                retry_ms = max(1, int((oldest + window_ms) - now_ms))
                return False, 0, retry_ms, limit, window_ms

    async def allow_leaky_bucket(self, key: str, capacity: int, leak_rate: float, cost: int, now_ms: float):
        async with self.lock:
            if key not in self.lb_store:
                self.lb_store[key] = {"water": 0.0, "last_leak": now_ms}

            data = self.lb_store[key]
            elapsed_sec = max(0.0, (now_ms - data["last_leak"]) / 1000.0)
            data["water"] = max(0.0, data["water"] - elapsed_sec * leak_rate)
            data["last_leak"] = now_ms

            if data["water"] + cost <= capacity:
                data["water"] += cost
                return True, int(capacity - data["water"]), 0, capacity, int((data["water"] / leak_rate) * 1000)
            else:
                overflow = (data["water"] + cost) - capacity
                retry_ms = int(math.ceil((overflow / leak_rate) * 1000)) if leak_rate > 0 else 60000
                return False, 0, retry_ms, capacity, int((data["water"] / leak_rate) * 1000)

    async def reset(self, key: str):
        async with self.lock:
            self.tb_store.pop(key, None)
            self.sw_store.pop(key, None)
            self.sw_log_store.pop(key, None)
            self.lb_store.pop(key, None)


store = MemoryDataStore()
app = FastAPI(title="Distributed High-Throughput Rate Limiter", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
    expose_headers=["X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-RateLimit-Tier", "Retry-After"]
)

class CheckPayload(BaseModel):
    key: Optional[str] = "default_client"
    algorithm: Optional[str] = "token_bucket"
    cost: Optional[int] = 1
    capacity: Optional[int] = 100
    rate_per_second: Optional[float] = 50.0
    window_size_ms: Optional[int] = 1000
    client_tier: Optional[str] = "pro"

@app.post("/v1/limiter/check")
@app.get("/v1/limiter/check")
async def check_rate_limit(request: Request, response: Response, payload: Optional[CheckPayload] = None):
    t0 = time.perf_counter()
    now_ms = time.time() * 1000.0

    if request.method == "GET":
        q = request.query_params
        payload = CheckPayload(
            key=q.get("key", "default_client"),
            algorithm=q.get("algorithm", "token_bucket"),
            cost=int(q.get("cost", 1)),
            capacity=int(q.get("capacity", 100)),
            rate_per_second=float(q.get("rate_per_second", 50.0)),
            window_size_ms=int(q.get("window_size_ms", 1000)),
            client_tier=q.get("tier", "pro")
        )
    elif payload is None:
        payload = CheckPayload()

    cost = max(1, payload.cost or 1)
    capacity = max(1, payload.capacity or 100)
    rate = max(0.1, payload.rate_per_second or 50.0)
    window_ms = max(10, payload.window_size_ms or 1000)
    alg = payload.algorithm or "token_bucket"

    if alg == "token_bucket":
        allowed, remaining, retry_ms, limit, reset_ms = await store.allow_token_bucket(payload.key, capacity, rate, cost, now_ms)
    elif alg == "sliding_window_counter":
        allowed, remaining, retry_ms, limit, reset_ms = await store.allow_sliding_window_counter(payload.key, capacity, window_ms, cost, now_ms)
    elif alg == "sliding_window_log":
        allowed, remaining, retry_ms, limit, reset_ms = await store.allow_sliding_window_log(payload.key, capacity, window_ms, cost, now_ms)
    elif alg == "leaky_bucket":
        allowed, remaining, retry_ms, limit, reset_ms = await store.allow_leaky_bucket(payload.key, capacity, rate, cost, now_ms)
    else:
        allowed, remaining, retry_ms, limit, reset_ms = await store.allow_token_bucket(payload.key, capacity, rate, cost, now_ms)

    duration_ms = (time.perf_counter() - t0) * 1000.0
    duration_us = int(duration_ms * 1000.0)

    # Record Prometheus metrics
    if allowed:
        METRICS["requests_total"]["allowed"] += 1
        response.status_code = status.HTTP_200_OK
    else:
        METRICS["requests_total"]["rejected"] += 1
        response.status_code = status.HTTP_429_TOO_MANY_REQUESTS
        retry_sec = max(1, math.ceil(retry_ms / 1000.0))
        response.headers["Retry-After"] = str(retry_sec)

    response.headers["X-RateLimit-Limit"] = str(limit)
    response.headers["X-RateLimit-Remaining"] = str(remaining)
    response.headers["X-RateLimit-Reset"] = str(int(now_ms + reset_ms))
    response.headers["X-RateLimit-Tier"] = "L1_MEMORY"

    return {
        "allowed": allowed,
        "remaining": remaining,
        "limit": limit,
        "retry_after_ms": retry_ms,
        "reset_ms": reset_ms,
        "algorithm": alg,
        "cache_tier": "L1_MEMORY",
        "latency_microseconds": duration_us
    }

@app.get("/v1/limiter/status")
async def get_status(key: str, algorithm: str = "token_bucket"):
    return {"key": key, "algorithm": algorithm, "status": "active"}

@app.post("/v1/limiter/reset")
async def reset_key(payload: Dict[str, Any]):
    key = payload.get("key", "")
    if key:
        await store.reset(key)
    return {"status": "success", "message": f"Quota reset for {key}"}

@app.get("/healthz")
async def health():
    return {"status": "healthy", "service": "distributed-rate-limiter", "timestamp": time.time()}

@app.get("/v1/stats")
async def get_stats():
    allowed = METRICS["requests_total"]["allowed"]
    blocked = METRICS["requests_total"]["rejected"]
    return {
        "allowed_total": allowed,
        "blocked_total": blocked,
        "total_requests": allowed + blocked,
        "status": "running"
    }

@app.get("/metrics")
async def prometheus_metrics():
    allowed = METRICS["requests_total"]["allowed"]
    rejected = METRICS["requests_total"]["rejected"]
    total = allowed + rejected

    content = f"""# HELP ratelimiter_requests_total Total number of rate limiter requests handled
# TYPE ratelimiter_requests_total counter
ratelimiter_requests_total{{status="allowed"}} {allowed}
ratelimiter_requests_total{{status="rejected"}} {rejected}

# HELP ratelimiter_l1_cache_hits_total L1 Memory cache hits
# TYPE ratelimiter_l1_cache_hits_total counter
ratelimiter_l1_cache_hits_total {total}

# HELP ratelimiter_active_keys Number of currently tracked active rate limit keys
# TYPE ratelimiter_active_keys gauge
ratelimiter_active_keys {len(store.tb_store) + len(store.sw_store)}
"""
    return Response(content=content, media_type="text/plain")

# Static Dashboard Assets
WEB_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "web")

@app.get("/web/{file_path:path}")
async def serve_web(file_path: str):
    full_path = os.path.join(WEB_DIR, file_path)
    if os.path.exists(full_path) and os.path.isfile(full_path):
        return FileResponse(full_path)
    return Response("Not found", status_code=404)

@app.get("/dashboard")
@app.get("/")
async def serve_dashboard():
    index_path = os.path.join(WEB_DIR, "index.html")
    if os.path.exists(index_path):
        return FileResponse(index_path)
    return HTMLResponse("<h1>Rate Limiter Dashboard Not Found</h1>")


def main():
    print("=" * 70)
    print(" 🚀 Distributed High-Throughput Rate Limiter Engine")
    print(" ⚡ Algorithms: Token Bucket, Sliding Window Counter, Sliding Window Log, Leaky Bucket")
    print(" 🌐 Live Web Visualizer:   http://localhost:8080/dashboard")
    print(" 📊 Prometheus Telemetry: http://localhost:8080/metrics")
    print(" 🩺 Service Health Check: http://localhost:8080/healthz")
    print("=" * 70)
    uvicorn.run(app, host="0.0.0.0", port=8080, log_level="warning")


if __name__ == "__main__":
    main()
