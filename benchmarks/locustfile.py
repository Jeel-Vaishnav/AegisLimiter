"""
Locust Load Testing Suite for Distributed Rate Limiter Service
Simulates high-concurrency client fleets with heterogeneous traffic tiers.
"""

import json
import random
from locust import HttpUser, task, between, events

CLIENT_POOLS = {
    "free_tier": [f"free_user_{i}" for i in range(50)],
    "pro_tier": [f"pro_user_{i}" for i in range(20)],
    "enterprise_tier": [f"enterprise_user_{i}" for i in range(5)],
}

ALGORITHMS = ["token_bucket", "sliding_window_counter", "sliding_window_log", "leaky_bucket"]

class RateLimiterUser(HttpUser):
    # Minimal wait time to generate high throughput
    wait_time = between(0.0001, 0.002)

    @task(10)
    def check_token_bucket(self):
        tier = random.choices(["free_tier", "pro_tier", "enterprise_tier"], weights=[70, 25, 5])[0]
        key = random.choice(CLIENT_POOLS[tier])

        payload = {
            "key": key,
            "algorithm": "token_bucket",
            "cost": 1,
            "capacity": 100,
            "rate_per_second": 50.0,
            "tier": tier,
        }

        with self.client.post("/v1/limiter/check", json=payload, catch_response=True) as response:
            if response.status_code in [200, 429]:
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")

    @task(5)
    def check_sliding_window(self):
        key = random.choice(CLIENT_POOLS["pro_tier"])
        payload = {
            "key": key,
            "algorithm": "sliding_window_counter",
            "cost": 1,
            "capacity": 150,
            "window_size_ms": 1000,
        }
        with self.client.post("/v1/limiter/check", json=payload, catch_response=True) as response:
            if response.status_code in [200, 429]:
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")

    @task(1)
    def check_status(self):
        key = random.choice(CLIENT_POOLS["free_tier"])
        self.client.get(f"/v1/limiter/status?key={key}&algorithm=token_bucket")
