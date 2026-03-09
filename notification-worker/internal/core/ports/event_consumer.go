package ports

import "github.com/omni-wallet/notification-worker/internal/core/domain"

type EventConsumer interface {
	Consume(handler func(event domain.TransactionEvent) error) error
	Close() error
}
