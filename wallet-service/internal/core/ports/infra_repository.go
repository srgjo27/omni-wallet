package ports

import (
	"context"
	"time"
)

// DistributedLockRepository defines the contract for acquiring and releasing
// distributed locks via Redis. This prevents race conditions when two concurrent
// requests try to modify the same wallet balance simultaneously.
type DistributedLockRepository interface {
	// Acquire tries to acquire a lock for the given key with a TTL.
	// Returns (true, nil) if the lock was obtained, or (false, nil) if already held.
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release releases the lock for the given key.
	Release(ctx context.Context, key string) error
}

// IdempotencyRepository defines the contract for storing and retrieving
// idempotency records in Redis. This ensures that a retried request (same
// reference_no) returns the same response without re-processing the operation.
type IdempotencyRepository interface {
	// Set stores the result payload for the given idempotency key with a TTL.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Get retrieves the stored result for the given idempotency key.
	// Returns ("", nil) if the key does not exist.
	Get(ctx context.Context, key string) (string, error)
}
