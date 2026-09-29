package grpc

import (
	"context"
	"fmt"
	"time"

	ratelimiter_pb "github.com/distributed-systems/rate-limiter/pkg/grpc/pb"
	"github.com/distributed-systems/rate-limiter/pkg/limiter"
	"github.com/distributed-systems/rate-limiter/pkg/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ratelimiter_pb.UnimplementedRateLimiterServiceServer
	limiterMap  map[limiter.Algorithm]limiter.RateLimiter
	rules       map[string]*limiter.Rule
	defaultAlg  limiter.Algorithm
	defaultCap  int64
	defaultRate float64
}

func NewServer(limiterMap map[limiter.Algorithm]limiter.RateLimiter, defaultAlg limiter.Algorithm, defaultCap int64, defaultRate float64) *Server {
	return &Server{
		limiterMap:  limiterMap,
		rules:       make(map[string]*limiter.Rule),
		defaultAlg:  defaultAlg,
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
	}
}

func (s *Server) CheckRateLimit(ctx context.Context, req *ratelimiter_pb.RateLimitRequest) (*ratelimiter_pb.RateLimitResponse, error) {
	start := time.Now()

	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key is required")
	}

	alg := mapProtoToAlgorithm(req.Algorithm)
	if alg == "" {
		alg = s.defaultAlg
	}

	lim, ok := s.limiterMap[alg]
	if !ok {
		lim = s.limiterMap[s.defaultAlg]
	}

	capacity := req.Capacity
	if capacity <= 0 {
		capacity = s.defaultCap
	}

	rate := req.RatePerSecond
	if rate <= 0 {
		rate = s.defaultRate
	}

	windowSize := time.Duration(req.WindowSizeMs) * time.Millisecond
	if windowSize <= 0 {
		windowSize = time.Second
	}

	res, err := lim.Allow(ctx, limiter.Request{
		Key:           req.Key,
		Algorithm:     alg,
		Cost:          req.Cost,
		Capacity:      capacity,
		RatePerSecond: rate,
		WindowSize:    windowSize,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rate limiter error: %v", err)
	}

	duration := time.Since(start).Seconds()
	statusStr := "allowed"
	if !res.Allowed {
		statusStr = "rejected"
	}
	metrics.RecordDecision("grpc", string(alg), statusStr, "default", duration, res.CacheTier)

	return &ratelimiter_pb.RateLimitResponse{
		Allowed:      res.Allowed,
		Remaining:    res.Remaining,
		Limit:        res.Limit,
		RetryAfterMs: res.RetryAfter.Milliseconds(),
		ResetMs:      res.ResetAfter.Milliseconds(),
		Algorithm:    req.Algorithm,
		CacheTier:    res.CacheTier,
	}, nil
}

func (s *Server) GetStatus(ctx context.Context, req *ratelimiter_pb.StatusRequest) (*ratelimiter_pb.StatusResponse, error) {
	alg := mapProtoToAlgorithm(req.Algorithm)
	lim, ok := s.limiterMap[alg]
	if !ok {
		lim = s.limiterMap[s.defaultAlg]
	}

	res, err := lim.GetStatus(ctx, req.Key, alg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get status: %v", err)
	}

	return &ratelimiter_pb.StatusResponse{
		Key:       req.Key,
		Algorithm: req.Algorithm,
		Remaining: res.Remaining,
		Limit:     res.Limit,
		ResetMs:   res.ResetAfter.Milliseconds(),
	}, nil
}

func (s *Server) ResetLimit(ctx context.Context, req *ratelimiter_pb.ResetRequest) (*ratelimiter_pb.ResetResponse, error) {
	alg := mapProtoToAlgorithm(req.Algorithm)
	lim, ok := s.limiterMap[alg]
	if !ok {
		lim = s.limiterMap[s.defaultAlg]
	}

	if err := lim.Reset(ctx, req.Key, alg); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset key: %v", err)
	}

	return &ratelimiter_pb.ResetResponse{
		Success: true,
		Message: fmt.Sprintf("Key %s reset successfully", req.Key),
	}, nil
}

func (s *Server) ConfigureRule(ctx context.Context, req *ratelimiter_pb.ConfigureRuleRequest) (*ratelimiter_pb.ConfigureRuleResponse, error) {
	if req.ClientTier == "" {
		return nil, status.Errorf(codes.InvalidArgument, "client_tier is required")
	}

	alg := mapProtoToAlgorithm(req.Algorithm)
	s.rules[req.ClientTier] = &limiter.Rule{
		ClientTier:    req.ClientTier,
		Algorithm:     alg,
		Capacity:      req.Capacity,
		RatePerSecond: req.RatePerSecond,
		WindowSize:    time.Duration(req.WindowSizeMs) * time.Millisecond,
	}

	return &ratelimiter_pb.ConfigureRuleResponse{
		Success: true,
		Message: fmt.Sprintf("Rule configured for tier %s", req.ClientTier),
	}, nil
}

func mapProtoToAlgorithm(p ratelimiter_pb.AlgorithmType) limiter.Algorithm {
	switch p {
	case ratelimiter_pb.AlgorithmType_TOKEN_BUCKET:
		return limiter.AlgorithmTokenBucket
	case ratelimiter_pb.AlgorithmType_SLIDING_WINDOW_COUNTER:
		return limiter.AlgorithmSlidingWindowCounter
	case ratelimiter_pb.AlgorithmType_SLIDING_WINDOW_LOG:
		return limiter.AlgorithmSlidingWindowLog
	case ratelimiter_pb.AlgorithmType_LEAKY_BUCKET:
		return limiter.AlgorithmLeakyBucket
	default:
		return ""
	}
}
