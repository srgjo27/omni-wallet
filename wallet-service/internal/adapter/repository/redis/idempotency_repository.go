package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type IdempotencyRepository struct {
	client *goredis.Client
}

func NewIdempotencyRepository(client *goredis.Client) *IdempotencyRepository {
	return &IdempotencyRepository{client: client}
}

func (r *IdempotencyRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("setting idempotency key %q: %w", key, err)
	}
	return nil
}

func (r *IdempotencyRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("getting idempotency key %q: %w", key, err)
	}
	return val, nil
}
