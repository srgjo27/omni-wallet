package ports

import "github.com/omni-wallet/notification-worker/internal/core/domain"

// EventConsumer defines the contract for consuming events from a message broker.
// Implementations must call handler for each received event and block until
// the context is cancelled or a fatal error occurs.
type EventConsumer interface {
	// Consume starts consuming events from the broker.
	// The provided handler is called for every received event.
	// Returns an error if the consumer cannot start or encounters a fatal failure.
	Consume(handler func(event domain.TransactionEvent) error) error

	// Close gracefully shuts down the consumer connection.
	Close() error
}
