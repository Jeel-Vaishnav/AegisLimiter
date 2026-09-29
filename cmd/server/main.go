package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/distributed-systems/rate-limiter/pkg/config"
	limiterGrpc "github.com/distributed-systems/rate-limiter/pkg/grpc"
	ratelimiter_pb "github.com/distributed-systems/rate-limiter/pkg/grpc/pb"
	limiterHttp "github.com/distributed-systems/rate-limiter/pkg/http"
	"github.com/distributed-systems/rate-limiter/pkg/limiter"
	limiterRedis "github.com/distributed-systems/rate-limiter/pkg/redis"
	"google.golang.org/grpc"
)

func main() {
	log.Println("=================================================================")
	log.Println(" Distributed High-Throughput Rate Limiter Service")
	log.Println(" Protocols: gRPC + REST/HTTP | Engine: Redis Lua + L1 In-Memory")
	log.Println("=================================================================")

	cfg := config.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	limiters := make(map[limiter.Algorithm]limiter.RateLimiter)

	// Attempt to connect to Redis
	rdbClient, err := limiterRedis.NewClient(ctx, cfg.RedisAddress(), cfg.RedisPassword, cfg.RedisDB, cfg.RedisPoolSize)
	if err != nil {
		log.Printf("[WARN] Redis unavailable at %s (%v). Falling back to high-performance In-Memory engine.", cfg.RedisAddress(), err)
		memLimiter := limiter.NewMemoryLimiter()
		limiters[limiter.AlgorithmTokenBucket] = memLimiter
		limiters[limiter.AlgorithmSlidingWindowCounter] = memLimiter
		limiters[limiter.AlgorithmSlidingWindowLog] = memLimiter
		limiters[limiter.AlgorithmLeakyBucket] = memLimiter
	} else {
		log.Printf("[INFO] Connected to Redis cluster/node at %s (Pool size: %d)", cfg.RedisAddress(), cfg.RedisPoolSize)
		tbRedis := limiter.NewTokenBucketLimiter(rdbClient)
		swRedis := limiter.NewSlidingWindowLimiter(rdbClient)
		lbRedis := limiter.NewLeakyBucketLimiter(rdbClient)

		if cfg.EnableL1Cache {
			log.Printf("[INFO] L1 In-Memory Cache Enabled (Batch size: %d tokens, TTL: %v)", cfg.L1BatchSize, cfg.L1CacheTTL)
			limiters[limiter.AlgorithmTokenBucket] = limiter.NewTieredLimiter(tbRedis, cfg.L1BatchSize, cfg.L1CacheTTL)
		} else {
			limiters[limiter.AlgorithmTokenBucket] = tbRedis
		}

		limiters[limiter.AlgorithmSlidingWindowCounter] = swRedis
		limiters[limiter.AlgorithmSlidingWindowLog] = swRedis
		limiters[limiter.AlgorithmLeakyBucket] = lbRedis
	}

	defaultAlg := limiter.Algorithm(cfg.DefaultAlgorithm)
	if _, ok := limiters[defaultAlg]; !ok {
		defaultAlg = limiter.AlgorithmTokenBucket
	}

	// 1. Start gRPC Server
	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("[FATAL] Failed to listen gRPC on port %s: %v", cfg.GRPCPort, err)
	}
	grpcServer := grpc.NewServer()
	grpcSvc := limiterGrpc.NewServer(limiters, defaultAlg, cfg.DefaultCapacity, cfg.DefaultRate)
	ratelimiter_pb.RegisterRateLimiterServiceServer(grpcServer, grpcSvc)

	go func() {
		log.Printf("[INFO] gRPC Server listening on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Printf("[ERROR] gRPC server error: %v", err)
		}
	}()

	// 2. Start HTTP / REST & Metrics Server
	handlers := limiterHttp.NewHandlers(limiters, defaultAlg, cfg.DefaultCapacity, cfg.DefaultRate)
	httpServer := limiterHttp.NewServer(":"+cfg.HTTPPort, handlers)

	go func() {
		log.Printf("[INFO] REST API, Prometheus Metrics & Dashboard listening on http://localhost:%s", cfg.HTTPPort)
		log.Printf("[INFO] Prometheus metrics available at: http://localhost:%s/metrics", cfg.HTTPPort)
		log.Printf("[INFO] Live Visualizer Dashboard at:   http://localhost:%s/dashboard", cfg.HTTPPort)
		if err := httpServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Printf("[ERROR] HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("[INFO] Received signal %v. Initiating graceful shutdown...", sig)

	grpcServer.GracefulStop()
	_ = httpServer.Close()
	if rdbClient != nil {
		_ = rdbClient.Close()
	}
	log.Println("[INFO] High-Throughput Rate Limiter service stopped gracefully.")
}
