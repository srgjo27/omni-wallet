package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	UserService UserServiceConfig
	RabbitMQ    RabbitMQConfig
	Xendit      XenditConfig
}

type AppConfig struct {
	Name string
	Port string
	Env  string
}

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

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
}

type UserServiceConfig struct {
	BaseURL string
}

type RabbitMQConfig struct {
	URL          string
	ExchangeName string
	Enabled bool
}

type XenditConfig struct {
	SecretKey string
	WebhookToken string
	CallbackURL string
	Enabled bool
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

func (r RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnvOrDefault("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	dbMaxOpen, err := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}
	dbMaxIdle, err := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}
	dbConnLifetime, err := time.ParseDuration(getEnvOrDefault("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}

	return &Config{
		App: AppConfig{
			Name: getEnvOrDefault("APP_NAME", "wallet-service"),
			Port: getEnvOrDefault("APP_PORT", "8082"),
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
		},
		UserService: UserServiceConfig{
			BaseURL: mustGetEnv("USER_SERVICE_BASE_URL"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:          getEnvOrDefault("RABBITMQ_URL", ""),
			ExchangeName: getEnvOrDefault("RABBITMQ_EXCHANGE", "omni.events"),
			Enabled:      getEnvOrDefault("RABBITMQ_URL", "") != "",
		},
		Xendit: XenditConfig{
			SecretKey:    getEnvOrDefault("XENDIT_SECRET_KEY", ""),
			WebhookToken: getEnvOrDefault("XENDIT_WEBHOOK_TOKEN", ""),
			CallbackURL:  getEnvOrDefault("XENDIT_CALLBACK_URL", ""),
			Enabled:      getEnvOrDefault("XENDIT_SECRET_KEY", "") != "",
		},
	}, nil
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
