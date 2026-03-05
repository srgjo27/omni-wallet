package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	wallethttp "github.com/omni-wallet/wallet-service/internal/adapter/handler/http"
	"github.com/omni-wallet/wallet-service/internal/adapter/httpclient"
	rabbitmqpub "github.com/omni-wallet/wallet-service/internal/adapter/messaging/rabbitmq"
	mysqlrepo "github.com/omni-wallet/wallet-service/internal/adapter/repository/mysql"
	redisrepo "github.com/omni-wallet/wallet-service/internal/adapter/repository/redis"
	"github.com/omni-wallet/wallet-service/internal/core/ports"
	"github.com/omni-wallet/wallet-service/internal/core/services"
	"github.com/omni-wallet/wallet-service/internal/platform/config"
	"github.com/omni-wallet/wallet-service/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	db, err := database.NewMySQLConnection(
		cfg.Database.DSN(),
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to MySQL: %v", err)
	}
	defer db.Close()
	log.Println("[INFO] Connected to MySQL")

	redisClient, err := database.NewRedisClient(
		cfg.Redis.RedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("[INFO] Connected to Redis")

	walletRepo := mysqlrepo.NewWalletRepository(db)
	txRepo := mysqlrepo.NewTransactionRepository(db)
	mutationRepo := mysqlrepo.NewMutationRepository(db)
	txProvider := mysqlrepo.NewTxProvider(db)
	lockRepo := redisrepo.NewLockRepository(redisClient)
	idempotencyRepo := redisrepo.NewIdempotencyRepository(redisClient)
	userClient := httpclient.NewUserServiceClient(cfg.UserService.BaseURL)

	var eventPublisher ports.EventPublisher
	if cfg.RabbitMQ.Enabled {
		pub, err := rabbitmqpub.NewEventPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.ExchangeName)
		if err != nil {
			log.Printf("[WARN] RabbitMQ not available, event publishing disabled: %v", err)
		} else {
			defer pub.Close()
			eventPublisher = pub
			log.Println("[INFO] Connected to RabbitMQ")
		}
	}

	walletService := services.NewWalletService(walletRepo, mutationRepo, txRepo)
	transferService := services.NewTransferService(
		walletRepo,
		txRepo,
		mutationRepo,
		txProvider,
		lockRepo,
		idempotencyRepo,
		userClient,
		eventPublisher,
	)

	walletHandler := wallethttp.NewWalletHandler(walletService, transferService, cfg.JWT.Secret)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": cfg.App.Name})
	})

	apiV1 := router.Group("/api/v1")
	walletHandler.RegisterRoutes(apiV1)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[INFO] %s is listening on port %s", cfg.App.Name, cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[FATAL] Forced shutdown: %v", err)
	}
	log.Println("[INFO] Server exited cleanly")
}
