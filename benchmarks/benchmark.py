#!/usr/bin/env python3
"""
High-Throughput Benchmark Harness for Distributed Rate Limiter Service
Uses asyncio + aiohttp/httpx to generate concurrent load and report percentiles.
"""

import asyncio
import time
import sys
import statistics
import argparse

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

try:
    import httpx
except ImportError:
    httpx = None

try:
    import aiohttp
except ImportError:
    aiohttp = None


async def run_benchmark(target_url: str, total_requests: int, concurrency: int, algorithm: str, capacity: int, rate: float):
    print("=" * 70)
    print(" 🚀 Distributed Rate Limiter - High Throughput Benchmark")
    print("=" * 70)
    print(f" Target URL:        {target_url}")
    print(f" Total Requests:    {total_requests:,}")
    print(f" Concurrency Level: {concurrency} workers")
    print(f" Algorithm:         {algorithm}")
    print(f" Capacity / Limit:  {capacity} tokens")
    print(f" Refill Rate:       {rate} tokens/sec")
    print("-" * 70)

    latencies_ms = []
    allowed_count = 0
    blocked_count = 0
    error_count = 0

    endpoint = f"{target_url.rstrip('/')}/v1/limiter/check"
    req_per_worker = total_requests // concurrency

    async def worker(worker_id: int):
        nonlocal allowed_count, blocked_count, error_count
        headers = {"Content-Type": "application/json"}
        
        # Test client pool: 10 distinct clients
        client_key = f"bench_client_{worker_id % 10}"

        payload = {
            "key": client_key,
            "algorithm": algorithm,
            "cost": 1,
            "capacity": capacity,
            "rate_per_second": rate,
            "window_size_ms": 1000,
        }

        if aiohttp:
            timeout = aiohttp.ClientTimeout(total=5.0)
            async with aiohttp.ClientSession(timeout=timeout) as session:
                for _ in range(req_per_worker):
                    t0 = time.perf_counter()
                    try:
                        async with session.post(endpoint, json=payload, headers=headers) as resp:
                            await resp.read()
                            t1 = time.perf_counter()
                            latencies_ms.append((t1 - t0) * 1000.0)
                            if resp.status == 200:
                                allowed_count += 1
                            elif resp.status == 429:
                                blocked_count += 1
                            else:
                                error_count += 1
                    except Exception as e:
                        error_count += 1
        elif httpx:
            limits = httpx.Limits(max_keepalive_connections=concurrency, max_connections=concurrency)
            async with httpx.AsyncClient(limits=limits, timeout=5.0) as client:
                for _ in range(req_per_worker):
                    t0 = time.perf_counter()
                    try:
                        resp = await client.post(endpoint, json=payload, headers=headers)
                        t1 = time.perf_counter()
                        latencies_ms.append((t1 - t0) * 1000.0)
                        if resp.status_code == 200:
                            allowed_count += 1
                        elif resp.status_code == 429:
                            blocked_count += 1
                        else:
                            error_count += 1
                    except Exception:
                        error_count += 1

    start_time = time.perf_counter()
    tasks = [asyncio.create_task(worker(i)) for i in range(concurrency)]
    await asyncio.gather(*tasks)
    total_time = time.perf_counter() - start_time

    completed = allowed_count + blocked_count + error_count
    rps = completed / total_time if total_time > 0 else 0

    latencies_ms.sort()
    p50 = statistics.median(latencies_ms) if latencies_ms else 0
    p90 = latencies_ms[int(len(latencies_ms) * 0.90)] if latencies_ms else 0
    p95 = latencies_ms[int(len(latencies_ms) * 0.95)] if latencies_ms else 0
    p99 = latencies_ms[int(len(latencies_ms) * 0.99)] if latencies_ms else 0
    mean_lat = statistics.mean(latencies_ms) if latencies_ms else 0
    min_lat = min(latencies_ms) if latencies_ms else 0
    max_lat = max(latencies_ms) if latencies_ms else 0

    print("\n" + "=" * 70)
    print(" 📊 BENCHMARK RESULTS")
    print("=" * 70)
    print(f" Total Time:          {total_time:.3f} seconds")
    print(f" Throughput:          {rps:,.1f} requests/sec")
    print(f" Allowed (200 OK):    {allowed_count:,} ({allowed_count/completed*100:.1f}%)" if completed else "N/A")
    print(f" Blocked (429 Rate):  {blocked_count:,} ({blocked_count/completed*100:.1f}%)" if completed else "N/A")
    print(f" Errors / Timeouts:   {error_count:,}")
    print("-" * 70)
    print(" ⚡ LATENCY PERCENTILES")
    print(f"   Min:               {min_lat:.3f} ms")
    print(f"   p50 (Median):      {p50:.3f} ms")
    print(f"   Mean:              {mean_lat:.3f} ms")
    print(f"   p90:               {p90:.3f} ms")
    print(f"   p95:               {p95:.3f} ms")
    print(f"   p99:               {p99:.3f} ms")
    print(f"   Max:               {max_lat:.3f} ms")
    print("=" * 70)


def main():
    parser = argparse.ArgumentParser(description="Rate Limiter Benchmark")
    parser.add_argument("--url", default="http://localhost:8080", help="Base URL of rate limiter service")
    parser.add_argument("-n", "--requests", type=int, default=10000, help="Total requests to send")
    parser.add_argument("-c", "--concurrency", type=int, default=50, help="Concurrent workers")
    parser.add_argument("-a", "--algorithm", default="token_bucket", choices=["token_bucket", "sliding_window_counter", "sliding_window_log", "leaky_bucket"])
    parser.add_argument("--capacity", type=int, default=200, help="Capacity / Limit")
    parser.add_argument("--rate", type=float, default=100.0, help="Refill rate per sec")
    args = parser.parse_args()

    asyncio.run(run_benchmark(args.url, args.requests, args.concurrency, args.algorithm, args.capacity, args.rate))


if __name__ == "__main__":
    main()
