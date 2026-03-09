package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/omni-wallet/notification-worker/internal/core/domain"
	"github.com/omni-wallet/notification-worker/internal/platform/messaging"
)

type ConsumerConfig struct {
	ExchangeName  string
	QueueName     string
	RoutingKey    string
	PrefetchCount int
}

type Consumer struct {
	conn    *messaging.RabbitMQConnection
	channel *amqp.Channel
	cfg     ConsumerConfig
}

func NewConsumer(conn *messaging.RabbitMQConnection, cfg ConsumerConfig) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open AMQP channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		cfg.ExchangeName, 
		"topic",         
		true,             
		false,          
		false,            
		false,            
		nil,              
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange %q: %w", cfg.ExchangeName, err)
	}

	if _, err := ch.QueueDeclare(
		cfg.QueueName,
		true,          
		false,         
		false,        
		false,         
		nil,         
	); err != nil {
		return nil, fmt.Errorf("failed to declare queue %q: %w", cfg.QueueName, err)
	}

	if err := ch.QueueBind(
		cfg.QueueName,
		cfg.RoutingKey,
		cfg.ExchangeName,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	if err := ch.Qos(cfg.PrefetchCount, 0, false); err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Printf("[rabbitmq-consumer] ready — exchange=%s queue=%s routing_key=%s",
		cfg.ExchangeName, cfg.QueueName, cfg.RoutingKey)

	return &Consumer{conn: conn, channel: ch, cfg: cfg}, nil
}

func (c *Consumer) Consume(handler func(event domain.TransactionEvent) error) error {
	deliveries, err := c.channel.Consume(
		c.cfg.QueueName,
		"",
		false,
		false,
		false,
		false,
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
			delivery.Nack(false, false)
			continue
		}

		if err := handler(event); err != nil {
			log.Printf("[rabbitmq-consumer] ERROR: handler returned error: %v — requeueing", err)
			delivery.Nack(false, true)
			continue
		}

		delivery.Ack(false)
	}

	return fmt.Errorf("AMQP delivery channel closed — broker disconnected")
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}
