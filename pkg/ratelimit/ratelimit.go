package ratelimit

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RateLimit struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimit(client *redis.Client, limit int, window time.Duration) *RateLimit {
	return &RateLimit{
		client: client,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimit) Allow(ctx context.Context, key string) (bool, error) {
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("rate limit incr: %w", err)
	}

	if count == 1 {
		rl.client.Expire(ctx, key, rl.window)
	}

	return count <= int64(rl.limit), nil
}
