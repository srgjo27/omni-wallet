package services

import (
	"fmt"
	"log"

	"github.com/omni-wallet/notification-worker/internal/core/domain"
	"github.com/omni-wallet/notification-worker/internal/core/ports"
)

// NotificationService orchestrates the receipt of transaction events and
// dispatches them to the appropriate notification channels.
type NotificationService struct {
	consumer ports.EventConsumer
	sender   ports.NotificationSender
}

func NewNotificationService(consumer ports.EventConsumer, sender ports.NotificationSender) *NotificationService {
	return &NotificationService{
		consumer: consumer,
		sender:   sender,
	}
}

// Run starts the event consumption loop.
// It blocks until the consumer returns (connection closed or fatal error).
func (s *NotificationService) Run() error {
	log.Println("[notification-service] starting event consumption loop")

	return s.consumer.Consume(func(event domain.TransactionEvent) error {
		log.Printf("[notification-service] received event: type=%s ref=%s user=%s amount=%d",
			event.EventType, event.ReferenceNo, event.UserID, event.Amount)

		if err := s.sender.Send(event); err != nil {
			// Log the error but do NOT return it so the consumer ACKs the
			// message and moves on. A dead-letter queue handles persistent failures.
			log.Printf("[notification-service] WARNING: failed to send notification: %v", err)
			return nil
		}

		return nil
	})
}

// buildMessage creates a human-readable notification message for each event type.
// Exported for testing purposes.
func BuildMessage(event domain.TransactionEvent) string {
	amountFormatted := fmt.Sprintf("Rp%.0f", float64(event.Amount)/100)

	switch event.EventType {
	case domain.EventTopupSuccess:
		return fmt.Sprintf("Top-up berhasil! Saldo kamu bertambah %s. Ref: %s", amountFormatted, event.ReferenceNo)
	case domain.EventTransferSuccess:
		return fmt.Sprintf("Transfer berhasil! %s telah dikirim. Ref: %s", amountFormatted, event.ReferenceNo)
	case domain.EventTransferFailed:
		return fmt.Sprintf("Transfer gagal! Transaksi %s tidak dapat diproses. Ref: %s", amountFormatted, event.ReferenceNo)
	default:
		return fmt.Sprintf("Notifikasi transaksi: %s. Ref: %s", event.EventType, event.ReferenceNo)
	}
}
