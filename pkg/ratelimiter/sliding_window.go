package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type slidingWindowLimiter struct {
	config Config
	rdb    *redis.Client
}

func (s *slidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", s.config.RedisKey, key)

	// Lua script for atomic sliding window
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local min_score = now - window

		-- Remove old entries
		redis.call("ZREMRANGEBYSCORE", key, "-inf", min_score)

		-- Count current entries
		local count = redis.call("ZCARD", key)

		if count < capacity then
			redis.call("ZADD", key, now, now)
			-- Set expiration to clean up
			redis.call("EXPIRE", key, window)
			return 1
		else
			return 0
		end
	`

	now := time.Now().Unix()
	result, err := s.rdb.Eval(ctx, script, []string{redisKey}, s.config.Capacity, s.config.Rate, now).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}
