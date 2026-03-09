package domain

import "time"

type EventType string

const (
	EventTopupSuccess EventType = "TOPUP_SUCCESS"
	EventTransferSuccess EventType = "TRANSFER_SUCCESS"
	EventTransferFailed EventType = "TRANSFER_FAILED"
)

type TransactionEvent struct {
	EventType   EventType `json:"event_type"`
	ReferenceNo string    `json:"reference_no"`
	UserID      string    `json:"user_id"`
	TargetUserID string    `json:"target_user_id,omitempty"`
	Amount       int64     `json:"amount"`
	OccurredAt   time.Time `json:"occurred_at"`
}
