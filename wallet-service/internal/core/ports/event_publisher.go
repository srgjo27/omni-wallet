package ports

import (
	"context"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.OutboundEvent) error
	Close() error
}
