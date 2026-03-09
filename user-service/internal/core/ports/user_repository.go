package ports

import (
	"context"

	"github.com/omni-wallet/user-service/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdatePin(ctx context.Context, userID string, pinHash string) error
	UpdateKYCStatus(ctx context.Context, userID string, status domain.KYCStatus) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int, error)
	GetStats(ctx context.Context) (totalUsers, verifiedUsers int, err error)
}

type UserCacheRepository interface {
	SetUserSession(ctx context.Context, userID string, token string, ttlSeconds int64) error
	GetUserSession(ctx context.Context, userID string) (string, error)
	DeleteUserSession(ctx context.Context, userID string) error
}
