package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("expected development, got %q", cfg.AppEnv)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Port)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("expected info, got %q", cfg.LogLevel)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("expected test, got %q", cfg.AppEnv)
	}

	if cfg.Port != 9090 {
		t.Fatalf("expected 9090, got %d", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("expected debug, got %q", cfg.LogLevel)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_PortOutOfRange(t *testing.T) {
	tests := []string{
		"0",
		"65536",
	}

	for _, port := range tests {
		t.Run(port, func(t *testing.T) {
			t.Setenv("HTTP_PORT", port)

			_, err := Load()
			if err == nil {
				t.Fatalf("expected error for port %s", port)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	const key = "MANAGER_TASK_TEST_ENV"

	_ = os.Unsetenv(key)

	if got := getEnv(key, "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}

	t.Setenv(key, "configured")

	if got := getEnv(key, "fallback"); got != "configured" {
		t.Fatalf("expected configured, got %q", got)
	}
}
