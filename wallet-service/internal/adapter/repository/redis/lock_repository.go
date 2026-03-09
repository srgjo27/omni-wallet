package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type LockRepository struct {
	client *goredis.Client
}

func NewLockRepository(client *goredis.Client) *LockRepository {
	return &LockRepository{client: client}
}

func (r *LockRepository) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("acquiring redis lock for key %q: %w", key, err)
	}
	return ok, nil
}

func (r *LockRepository) Release(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("releasing redis lock for key %q: %w", key, err)
	}
	return nil
}
