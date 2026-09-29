#!/usr/bin/env python3
"""
Traffic Scenario Simulator for Distributed Rate Limiter Service
Generates realistic client traffic patterns:
  1. Steady state normal traffic
  2. Sudden burst / flash crowd
  3. Distributed DDoS surge
"""

import time
import random
import requests
import sys

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

BASE_URL = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8080"
ENDPOINT = f"{BASE_URL}/v1/limiter/check"

CLIENT_PROFILES = [
    {"name": "Frontend Web App", "key": "app_frontend", "tier": "pro", "rate": 80, "burst": 150},
    {"name": "Mobile iOS Client", "key": "app_ios", "tier": "free", "rate": 15, "burst": 30},
    {"name": "Mobile Android Client", "key": "app_android", "tier": "free", "rate": 15, "burst": 30},
    {"name": "Payment Service Backend", "key": "svc_payments", "tier": "enterprise", "rate": 500, "burst": 1000},
    {"name": "Analytics Ingest Agent", "key": "svc_analytics", "tier": "pro", "rate": 100, "burst": 200},
]

def send_check(key: str, algorithm: str = "token_bucket", capacity: int = 100, rate: float = 50.0):
    payload = {
        "key": key,
        "algorithm": algorithm,
        "cost": 1,
        "capacity": capacity,
        "rate_per_second": rate,
    }
    try:
        r = requests.post(ENDPOINT, json=payload, timeout=2.0)
        return r.status_code == 200, r.headers.get("X-RateLimit-Remaining", "0")
    except Exception as e:
        return False, "ERR"

def main():
    print("=" * 60)
    print(" 🚗 Rate Limiter Real-Time Traffic Generator")
    print(f" Target Service: {BASE_URL}")
    print(" Press Ctrl+C to terminate simulation")
    print("=" * 60)

    step = 0
    while True:
        step += 1
        # Random burst every 15 steps
        is_surge = (step % 20) in [15, 16, 17]
        multiplier = 6 if is_surge else 1

        if is_surge:
            print(f" 🔥 [STEP {step}] SIMULATING TRAFFIC SPIKE / BURST!")
        else:
            print(f" 🟢 [STEP {step}] Normal Steady Traffic Flow")

        for profile in CLIENT_PROFILES:
            count = random.randint(1, 4) * multiplier
            for _ in range(count):
                alg = "token_bucket" if step % 2 == 0 else "sliding_window_counter"
                allowed, remaining = send_check(profile["key"], alg, profile["burst"], profile["rate"])
                symbol = "✅ ALLOWED" if allowed else "⛔ RATE-LIMITED (429)"
                print(f"   [{profile['name']}] -> {symbol} | Rem: {remaining}")

        time.sleep(0.5)

if __name__ == "__main__":
    main()
