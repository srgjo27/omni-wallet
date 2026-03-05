package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	rabbitmqconsumer "github.com/omni-wallet/notification-worker/internal/adapter/consumer/rabbitmq"
	"github.com/omni-wallet/notification-worker/internal/adapter/sender"
	"github.com/omni-wallet/notification-worker/internal/core/services"
	"github.com/omni-wallet/notification-worker/internal/platform/config"
	"github.com/omni-wallet/notification-worker/internal/platform/messaging"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	log.Printf("[notification-worker] starting (env=%s)", cfg.App.Env)

	rabbitConn, err := messaging.NewRabbitMQConnection(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("[notification-worker] failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	consumer, err := rabbitmqconsumer.NewConsumer(rabbitConn, rabbitmqconsumer.ConsumerConfig{
		ExchangeName:  cfg.RabbitMQ.ExchangeName,
		QueueName:     cfg.RabbitMQ.QueueName,
		RoutingKey:    cfg.RabbitMQ.RoutingKey,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	})
	if err != nil {
		log.Fatalf("[notification-worker] failed to create consumer: %v", err)
	}
	defer consumer.Close()

	notificationSender := sender.NewLogNotificationSender()
	notificationService := services.NewNotificationService(consumer, notificationSender)

	errCh := make(chan error, 1)
	go func() {
		errCh <- notificationService.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("[notification-worker] received signal %s — shutting down", sig)
	case err := <-errCh:
		log.Printf("[notification-worker] consumer error: %v — shutting down", err)
	}

	log.Println("[notification-worker] stopped")
}
