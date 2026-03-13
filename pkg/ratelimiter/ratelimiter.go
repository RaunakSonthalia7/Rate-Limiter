package ratelimiter

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Algorithm represents the rate limiting algorithm
type Algorithm string

const (
	TokenBucket   Algorithm = "token_bucket"
	SlidingWindow Algorithm = "sliding_window"
)

// Config holds the configuration for the rate limiter
type Config struct {
	Algorithm Algorithm
	Capacity  int64  // For token bucket: max tokens, for sliding window: max requests
	Rate      int64  // For token bucket: refill rate per second, for sliding window: window size in seconds
	RedisKey  string // Base key for Redis storage
}

// RateLimiter interface
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// NewRateLimiter creates a new rate limiter based on the config
func NewRateLimiter(config Config, rdb *redis.Client) RateLimiter {
	switch config.Algorithm {
	case TokenBucket:
		return &tokenBucketLimiter{
			config: config,
			rdb:    rdb,
		}
	case SlidingWindow:
		return &slidingWindowLimiter{
			config: config,
			rdb:    rdb,
		}
	default:
		panic("unsupported algorithm")
	}
}
