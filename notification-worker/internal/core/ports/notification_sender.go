package ports

import "github.com/omni-wallet/notification-worker/internal/core/domain"

type NotificationSender interface {
	Send(event domain.TransactionEvent) error
}
