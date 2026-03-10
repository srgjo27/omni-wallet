package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/omni-wallet/user-service/internal/core/domain"
)

const (
	defaultDialRetries    = 10
	defaultRetryBaseDelay = 2 * time.Second
)

type UserEventPublisher struct {
	url          string
	exchangeName string
	conn         *amqp.Connection
	channel      *amqp.Channel
	mu           sync.Mutex
}

func NewUserEventPublisher(url, exchangeName string) (*UserEventPublisher, error) {
	p := &UserEventPublisher{
		url:          url,
		exchangeName: exchangeName,
	}
	if err := p.connect(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *UserEventPublisher) connect() error {
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

			if err := ch.ExchangeDeclare(
				p.exchangeName,
				"topic",
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				ch.Close()
				conn.Close()
				lastErr = err
				continue
			}

			p.conn = conn
			p.channel = ch
			log.Printf("[user-rabbitmq-publisher] connected to broker (attempt %d)", attempt)
			return nil
		}
		lastErr = err
		log.Printf("[user-rabbitmq-publisher] attempt %d/%d failed: %v — retrying in %s",
			attempt, defaultDialRetries, err, delay)
		time.Sleep(delay)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", defaultDialRetries, lastErr)
}

func routingKey(eventType domain.UserEventType) string {
	switch eventType {
	case domain.UserEventRegistered:
		return "user.registered"
	default:
		return "user.unknown"
	}
}

func (p *UserEventPublisher) Publish(ctx context.Context, event domain.UserOutboundEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling event: %w", err)
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         body,
	}

	key := routingKey(event.EventType)

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.channel.PublishWithContext(ctx, p.exchangeName, key, false, false, msg); err != nil {
		log.Printf("[user-rabbitmq-publisher] publish failed (%v) — reconnecting", err)
		if reconnErr := p.connect(); reconnErr != nil {
			return fmt.Errorf("reconnect failed: %w", reconnErr)
		}
		return p.channel.PublishWithContext(ctx, p.exchangeName, key, false, false, msg)
	}

	return nil
}

func (p *UserEventPublisher) Close() error {
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
