package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv   string
	Port     int
	LogLevel string

	Database DatabaseConfig
	Redis    RedisConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime int
	MaxConnIdleTime int
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     int
	CacheTTL int
}

func Load() (Config, error) {
	httpPort, err := getIntEnv("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, fmt.Errorf("load HTTP_PORT: %w", err)
	}

	if httpPort < 1 || httpPort > 65535 {
		return Config{}, fmt.Errorf(
			"HTTP_PORT must be between 1 and 65535",
		)
	}

	databasePort, err := getIntEnv("DATABASE_PORT", 5432)
	if err != nil {
		return Config{}, fmt.Errorf("load DATABASE_PORT: %w", err)
	}

	maxConns, err := getIntEnv("DATABASE_MAX_CONNS", 10)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load DATABASE_MAX_CONNS: %w",
			err,
		)
	}

	minConns, err := getIntEnv("DATABASE_MIN_CONNS", 2)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load DATABASE_MIN_CONNS: %w",
			err,
		)
	}

	maxConnLifetime, err := getIntEnv(
		"DATABASE_MAX_CONN_LIFETIME_MINUTES",
		30,
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load DATABASE_MAX_CONN_LIFETIME_MINUTES: %w",
			err,
		)
	}

	maxConnIdleTime, err := getIntEnv(
		"DATABASE_MAX_CONN_IDLE_TIME_MINUTES",
		5,
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load DATABASE_MAX_CONN_IDLE_TIME_MINUTES: %w",
			err,
		)
	}

	if maxConns < 1 {
		return Config{}, fmt.Errorf(
			"DATABASE_MAX_CONNS must be greater than 0",
		)
	}

	if minConns < 0 || minConns > maxConns {
		return Config{}, fmt.Errorf(
			"DATABASE_MIN_CONNS must be between 0 and DATABASE_MAX_CONNS",
		)
	}

	redisPort, err := getIntEnv("REDIS_PORT", 6379)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load REDIS_PORT: %w",
			err,
		)
	}

	redisCacheTTL, err := getIntEnv(
		"REDIS_CACHE_TTL_SECONDS",
		60,
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load REDIS_CACHE_TTL_SECONDS: %w",
			err,
		)
	}

	if redisPort < 1 || redisPort > 65535 {
		return Config{}, fmt.Errorf(
			"REDIS_PORT must be between 1 and 65535",
		)
	}

	if redisCacheTTL < 1 {
		return Config{}, fmt.Errorf(
			"REDIS_CACHE_TTL_SECONDS must be greater than 0",
		)
	}

	return Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		Port:     httpPort,
		LogLevel: getEnv("LOG_LEVEL", "info"),

		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            databasePort,
			User:            getEnv("DATABASE_USER", "task_manager"),
			Password:        getEnv("DATABASE_PASSWORD", "task_manager"),
			Name:            getEnv("DATABASE_NAME", "task_manager"),
			SSLMode:         getEnv("DATABASE_SSLMODE", "disable"),
			MaxConns:        int32(maxConns),
			MinConns:        int32(minConns),
			MaxConnLifetime: maxConnLifetime,
			MaxConnIdleTime: maxConnIdleTime,
		},

		Redis: RedisConfig{
			Enabled:  getBoolEnv("REDIS_ENABLED", true),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			CacheTTL: redisCacheTTL,
		},
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getIntEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)

	if value == "" {
		return fallback, nil
	}

	return strconv.Atoi(value)
}

func getBoolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
