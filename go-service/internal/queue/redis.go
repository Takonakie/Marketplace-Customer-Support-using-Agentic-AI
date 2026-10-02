package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisURL string) (*redis.Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback for simple host:port strings
		opts = &redis.Options{
			Addr: redisURL,
		}
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return rdb, nil
}
