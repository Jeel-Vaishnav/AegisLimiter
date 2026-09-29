package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb             *redis.Client
	tokenBucketSHA  string
	slidingWindowSHA string
	slidingLogSHA   string
	leakyBucketSHA  string
}

func NewClient(ctx context.Context, addr, password string, db, poolSize int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              db,
		PoolSize:        poolSize,
		MinIdleConns:    poolSize / 4,
		DialTimeout:     2 * time.Second,
		ReadTimeout:     1 * time.Second,
		WriteTimeout:    1 * time.Second,
		MaxRetries:      2,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 64 * time.Millisecond,
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed at %s: %w", addr, err)
	}

	c := &Client{rdb: rdb}

	// Pre-load Lua scripts to compute SHAs
	var err error
	if c.tokenBucketSHA, err = rdb.ScriptLoad(ctx, TokenBucketLua).Result(); err != nil {
		log.Printf("[WARN] Failed to preload TokenBucketLua: %v", err)
	}
	if c.slidingWindowSHA, err = rdb.ScriptLoad(ctx, SlidingWindowCounterLua).Result(); err != nil {
		log.Printf("[WARN] Failed to preload SlidingWindowCounterLua: %v", err)
	}
	if c.slidingLogSHA, err = rdb.ScriptLoad(ctx, SlidingWindowLogLua).Result(); err != nil {
		log.Printf("[WARN] Failed to preload SlidingWindowLogLua: %v", err)
	}
	if c.leakyBucketSHA, err = rdb.ScriptLoad(ctx, LeakyBucketLua).Result(); err != nil {
		log.Printf("[WARN] Failed to preload LeakyBucketLua: %v", err)
	}

	return c, nil
}

func (c *Client) Raw() *redis.Client {
	return c.rdb
}

func (c *Client) EvalTokenBucket(ctx context.Context, key string, capacity int64, refillRate float64, requested int64, nowMs int64) ([]interface{}, error) {
	refillPerMs := refillRate / 1000.0
	keys := []string{key}
	args := []interface{}{capacity, refillPerMs, requested, nowMs}

	if c.tokenBucketSHA != "" {
		res, err := c.rdb.EvalSha(ctx, c.tokenBucketSHA, keys, args...).Result()
		if err == nil {
			return res.([]interface{}), nil
		}
	}
	// Fallback to Eval
	res, err := c.rdb.Eval(ctx, TokenBucketLua, keys, args...).Result()
	if err != nil {
		return nil, err
	}
	return res.([]interface{}), nil
}

func (c *Client) EvalSlidingWindowCounter(ctx context.Context, key string, limit int64, windowSizeMs int64, requested int64, nowMs int64) ([]interface{}, error) {
	keys := []string{key}
	args := []interface{}{limit, windowSizeMs, requested, nowMs}

	if c.slidingWindowSHA != "" {
		res, err := c.rdb.EvalSha(ctx, c.slidingWindowSHA, keys, args...).Result()
		if err == nil {
			return res.([]interface{}), nil
		}
	}
	res, err := c.rdb.Eval(ctx, SlidingWindowCounterLua, keys, args...).Result()
	if err != nil {
		return nil, err
	}
	return res.([]interface{}), nil
}

func (c *Client) EvalSlidingWindowLog(ctx context.Context, key string, limit int64, windowSizeMs int64, requested int64, nowMs int64, nonce string) ([]interface{}, error) {
	keys := []string{key}
	args := []interface{}{limit, windowSizeMs, requested, nowMs, nonce}

	if c.slidingLogSHA != "" {
		res, err := c.rdb.EvalSha(ctx, c.slidingLogSHA, keys, args...).Result()
		if err == nil {
			return res.([]interface{}), nil
		}
	}
	res, err := c.rdb.Eval(ctx, SlidingWindowLogLua, keys, args...).Result()
	if err != nil {
		return nil, err
	}
	return res.([]interface{}), nil
}

func (c *Client) EvalLeakyBucket(ctx context.Context, key string, capacity int64, leakRate float64, requested int64, nowMs int64) ([]interface{}, error) {
	leakPerMs := leakRate / 1000.0
	keys := []string{key}
	args := []interface{}{capacity, leakPerMs, requested, nowMs}

	if c.leakyBucketSHA != "" {
		res, err := c.rdb.EvalSha(ctx, c.leakyBucketSHA, keys, args...).Result()
		if err == nil {
			return res.([]interface{}), nil
		}
	}
	res, err := c.rdb.Eval(ctx, LeakyBucketLua, keys, args...).Result()
	if err != nil {
		return nil, err
	}
	return res.([]interface{}), nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
