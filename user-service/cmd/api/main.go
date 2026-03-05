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

	userhttp "github.com/omni-wallet/user-service/internal/adapter/handler/http"
	mysqlrepo "github.com/omni-wallet/user-service/internal/adapter/repository/mysql"
	redisrepo "github.com/omni-wallet/user-service/internal/adapter/repository/redis"
	"github.com/omni-wallet/user-service/internal/adapter/walletclient"
	"github.com/omni-wallet/user-service/internal/core/services"
	"github.com/omni-wallet/user-service/internal/platform/config"
	"github.com/omni-wallet/user-service/internal/platform/database"
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
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("[ERROR] Closing MySQL connection: %v", err)
		}
	}()
	log.Println("[INFO] Connected to MySQL")

	redisClient, err := database.NewRedisClient(
		cfg.Redis.RedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to Redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("[ERROR] Closing Redis connection: %v", err)
		}
	}()
	log.Println("[INFO] Connected to Redis")

	userRepo := mysqlrepo.NewUserRepository(db)
	cacheRepo := redisrepo.NewUserCacheRepository(redisClient)

	wClient := walletclient.New(cfg.Wallet.ServiceBaseURL)
	log.Printf("[INFO] Wallet service URL: %s", cfg.Wallet.ServiceBaseURL)

	userService := services.NewUserService(userRepo, cacheRepo, wClient, cfg.JWT.Secret, cfg.JWT.TTL)

	userHandler := userhttp.NewUserHandler(userService)

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
	userHandler.RegisterRoutes(apiV1)

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
