package config

import (
	"fmt"
	"os"
	"time"
)

// Config contains application configuration.
type Config struct {
	AppEnvironment  string
	HTTPAddress     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	DatabaseURL     string
	UploadDirectory string
}

// Load reads application configuration from environment variables.
func Load() (Config, error) {
	readTimeout, err := getDuration("HTTP_READ_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := getDuration("HTTP_WRITE_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}

	idleTimeout, err := getDuration("HTTP_IDLE_TIMEOUT", "60s")
	if err != nil {
		return Config{}, err
	}

	port := getEnvironmentVariable("APP_PORT", "8081")

	return Config{
		AppEnvironment: getEnvironmentVariable(
			"APP_ENV",
			"development",
		),
		HTTPAddress:  ":" + port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		DatabaseURL: getEnvironmentVariable(
			"DATABASE_URL",
			"postgres://student_app:student_secret@localhost:5432/student_enrollment?sslmode=disable",
		),
		UploadDirectory: getEnvironmentVariable(
			"UPLOAD_DIR",
			"./uploads",
		),
	}, nil
}

// getEnvironmentVariable returns an environment variable or its default value.
func getEnvironmentVariable(
	key string,
	defaultValue string,
) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

// getDuration reads and parses a duration environment variable.
func getDuration(
	key string,
	defaultValue string,
) (time.Duration, error) {
	value := getEnvironmentVariable(key, defaultValue)

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid duration for %s: %w",
			key,
			err,
		)
	}

	return duration, nil
}
