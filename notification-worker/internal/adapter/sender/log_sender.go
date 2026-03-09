package sender

import (
	"log"

	"github.com/omni-wallet/notification-worker/internal/core/domain"
	"github.com/omni-wallet/notification-worker/internal/core/services"
)

type LogNotificationSender struct{}

func NewLogNotificationSender() *LogNotificationSender {
	return &LogNotificationSender{}
}

func (s *LogNotificationSender) Send(event domain.TransactionEvent) error {
	message := services.BuildMessage(event)

	log.Printf("[notification] ➤ user_id=%s | %s", event.UserID, message)
	return nil
}
