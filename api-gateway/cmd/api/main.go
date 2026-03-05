package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/omni-wallet/api-gateway/internal/adapter/middleware"
	"github.com/omni-wallet/api-gateway/internal/adapter/proxy"
	rediscache "github.com/omni-wallet/api-gateway/internal/platform/cache"
	"github.com/omni-wallet/api-gateway/internal/platform/config"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	redisClient, err := rediscache.NewRedisClient(rediscache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatalf("[api-gateway] failed to connect to Redis: %v", err)
	}
	log.Println("[api-gateway] Redis connected")

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	engine.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "service": cfg.App.Name})
	})

	authMW := middleware.AuthMiddleware(cfg.JWT.Secret)
	rateMW := middleware.RateLimiter(redisClient, cfg.RateLimit.RequestsPerWindow, cfg.RateLimit.WindowDuration)

	if err := proxy.RegisterRoutes(engine, authMW, rateMW,
		cfg.Upstream.UserServiceURL,
		cfg.Upstream.WalletServiceURL,
	); err != nil {
		log.Fatalf("[api-gateway] failed to register routes: %v", err)
	}

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[api-gateway] listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[api-gateway] server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[api-gateway] shutting down gracefully…")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[api-gateway] forced shutdown: %v", err)
	}
	log.Println("[api-gateway] stopped")
}
