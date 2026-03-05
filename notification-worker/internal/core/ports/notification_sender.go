package ports

import "github.com/omni-wallet/notification-worker/internal/core/domain"

// NotificationSender defines the contract for sending notifications to end users.
// Different implementations can deliver via push notification, email, SMS, etc.
type NotificationSender interface {
	// Send delivers a notification derived from the given transaction event.
	// Returns an error if the notification cannot be delivered.
	Send(event domain.TransactionEvent) error
}
