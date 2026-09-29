package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server ports
	HTTPPort string
	GRPCPort string

	// Redis connection
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisPoolSize int

	// Rate Limiting Defaults
	DefaultAlgorithm string // "token_bucket", "sliding_window_counter", "sliding_window_log", "leaky_bucket"
	DefaultCapacity  int64  // default burst capacity (e.g. 100 requests)
	DefaultRate      float64 // default refill rate (requests / second, e.g. 50.0)
	DefaultWindowMs  int64  // default window in ms (e.g. 1000 for 1 second)

	// L1 Cache (Tiered high throughput)
	EnableL1Cache   bool
	L1BatchSize     int64         // number of tokens pre-allocated per reservation
	L1CacheTTL      time.Duration // in-memory TTL before re-syncing with Redis
}

func LoadConfig() *Config {
	return &Config{
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		GRPCPort:         getEnv("GRPC_PORT", "50051"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvInt("REDIS_DB", 0),
		RedisPoolSize:    getEnvInt("REDIS_POOL_SIZE", 256),
		DefaultAlgorithm: getEnv("DEFAULT_ALGORITHM", "token_bucket"),
		DefaultCapacity:  getEnvInt64("DEFAULT_CAPACITY", 100),
		DefaultRate:      getEnvFloat64("DEFAULT_RATE", 100.0),
		DefaultWindowMs:  getEnvInt64("DEFAULT_WINDOW_MS", 1000),
		EnableL1Cache:    getEnvBool("ENABLE_L1_CACHE", true),
		L1BatchSize:      getEnvInt64("L1_BATCH_SIZE", 10),
		L1CacheTTL:       time.Duration(getEnvInt("L1_CACHE_TTL_MS", 500)) * time.Millisecond,
	}
}

func (c *Config) RedisAddress() string {
	return c.RedisHost + ":" + c.RedisPort
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvFloat64(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}
