package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	userProvisioningQueue      = "wallet.user.provisioning"
	userRegisteredRoutingKey   = "user.registered"
	consumerDialRetries        = 10
	consumerRetryBaseDelay     = 2 * time.Second
)

type UserRegisteredEvent struct {
	EventType  string    `json:"event_type"`
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	OccurredAt time.Time `json:"occurred_at"`
}

type WalletProvisioner interface {
	ProvisionWalletForUser(ctx context.Context, userID string) error
}

type UserEventConsumer struct {
	url          string
	exchangeName string
	conn         *amqp.Connection
	channel      *amqp.Channel
	provisioner  WalletProvisioner
}

func NewUserEventConsumer(url, exchangeName string, provisioner WalletProvisioner) (*UserEventConsumer, error) {
	c := &UserEventConsumer{
		url:          url,
		exchangeName: exchangeName,
		provisioner:  provisioner,
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *UserEventConsumer) connect() error {
	var lastErr error
	delay := consumerRetryBaseDelay

	for attempt := 1; attempt <= consumerDialRetries; attempt++ {
		conn, err := amqp.Dial(c.url)
		if err == nil {
			ch, err := conn.Channel()
			if err != nil {
				conn.Close()
				lastErr = err
				continue
			}

			if err := ch.ExchangeDeclare(
				c.exchangeName,
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

			if _, err := ch.QueueDeclare(
				userProvisioningQueue,
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

			if err := ch.QueueBind(
				userProvisioningQueue,
				userRegisteredRoutingKey,
				c.exchangeName,
				false,
				nil,
			); err != nil {
				ch.Close()
				conn.Close()
				lastErr = err
				continue
			}

			c.conn = conn
			c.channel = ch
			log.Printf("[wallet-user-consumer] connected to broker (attempt %d)", attempt)
			return nil
		}
		lastErr = err
		log.Printf("[wallet-user-consumer] attempt %d/%d failed: %v — retrying in %s",
			attempt, consumerDialRetries, err, delay)
		time.Sleep(delay)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", consumerDialRetries, lastErr)
}

func (c *UserEventConsumer) Consume() error {
	deliveries, err := c.channel.Consume(
		userProvisioningQueue,
		"",    
		false, 
		false, 
		false, 
		false, 
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("[wallet-user-consumer] waiting for USER_REGISTERED messages…")

	for delivery := range deliveries {
		var evt UserRegisteredEvent
		if err := json.Unmarshal(delivery.Body, &evt); err != nil {
			log.Printf("[wallet-user-consumer] ERROR: cannot parse message: %v — discarding", err)
			delivery.Nack(false, false)
			continue
		}

		if evt.UserID == "" {
			log.Printf("[wallet-user-consumer] ERROR: empty user_id — discarding")
			delivery.Nack(false, false)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := c.provisioner.ProvisionWalletForUser(ctx, evt.UserID)
		cancel()

		if err != nil {
			log.Printf("[wallet-user-consumer] ERROR: provision wallet for user_id=%s: %v — requeueing", evt.UserID, err)
			delivery.Nack(false, true) 
			continue
		}

		log.Printf("[wallet-user-consumer] wallet provisioned for user_id=%s", evt.UserID)
		delivery.Ack(false)
	}

	return fmt.Errorf("AMQP delivery channel closed — broker disconnected")
}

func (c *UserEventConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil && !c.conn.IsClosed() {
		c.conn.Close()
	}
}
