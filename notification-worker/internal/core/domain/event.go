package domain

import "time"

type EventType string

const (
	EventTopupSuccess EventType = "TOPUP_SUCCESS"
	EventTransferSuccess EventType = "TRANSFER_SUCCESS"
	EventTransferFailed EventType = "TRANSFER_FAILED"
)

// TransactionEvent is the message schema published by the Wallet Service
// and consumed by this worker to trigger downstream notifications.
type TransactionEvent struct {
	EventType   EventType `json:"event_type"`
	ReferenceNo string    `json:"reference_no"`
	UserID      string    `json:"user_id"`
	// TargetUserID is only populated for P2P transfer events.
	TargetUserID string    `json:"target_user_id,omitempty"`
	Amount       int64     `json:"amount"`
	OccurredAt   time.Time `json:"occurred_at"`
}
