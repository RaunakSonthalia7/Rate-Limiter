# Rate Limiter

A production-ready rate limiter in Go supporting Token Bucket and Sliding Window algorithms with Redis storage.

## Features

- Two algorithms: Token Bucket and Sliding Window
- Redis-backed for distributed consistency
- Atomic operations using Lua scripts
- Configurable via environment variables
- HTTP middleware for easy integration

## Algorithms

### Token Bucket
- Allows bursts up to capacity
- Refills at a constant rate
- Smooth rate limiting

### Sliding Window
- Counts requests in a moving time window
- More precise for burst control
- Removes old requests automatically

## Configuration

Set environment variables:

- `RATE_LIMIT_ALGORITHM`: `token_bucket` or `sliding_window` (default: `token_bucket`)
- `RATE_LIMIT_CAPACITY`: Max tokens/requests (default: 10)
- `RATE_LIMIT_RATE`: For token bucket: seconds per token (default: 6), for sliding window: window size in seconds (default: 60)
- `RATE_LIMIT_REDIS_KEY`: Base Redis key (default: `ratelimit`)
- `REDIS_ADDR`: Redis address (default: `localhost:6379`)
- `PORT`: Server port (default: `8080`)

## Usage

1. Start Redis server
2. Set environment variables as needed
3. Run the server: `go run cmd/server/main.go`
4. Make requests to `http://localhost:8080/api`

The middleware uses the client IP address as the rate limit key.

## Docker

### Option 1: Using Docker Compose (Recommended)
For a complete setup with Redis:

```bash
docker-compose up
```

This will start both Redis and the rate limiter service.

### Option 2: Run Redis Separately
Start Redis in a container:

```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

Then build and run the rate limiter:

```bash
docker build -t rate-limiter .
docker run -p 8080:8080 -e REDIS_ADDR=host.docker.internal:6379 rate-limiter
```

### Option 3: Standalone Build
Build the image:

```bash
docker build -t rate-limiter .
```

Run the container (requires Redis running separately):

```bash
docker run -p 8080:8080 -e REDIS_ADDR=host.docker.internal:6379 rate-limiter
```

## Consistency

Operations are atomic using Redis Lua scripts, ensuring consistency in distributed environments.