package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/distributed-systems/rate-limiter/pkg/limiter"
	"github.com/distributed-systems/rate-limiter/pkg/metrics"
)

type CheckRequest struct {
	Key           string  `json:"key"`
	Algorithm     string  `json:"algorithm"`
	Cost          int64   `json:"cost"`
	Capacity      int64   `json:"capacity"`
	RatePerSecond float64 `json:"rate_per_second"`
	WindowSizeMs  int64   `json:"window_size_ms"`
	ClientTier    string  `json:"client_tier"`
}

type CheckResponse struct {
	Allowed      bool   `json:"allowed"`
	Remaining    int64  `json:"remaining"`
	Limit        int64  `json:"limit"`
	RetryAfterMs int64  `json:"retry_after_ms"`
	ResetMs      int64  `json:"reset_ms"`
	Algorithm    string `json:"algorithm"`
	CacheTier    string `json:"cache_tier"`
	LatencyUs    int64  `json:"latency_microseconds"`
}

type Handlers struct {
	mu          sync.RWMutex
	limiterMap  map[limiter.Algorithm]limiter.RateLimiter
	rules       map[string]*limiter.Rule
	defaultAlg  limiter.Algorithm
	defaultCap  int64
	defaultRate float64
	startTime   time.Time
	allowedCount int64
	blockedCount int64
}

func NewHandlers(limiterMap map[limiter.Algorithm]limiter.RateLimiter, defaultAlg limiter.Algorithm, defaultCap int64, defaultRate float64) *Handlers {
	h := &Handlers{
		limiterMap:  limiterMap,
		rules:       make(map[string]*limiter.Rule),
		defaultAlg:  defaultAlg,
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
		startTime:   time.Now(),
	}

	// Default rules for standard tiers
	h.rules["free"] = &limiter.Rule{ClientTier: "free", Algorithm: limiter.AlgorithmTokenBucket, Capacity: 20, RatePerSecond: 10, WindowSize: time.Second}
	h.rules["pro"] = &limiter.Rule{ClientTier: "pro", Algorithm: limiter.AlgorithmTokenBucket, Capacity: 200, RatePerSecond: 100, WindowSize: time.Second}
	h.rules["enterprise"] = &limiter.Rule{ClientTier: "enterprise", Algorithm: limiter.AlgorithmTokenBucket, Capacity: 2000, RatePerSecond: 1000, WindowSize: time.Second}

	return h
}

func (h *Handlers) Check(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req CheckRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}
	} else {
		// Support GET query parameters for quick testing
		q := r.URL.Query()
		req.Key = q.Get("key")
		req.Algorithm = q.Get("algorithm")
		req.ClientTier = q.Get("tier")
		if costStr := q.Get("cost"); costStr != "" {
			req.Cost, _ = strconv.ParseInt(costStr, 10, 64)
		}
		if capStr := q.Get("capacity"); capStr != "" {
			req.Capacity, _ = strconv.ParseInt(capStr, 10, 64)
		}
		if rateStr := q.Get("rate"); rateStr != "" {
			req.RatePerSecond, _ = strconv.ParseFloat(rateStr, 64)
		}
	}

	if req.Key == "" {
		req.Key = r.RemoteAddr
	}
	if req.Cost <= 0 {
		req.Cost = 1
	}

	// Resolve Tier Defaults if specified
	h.mu.RLock()
	if rule, exists := h.rules[req.ClientTier]; exists {
		if req.Capacity <= 0 {
			req.Capacity = rule.Capacity
		}
		if req.RatePerSecond <= 0 {
			req.RatePerSecond = rule.RatePerSecond
		}
		if req.Algorithm == "" {
			req.Algorithm = string(rule.Algorithm)
		}
	}
	h.mu.RUnlock()

	alg := limiter.Algorithm(req.Algorithm)
	if alg == "" {
		alg = h.defaultAlg
	}

	lim, ok := h.limiterMap[alg]
	if !ok {
		lim = h.limiterMap[h.defaultAlg]
		alg = h.defaultAlg
	}

	capacity := req.Capacity
	if capacity <= 0 {
		capacity = h.defaultCap
	}
	rate := req.RatePerSecond
	if rate <= 0 {
		rate = h.defaultRate
	}

	windowSize := time.Duration(req.WindowSizeMs) * time.Millisecond
	if windowSize <= 0 {
		windowSize = time.Second
	}

	res, err := lim.Allow(r.Context(), limiter.Request{
		Key:           req.Key,
		Algorithm:     alg,
		Cost:          req.Cost,
		Capacity:      capacity,
		RatePerSecond: rate,
		WindowSize:    windowSize,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"limiter error: %v"}`, err), http.StatusInternalServerError)
		return
	}

	durationSec := time.Since(start).Seconds()
	durationUs := time.Since(start).Microseconds()
	statusStr := "allowed"
	if !res.Allowed {
		statusStr = "rejected"
	}
	tier := req.ClientTier
	if tier == "" {
		tier = "default"
	}
	metrics.RecordDecision("rest", string(alg), statusStr, tier, durationSec, res.CacheTier)

	// Set standard rate limiting response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(res.Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(res.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(res.ResetAfter).UnixMilli(), 10))
	w.Header().Set("X-RateLimit-Tier", res.CacheTier)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	resp := CheckResponse{
		Allowed:      res.Allowed,
		Remaining:    res.Remaining,
		Limit:        res.Limit,
		RetryAfterMs: res.RetryAfter.Milliseconds(),
		ResetMs:      res.ResetAfter.Milliseconds(),
		Algorithm:    string(alg),
		CacheTier:    res.CacheTier,
		LatencyUs:    durationUs,
	}

	if res.Allowed {
		h.mu.Lock()
		h.allowedCount++
		h.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	} else {
		h.mu.Lock()
		h.blockedCount++
		h.mu.Unlock()
		retrySec := int64(res.RetryAfter.Seconds())
		if retrySec < 1 {
			retrySec = 1
		}
		w.Header().Set("Retry-After", strconv.FormatInt(retrySec, 10))
		w.WriteHeader(http.StatusTooManyRequests)
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, `{"error":"key query parameter required"}`, http.StatusBadRequest)
		return
	}
	alg := limiter.Algorithm(r.URL.Query().Get("algorithm"))
	if alg == "" {
		alg = h.defaultAlg
	}

	lim := h.limiterMap[alg]
	if lim == nil {
		lim = h.limiterMap[h.defaultAlg]
	}

	res, err := lim.GetStatus(r.Context(), key, alg)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *Handlers) Reset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key       string `json:"key"`
		Algorithm string `json:"algorithm"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Key == "" {
		http.Error(w, `{"error":"key required"}`, http.StatusBadRequest)
		return
	}

	for _, lim := range h.limiterMap {
		_ = lim.Reset(r.Context(), body.Key, limiter.Algorithm(body.Algorithm))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(`{"status":"success","message":"rate limit reset"}`))
}

func (h *Handlers) Rules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodPost {
		var rule limiter.Rule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, `{"error":"invalid rule JSON"}`, http.StatusBadRequest)
			return
		}
		if rule.ClientTier == "" {
			http.Error(w, `{"error":"client_tier required"}`, http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		h.rules[rule.ClientTier] = &rule
		h.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(rule)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	_ = json.NewEncoder(w).Encode(h.rules)
}

func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(`{"status":"healthy","uptime_seconds":` + strconv.FormatInt(int64(time.Since(h.startTime).Seconds()), 10) + `}`))
}

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	allowed := h.allowedCount
	blocked := h.blockedCount
	uptime := int64(time.Since(h.startTime).Seconds())
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stats := map[string]interface{}{
		"allowed_total":  allowed,
		"blocked_total":  blocked,
		"total_requests": allowed + blocked,
		"uptime_seconds": uptime,
		"status":         "running",
	}
	_ = json.NewEncoder(w).Encode(stats)
}
