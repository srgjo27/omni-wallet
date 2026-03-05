package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// IdempotencyRepository implements ports.IdempotencyRepository using Redis.
// It caches the serialised result of processed transactions so that retried
// requests with the same reference_no return the cached result immediately
// without re-executing the business logic.
type IdempotencyRepository struct {
	client *goredis.Client
}

func NewIdempotencyRepository(client *goredis.Client) *IdempotencyRepository {
	return &IdempotencyRepository{client: client}
}

// Set stores the value string under key with the given TTL.
func (r *IdempotencyRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("setting idempotency key %q: %w", key, err)
	}
	return nil
}

// Get retrieves the value for key. Returns ("", nil) when the key does not exist.
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
