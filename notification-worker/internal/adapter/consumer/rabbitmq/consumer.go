package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/omni-wallet/notification-worker/internal/core/domain"
	"github.com/omni-wallet/notification-worker/internal/platform/messaging"
)

// ConsumerConfig holds all parameters needed to declare and bind an AMQP consumer.
type ConsumerConfig struct {
	ExchangeName  string
	QueueName     string
	RoutingKey    string
	PrefetchCount int
}

// Consumer implements ports.EventConsumer using RabbitMQ (AMQP 0.9.1).
type Consumer struct {
	conn    *messaging.RabbitMQConnection
	channel *amqp.Channel
	cfg     ConsumerConfig
}

// NewConsumer creates and initialises a RabbitMQ consumer, declaring the
// exchange and queue if they do not already exist.
func NewConsumer(conn *messaging.RabbitMQConnection, cfg ConsumerConfig) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open AMQP channel: %w", err)
	}

	// Declare a durable topic exchange.
	// Topic routing lets us bind with wildcard patterns like "transaction.#".
	if err := ch.ExchangeDeclare(
		cfg.ExchangeName, // name
		"topic",          // kind
		true,             // durable — survives broker restart
		false,            // auto-delete
		false,            // internal
		false,            // no-wait
		nil,              // args
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange %q: %w", cfg.ExchangeName, err)
	}

	// Declare a durable queue for this worker so messages accumulate when the
	// worker is restarted and are not lost.
	if _, err := ch.QueueDeclare(
		cfg.QueueName, // name
		true,          // durable
		false,         // auto-delete
		false,         // exclusive
		false,         // no-wait
		nil,           // args
	); err != nil {
		return nil, fmt.Errorf("failed to declare queue %q: %w", cfg.QueueName, err)
	}

	// Bind the queue to the exchange with the routing key pattern.
	if err := ch.QueueBind(
		cfg.QueueName,
		cfg.RoutingKey,
		cfg.ExchangeName,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Limit unacknowledged messages in-flight to control worker concurrency.
	if err := ch.Qos(cfg.PrefetchCount, 0, false); err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Printf("[rabbitmq-consumer] ready — exchange=%s queue=%s routing_key=%s",
		cfg.ExchangeName, cfg.QueueName, cfg.RoutingKey)

	return &Consumer{conn: conn, channel: ch, cfg: cfg}, nil
}

// Consume starts the AMQP delivery loop. It blocks until the channel is closed
// or the broker disconnects. Each message is ACKed after successful processing
// and NACKed (with requeue=false) on JSON parse failure.
func (c *Consumer) Consume(handler func(event domain.TransactionEvent) error) error {
	deliveries, err := c.channel.Consume(
		c.cfg.QueueName,
		"",    // consumer tag — broker generates one
		false, // auto-ack disabled — we ACK manually for reliability
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register AMQP consumer: %w", err)
	}

	log.Println("[rabbitmq-consumer] waiting for messages…")

	for delivery := range deliveries {
		var event domain.TransactionEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			log.Printf("[rabbitmq-consumer] ERROR: cannot parse message body: %v — discarding", err)
			// NACK without requeue to prevent a poison-message infinite loop.
			delivery.Nack(false, false)
			continue
		}

		if err := handler(event); err != nil {
			log.Printf("[rabbitmq-consumer] ERROR: handler returned error: %v — requeueing", err)
			// NACK with requeue so the message is retried.
			delivery.Nack(false, true)
			continue
		}

		delivery.Ack(false)
	}

	return fmt.Errorf("AMQP delivery channel closed — broker disconnected")
}

// Close releases the AMQP channel.
func (c *Consumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}
