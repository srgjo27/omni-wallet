package messaging

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	reconnectMaxAttempts = 10
	reconnectBaseDelay   = 2 * time.Second
)

type RabbitMQConnection struct {
	url  string
	conn *amqp.Connection
}

func NewRabbitMQConnection(url string) (*RabbitMQConnection, error) {
	r := &RabbitMQConnection{url: url}
	if err := r.connect(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RabbitMQConnection) connect() error {
	var lastErr error
	delay := reconnectBaseDelay

	for attempt := 1; attempt <= reconnectMaxAttempts; attempt++ {
		conn, err := amqp.Dial(r.url)
		if err == nil {
			r.conn = conn
			log.Printf("[rabbitmq] connected to broker (attempt %d)", attempt)
			return nil
		}
		lastErr = err
		log.Printf("[rabbitmq] connection attempt %d/%d failed: %v — retrying in %s",
			attempt, reconnectMaxAttempts, err, delay)
		time.Sleep(delay)
		delay = min(delay*2, 30*time.Second)
	}

	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", reconnectMaxAttempts, lastErr)
}

func (r *RabbitMQConnection) Channel() (*amqp.Channel, error) {
	if r.conn == nil || r.conn.IsClosed() {
		if err := r.connect(); err != nil {
			return nil, err
		}
	}
	return r.conn.Channel()
}

func (r *RabbitMQConnection) Close() error {
	if r.conn != nil && !r.conn.IsClosed() {
		return r.conn.Close()
	}
	return nil
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
