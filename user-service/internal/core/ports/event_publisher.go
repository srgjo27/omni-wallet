package ports

import (
	"context"

	"github.com/omni-wallet/user-service/internal/core/domain"
)

type UserEventPublisher interface {
	Publish(ctx context.Context, event domain.UserOutboundEvent) error
	Close() error
}
