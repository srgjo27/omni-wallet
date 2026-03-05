package ports

import (
	"context"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

// EventPublisher defines the contract for publishing domain events to an
// external message broker. Implementations must be safe for concurrent use.
type EventPublisher interface {
	// Publish sends a single domain event to the message broker.
	// Returns nil on success. A publish failure should be treated as non-fatal
	// by the caller (the DB is already consistent at this point).
	Publish(ctx context.Context, event domain.OutboundEvent) error

	// Close releases any open connections to the broker.
	Close() error
}
