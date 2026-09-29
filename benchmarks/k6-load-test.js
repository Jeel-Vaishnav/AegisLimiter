import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom Prometheus-friendly k6 metrics
const allowedRate = new Rate('rate_limiter_allowed_ratio');
const blockedRate = new Rate('rate_limiter_blocked_ratio');
const decisionLatency = new Trend('rate_limiter_decision_latency_ms', true);
const totalRequests = new Counter('rate_limiter_total_evaluations');

export const options = {
  scenarios: {
    // 1. High-Throughput Sustained Load (Targeting 10,000+ req/sec)
    sustained_high_throughput: {
      executor: 'ramping-arrival-rate',
      startRate: 500,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 1500,
      stages: [
        { target: 2000, duration: '10s' },  // Warm-up to 2,000 rps
        { target: 6000, duration: '15s' },  // Ramp to 6,000 rps
        { target: 12000, duration: '20s' }, // Peak at 12,000+ rps (exceeds 10k target)
        { target: 12000, duration: '15s' }, // Hold at 12,000 rps
        { target: 1000, duration: '10s' },  // Cooldown
      ],
    },
    // 2. DDoS Burst Simulation
    spike_traffic: {
      executor: 'constant-arrival-rate',
      rate: 8000,
      timeUnit: '1s',
      duration: '15s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      startTime: '75s',
    },
  },
  thresholds: {
    // 95% of requests must complete with sub-millisecond to low millisecond latency
    http_req_duration: ['p(95)<5', 'p(99)<15'],
    // 0 network or system crash errors (429 is expected rate-limiting, not an error)
    'http_req_failed': ['rate<0.01'], 
  },
};

const BASE_URL = __ENV.RATE_LIMITER_URL || 'http://localhost:8080';

// Pool of 100 realistic client identifiers (80/20 Pareto distribution)
const CLIENT_KEYS = [];
for (let i = 0; i < 100; i++) {
  CLIENT_KEYS.push(`client_api_key_${i.toString().padStart(3, '0')}`);
}

export default function () {
  // Select client key with Pareto distribution (top 10% generate 70% of traffic)
  let keyIndex;
  const rand = Math.random();
  if (rand < 0.70) {
    keyIndex = Math.floor(Math.random() * 10);
  } else {
    keyIndex = 10 + Math.floor(Math.random() * 90);
  }
  const clientKey = CLIENT_KEYS[keyIndex];

  // Alternating algorithms across clients to test mixed workloads
  const algorithm = (keyIndex % 2 === 0) ? 'token_bucket' : 'sliding_window_counter';

  const payload = JSON.stringify({
    key: clientKey,
    algorithm: algorithm,
    cost: 1,
    capacity: 150,
    rate_per_second: 100.0,
    window_size_ms: 1000,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Client-ID': clientKey,
    },
    // Don't mark HTTP 429 as failed HTTP request in k6
    expectedStatuses: [200, 429],
  };

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/v1/limiter/check`, payload, params);
  const latency = Date.now() - startTime;

  totalRequests.add(1);
  decisionLatency.add(latency);

  const isAllowed = res.status === 200;
  const isBlocked = res.status === 429;

  allowedRate.add(isAllowed);
  blockedRate.add(isBlocked);

  check(res, {
    'valid response status': (r) => r.status === 200 || r.status === 429,
    'has rate limit headers': (r) => r.headers['X-Ratelimit-Limit'] !== undefined,
    'sub-millisecond or fast decision': () => latency <= 10,
  });
}
