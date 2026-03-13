package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type tokenBucketLimiter struct {
	config Config
	rdb    *redis.Client
}

func (t *tokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", t.config.RedisKey, key)

	// Lua script for atomic token bucket
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refillRate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local tokens = redis.call("HGET", key, "tokens")
		local last_refill = redis.call("HGET", key, "last_refill")

		if tokens == false then
			tokens = capacity
			last_refill = now
			redis.call("HSET", key, "tokens", tokens, "last_refill", last_refill)
			redis.call("EXPIRE", key, 86400)  -- expire in 1 day
		else
			tokens = tonumber(tokens)
			last_refill = tonumber(last_refill)
			local time_passed = now - last_refill
			local refill_amount = time_passed / refillRate
			tokens = math.min(capacity, tokens + refill_amount)
			last_refill = now
			redis.call("HSET", key, "tokens", tokens, "last_refill", last_refill)
		end

		if tokens >= 1 then
			tokens = tokens - 1
			redis.call("HSET", key, "tokens", tokens)
			return 1
		else
			return 0
		end
	`

	now := time.Now().Unix()
	result, err := t.rdb.Eval(ctx, script, []string{redisKey}, t.config.Capacity, t.config.Rate, now).Result()
	fmt.Println("result", result)
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}
