package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// LockRepository implements ports.DistributedLockRepository using Redis SETNX.
// Each wallet operation acquires a lock before mutating balances, preventing
// race conditions under high concurrency.
type LockRepository struct {
	client *goredis.Client
}

func NewLockRepository(client *goredis.Client) *LockRepository {
	return &LockRepository{client: client}
}

// Acquire tries to set the key only if it does not already exist (NX) with a TTL.
// Returns true if the lock was obtained, false if another process holds it.
func (r *LockRepository) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("acquiring redis lock for key %q: %w", key, err)
	}
	return ok, nil
}

// Release deletes the lock key, making it available for the next caller.
func (r *LockRepository) Release(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("releasing redis lock for key %q: %w", key, err)
	}
	return nil
}
