package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AppConfig holds all configuration loaded from environment variables.
type AppConfig struct {
	App       AppSettings
	JWT       JWTConfig
	Redis     RedisConfig
	Upstream  UpstreamConfig
	RateLimit RateLimitConfig
}

// AppSettings contains general application settings.
type AppSettings struct {
	Name string
	Port string
	Env  string
}

// JWTConfig contains JWT verification settings (shared secret with User Service).
type JWTConfig struct {
	Secret string
}

// RedisConfig contains Redis connection settings for distributed rate limiting.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// UpstreamConfig contains the base URLs of downstream microservices.
type UpstreamConfig struct {
	UserServiceURL   string
	WalletServiceURL string
}

// RateLimitConfig defines the sliding-window rate limit applied per IP.
type RateLimitConfig struct {
	// RequestsPerWindow is the maximum number of requests allowed per IP.
	RequestsPerWindow int
	// WindowDuration is the rolling window size (e.g., 1m).
	WindowDuration time.Duration
}

// Load reads all required configuration from environment variables.
// Panics immediately on missing critical configuration so the service
// fails fast at startup rather than at runtime.
func Load() *AppConfig {
	rateRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "60"))
	windowSecs, _ := strconv.Atoi(getEnv("RATE_LIMIT_WINDOW_SECONDS", "60"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &AppConfig{
		App: AppSettings{
			Name: getEnv("APP_NAME", "api-gateway"),
			Port: mustGetEnv("APP_PORT"),
			Env:  getEnv("APP_ENV", "development"),
		},
		JWT: JWTConfig{
			Secret: mustGetEnv("JWT_SECRET"),
		},
		Redis: RedisConfig{
			Host:     mustGetEnv("REDIS_HOST"),
			Port:     mustGetEnv("REDIS_PORT"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Upstream: UpstreamConfig{
			UserServiceURL:   mustGetEnv("USER_SERVICE_URL"),
			WalletServiceURL: mustGetEnv("WALLET_SERVICE_URL"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerWindow: rateRequests,
			WindowDuration:    time.Duration(windowSecs) * time.Second,
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
