package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	App       AppSettings
	JWT       JWTConfig
	Redis     RedisConfig
	Upstream  UpstreamConfig
	RateLimit RateLimitConfig
}

type AppSettings struct {
	Name string
	Port string
	Env  string
}

type JWTConfig struct {
	Secret string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type UpstreamConfig struct {
	UserServiceURL   string
	WalletServiceURL string
}

type RateLimitConfig struct {
	RequestsPerWindow int
	WindowDuration time.Duration
}

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
