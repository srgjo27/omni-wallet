package domain

import "time"

type OutboundEventType string

const (
	OutboundEventTopupSuccess    OutboundEventType = "TOPUP_SUCCESS"
	OutboundEventTransferSuccess OutboundEventType = "TRANSFER_SUCCESS"
	OutboundEventTransferFailed  OutboundEventType = "TRANSFER_FAILED"
)

type OutboundEvent struct {
	EventType    OutboundEventType `json:"event_type"`
	ReferenceNo  string            `json:"reference_no"`
	UserID       string            `json:"user_id"`
	TargetUserID string            `json:"target_user_id,omitempty"`
	Amount       int64             `json:"amount"`
	OccurredAt   time.Time         `json:"occurred_at"`
}
