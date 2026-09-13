package config

import (
	"testing"
	"time"
)

func TestLoadDefaultConfig(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_PORT", "")
	t.Setenv("HTTP_READ_TIMEOUT", "")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("HTTP_IDLE_TIMEOUT", "")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if cfg.AppEnvironment != "development" {
		t.Errorf(
			"expected development environment, but got %q",
			cfg.AppEnvironment,
		)
	}

	if cfg.HTTPAddress != ":8081" {
		t.Errorf(
			"expected HTTP address :8081, but got %q",
			cfg.HTTPAddress,
		)
	}

	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf(
			"expected read timeout 10s, but got %s",
			cfg.ReadTimeout,
		)
	}
}

func TestLoadCustomConfig(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "6s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "30s")
	t.Setenv(
		"DATABASE_URL",
		"postgres://test:test@localhost:5432/test",
	)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if cfg.AppEnvironment != "test" {
		t.Errorf(
			"expected test environment, but got %q",
			cfg.AppEnvironment,
		)
	}

	if cfg.HTTPAddress != ":9090" {
		t.Errorf(
			"expected HTTP address :9090, but got %q",
			cfg.HTTPAddress,
		)
	}

	if cfg.WriteTimeout != 6*time.Second {
		t.Errorf(
			"expected write timeout 6s, but got %s",
			cfg.WriteTimeout,
		)
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT", "invalid-duration")

	_, err := Load()

	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}
