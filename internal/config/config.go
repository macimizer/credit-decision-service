package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName         string
	HTTPAddr        string
	LogLevel        string
	PostgresDSN     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	CacheTTL        time.Duration
	EventStreamName string
	WorkerCount     int
	AutoMigrate     bool
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		AppName:         getEnv("APP_NAME", "credit-decision-service"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		PostgresDSN:     getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/credit_service?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvAsInt("REDIS_DB", 0),
		CacheTTL:        getEnvAsDuration("CACHE_TTL", 5*time.Minute),
		EventStreamName: getEnv("EVENT_STREAM_NAME", "credit-events"),
		WorkerCount:     getEnvAsInt("WORKER_COUNT", 4),
		AutoMigrate:     getEnvAsBool("AUTO_MIGRATE", true),
		ReadTimeout:     getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
