package config

import (
	"fmt"
	"os"
)

// AppConfig holds all configuration loaded from environment variables.
type AppConfig struct {
	App      AppSettings
	RabbitMQ RabbitMQConfig
}

// AppSettings contains general application settings.
type AppSettings struct {
	Name string
	Env  string
}

// RabbitMQConfig contains AMQP connection settings.
type RabbitMQConfig struct {
	URL          string // amqp://user:pass@host:5672/vhost
	ExchangeName string
	QueueName    string
	RoutingKey   string
	// PrefetchCount controls how many unacknowledged messages the broker
	// delivers to this worker at a time (worker concurrency throttle).
	PrefetchCount int
}

// Load reads all required configuration from environment variables.
// Panics immediately on missing critical configuration.
func Load() *AppConfig {
	prefetchCount := 10
	if v := os.Getenv("RABBITMQ_PREFETCH_COUNT"); v != "" {
		fmt.Sscan(v, &prefetchCount)
	}

	return &AppConfig{
		App: AppSettings{
			Name: getEnv("APP_NAME", "notification-worker"),
			Env:  getEnv("APP_ENV", "development"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:           mustGetEnv("RABBITMQ_URL"),
			ExchangeName:  getEnv("RABBITMQ_EXCHANGE", "omni.events"),
			QueueName:     getEnv("RABBITMQ_QUEUE", "notification.worker.queue"),
			RoutingKey:    getEnv("RABBITMQ_ROUTING_KEY", "transaction.#"),
			PrefetchCount: prefetchCount,
		},
	}
}

// mustGetEnv reads an environment variable and panics if it is not set.
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return val
}

// getEnv reads an environment variable and returns fallback if it is not set.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
