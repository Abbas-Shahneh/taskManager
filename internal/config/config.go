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
}

func Load() (Config, error) {
	port, err := getIntEnv("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, fmt.Errorf("load HTTP_PORT: %w", err)
	}

	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}

	return Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		Port:     port,
		LogLevel: getEnv("LOG_LEVEL", "info"),
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
