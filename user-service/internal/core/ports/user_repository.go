package ports

import (
	"context"

	"github.com/omni-wallet/user-service/internal/core/domain"
)

// UserRepository defines the contract (interface) that any storage implementation
// must satisfy. The service layer depends on this interface, not on any concrete DB driver.
type UserRepository interface {
	// Create persists a new user record and returns the created user.
	Create(ctx context.Context, user *domain.User) (*domain.User, error)

	// FindByID retrieves a user by their unique ID.
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// UpdatePin persists a new transaction PIN hash for the given user.
	UpdatePin(ctx context.Context, userID string, pinHash string) error

	// UpdateKYCStatus updates the KYC status of the given user.
	UpdateKYCStatus(ctx context.Context, userID string, status domain.KYCStatus) error

	// ExistsByEmail checks whether a user with the given email already exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ListUsers returns a paginated list of all users and the total count.
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int, error)
}

// UserCacheRepository defines the contract for caching user-related data (e.g. Redis).
type UserCacheRepository interface {
	// SetUserSession stores a JWT session token for the given user ID with a TTL.
	SetUserSession(ctx context.Context, userID string, token string, ttlSeconds int64) error

	// GetUserSession retrieves the session token for the given user ID.
	GetUserSession(ctx context.Context, userID string) (string, error)

	// DeleteUserSession removes the session token for the given user ID (logout).
	DeleteUserSession(ctx context.Context, userID string) error
}
