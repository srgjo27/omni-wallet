package config

import (
	"fmt"
	"os"
)

type AppConfig struct {
	App      AppSettings
	RabbitMQ RabbitMQConfig
}

type AppSettings struct {
	Name string
	Env  string
}

type RabbitMQConfig struct {
	URL          string
	ExchangeName string
	QueueName    string
	RoutingKey   string
	PrefetchCount int
}

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

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
