package domain

import "time"

type UserEventType string

const (
	UserEventRegistered UserEventType = "USER_REGISTERED"
)

type UserOutboundEvent struct {
	EventType  UserEventType `json:"event_type"`
	UserID     string        `json:"user_id"`
	Email      string        `json:"email"`
	OccurredAt time.Time     `json:"occurred_at"`
}
