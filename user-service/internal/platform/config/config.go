package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config centralises all runtime configuration loaded from environment variables.
// No secret or credential is ever hardcoded here.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Wallet   WalletConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name string
	Port string
	Env  string // "development" | "production"
}

// DatabaseConfig holds MySQL connection settings.
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret string
	TTL    time.Duration // e.g. 24h
}

// WalletConfig holds the URL used to call the wallet-service.
type WalletConfig struct {
	ServiceBaseURL string
}

// DSN builds the MySQL Data Source Name string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// RedisAddr returns the Redis host:port address.
func (r RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// Load reads configuration from environment variables. Call this once at startup.
// It will also attempt to load a .env file if it exists (useful for local development).
func Load() (*Config, error) {
	// Best-effort load of .env file; ignored if not found (e.g. in production containers).
	_ = godotenv.Load()

	jwtTTL, err := time.ParseDuration(getEnvOrDefault("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL value: %w", err)
	}

	redisDB, err := strconv.Atoi(getEnvOrDefault("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB value: %w", err)
	}

	dbMaxOpen, err := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS value: %w", err)
	}

	dbMaxIdle, err := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS value: %w", err)
	}

	dbConnLifetime, err := time.ParseDuration(getEnvOrDefault("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME value: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name: getEnvOrDefault("APP_NAME", "user-service"),
			Port: getEnvOrDefault("APP_PORT", "8081"),
			Env:  getEnvOrDefault("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:            mustGetEnv("DB_HOST"),
			Port:            getEnvOrDefault("DB_PORT", "3306"),
			User:            mustGetEnv("DB_USER"),
			Password:        mustGetEnv("DB_PASSWORD"),
			Name:            mustGetEnv("DB_NAME"),
			MaxOpenConns:    dbMaxOpen,
			MaxIdleConns:    dbMaxIdle,
			ConnMaxLifetime: dbConnLifetime,
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     getEnvOrDefault("REDIS_PORT", "6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret: mustGetEnv("JWT_SECRET"),
			TTL:    jwtTTL,
		},
		Wallet: WalletConfig{
			ServiceBaseURL: getEnvOrDefault("WALLET_SERVICE_BASE_URL", "http://wallet-service:8082"),
		},
	}

	return cfg, nil
}

// mustGetEnv panics if the given environment variable is not set,
// ensuring misconfigured deployments are caught immediately at startup.
func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return value
}

// getEnvOrDefault returns the value of the environment variable or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
