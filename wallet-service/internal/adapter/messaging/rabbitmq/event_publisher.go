package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

const (
	defaultDialRetries  = 10
	defaultRetryBaseDelay = 2 * time.Second
)

// EventPublisher publishes domain events to a RabbitMQ topic exchange.
// It establishes a single persistent channel and reconnects on failure.
type EventPublisher struct {
	url          string
	exchangeName string
	conn         *amqp.Connection
	channel      *amqp.Channel
	mu           sync.Mutex
}

// NewEventPublisher dials the broker, declares the exchange, and returns a
// ready-to-use EventPublisher.
func NewEventPublisher(url, exchangeName string) (*EventPublisher, error) {
	p := &EventPublisher{
		url:          url,
		exchangeName: exchangeName,
	}

	if err := p.connect(); err != nil {
		return nil, err
	}

	return p, nil
}

// connect dials the AMQP broker and creates a publishing channel with retry.
func (p *EventPublisher) connect() error {
	var lastErr error
	delay := defaultRetryBaseDelay

	for attempt := 1; attempt <= defaultDialRetries; attempt++ {
		conn, err := amqp.Dial(p.url)
		if err == nil {
			ch, err := conn.Channel()
			if err != nil {
				conn.Close()
				lastErr = err
				continue
			}

			// Declare a durable topic exchange.
			if err := ch.ExchangeDeclare(
				p.exchangeName,
				"topic",
				true,  // durable
				false, // auto-delete
				false, // internal
				false, // no-wait
				nil,
			); err != nil {
				ch.Close()
				conn.Close()
				lastErr = err
				continue
			}

			p.conn = conn
			p.channel = ch
			log.Printf("[rabbitmq-publisher] connected to broker (attempt %d)", attempt)
			return nil
		}
		lastErr = err
		log.Printf("[rabbitmq-publisher] connection attempt %d/%d failed: %v — retrying in %s",
			attempt, defaultDialRetries, err, delay)
		time.Sleep(delay)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", defaultDialRetries, lastErr)
}

// routingKey maps an event type to an AMQP topic routing key.
func routingKey(eventType domain.OutboundEventType) string {
	switch eventType {
	case domain.OutboundEventTopupSuccess:
		return "transaction.topup.success"
	case domain.OutboundEventTransferSuccess:
		return "transaction.transfer.success"
	case domain.OutboundEventTransferFailed:
		return "transaction.transfer.failed"
	default:
		return "transaction.unknown"
	}
}

// Publish serialises the event to JSON and publishes it to the exchange.
// Uses a mutex to protect the channel for concurrent callers.
// On channel-level errors it attempts to reconnect before returning.
func (p *EventPublisher) Publish(ctx context.Context, event domain.OutboundEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling event: %w", err)
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // survives broker restart
		Timestamp:    time.Now(),
		Body:         body,
	}

	key := routingKey(event.EventType)

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.channel.PublishWithContext(ctx,
		p.exchangeName, key,
		false, // mandatory
		false, // immediate
		msg,
	); err != nil {
		// Channel is broken — reconnect and retry once.
		log.Printf("[rabbitmq-publisher] publish failed (%v) — reconnecting", err)
		if reconnErr := p.connect(); reconnErr != nil {
			return fmt.Errorf("reconnect failed: %w", reconnErr)
		}
		return p.channel.PublishWithContext(ctx, p.exchangeName, key, false, false, msg)
	}

	return nil
}

func (p *EventPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil && !p.conn.IsClosed() {
		return p.conn.Close()
	}
	return nil
}
