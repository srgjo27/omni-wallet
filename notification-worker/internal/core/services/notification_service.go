package services

import (
	"fmt"
	"log"

	"github.com/omni-wallet/notification-worker/internal/core/domain"
	"github.com/omni-wallet/notification-worker/internal/core/ports"
)

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

func (s *NotificationService) Run() error {
	log.Println("[notification-service] starting event consumption loop")

	return s.consumer.Consume(func(event domain.TransactionEvent) error {
		log.Printf("[notification-service] received event: type=%s ref=%s user=%s amount=%d",
			event.EventType, event.ReferenceNo, event.UserID, event.Amount)

		if err := s.sender.Send(event); err != nil {
			log.Printf("[notification-service] WARNING: failed to send notification: %v", err)
			return nil
		}

		return nil
	})
}

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
