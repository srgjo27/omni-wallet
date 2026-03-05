package sender

import (
	"log"

	"github.com/omni-wallet/notification-worker/internal/core/domain"
	"github.com/omni-wallet/notification-worker/internal/core/services"
)

// LogNotificationSender is a mock NotificationSender that writes the
// notification to stdout. Replace this with a real push-notification or
// email adapter in production (e.g., FCM, Twilio, SMTP).
type LogNotificationSender struct{}

// NewLogNotificationSender creates a LogNotificationSender.
func NewLogNotificationSender() *LogNotificationSender {
	return &LogNotificationSender{}
}

// Send formats the event into a human-readable message and logs it.
func (s *LogNotificationSender) Send(event domain.TransactionEvent) error {
	message := services.BuildMessage(event)

	log.Printf("[notification] ➤ user_id=%s | %s", event.UserID, message)

	// In production this would:
	//   1. Look up the user's device token / email from User Service.
	//   2. Call FCM / APNS for push notification, or an SMTP relay for email.
	return nil
}
