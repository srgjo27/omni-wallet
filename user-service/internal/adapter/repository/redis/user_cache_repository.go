package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "user_session:"

type UserCacheRepository struct {
	client *goredis.Client
}

func NewUserCacheRepository(client *goredis.Client) *UserCacheRepository {
	return &UserCacheRepository{client: client}
}

// SetUserSession stores the JWT token for the given user ID with a TTL.
func (r *UserCacheRepository) SetUserSession(ctx context.Context, userID string, token string, ttlSeconds int64) error {
	key := sessionKeyPrefix + userID
	ttl := time.Duration(ttlSeconds) * time.Second

	if err := r.client.Set(ctx, key, token, ttl).Err(); err != nil {
		return fmt.Errorf("setting user session in redis: %w", err)
	}

	return nil
}

// GetUserSession retrieves the cached JWT token for the given user ID.
func (r *UserCacheRepository) GetUserSession(ctx context.Context, userID string) (string, error) {
	key := sessionKeyPrefix + userID

	token, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return "", nil // session not found — not an error
		}
		return "", fmt.Errorf("getting user session from redis: %w", err)
	}

	return token, nil
}

// DeleteUserSession removes the cached session for the given user ID (logout).
func (r *UserCacheRepository) DeleteUserSession(ctx context.Context, userID string) error {
	key := sessionKeyPrefix + userID

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("deleting user session from redis: %w", err)
	}

	return nil
}
