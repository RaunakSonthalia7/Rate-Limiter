package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"rate-limiter/pkg/ratelimiter"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	})
	defer rdb.Close()

	// Test connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Create rate limiter config from env
	algorithm := ratelimiter.Algorithm(getEnv("RATE_LIMIT_ALGORITHM", "token_bucket"))
	capacity, _ := strconv.ParseInt(getEnv("RATE_LIMIT_CAPACITY", "10"), 10, 64)
	rate, _ := strconv.ParseInt(getEnv("RATE_LIMIT_RATE", "6"), 10, 64)
	redisKey := getEnv("RATE_LIMIT_REDIS_KEY", "ratelimit")

	config := ratelimiter.Config{
		Algorithm: algorithm,
		Capacity:  capacity,
		Rate:      rate,
		RedisKey:  redisKey,
	}
	limiter := ratelimiter.NewRateLimiter(config, rdb)

	// HTTP server with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/api", rateLimitMiddleware(limiter, apiHandler))

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on :%s with algorithm: %s", port, algorithm)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func rateLimitMiddleware(limiter ratelimiter.RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use client IP as key (without port)
		key := strings.Split(r.RemoteAddr, ":")[0]

		allowed, err := limiter.Allow(r.Context(), key)
		log.Printf("Rate limit check for key %s: allowed=%v, err=%v", key, allowed, err)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			log.Printf("Rate limit exceeded for key %s", key)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API response at %s", time.Now().Format(time.RFC3339))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
