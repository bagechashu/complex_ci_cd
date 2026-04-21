package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// Server
	ServerPort    int
	ServerHost    string
	Environment   string

	// Database
	DatabasePath  string

	// Encryption
	EncryptionKey string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerPort:    getEnvInt("SERVER_PORT", 8080),
		ServerHost:    getEnvStr("SERVER_HOST", "0.0.0.0"),
		Environment:   getEnvStr("ENVIRONMENT", "development"),
		DatabasePath:  getEnvStr("DATABASE_PATH", "./release_control.db"),
		EncryptionKey: getEnvStr("ENCRYPTION_KEY", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate port
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid port: %d", c.ServerPort)
	}

	// Validate environment
	if c.Environment != "development" && c.Environment != "staging" && c.Environment != "production" {
		return fmt.Errorf("invalid environment: %s (must be development, staging, or production)", c.Environment)
	}

	// Validate encryption key in production
	if c.Environment == "production" && c.EncryptionKey == "" {
		return errors.New("ENCRYPTION_KEY environment variable must be set in production")
	}

	// In development, set a default if not provided
	if c.EncryptionKey == "" {
		c.EncryptionKey = "dev-key-change-in-production"
	}

	return nil
}

// ServerAddr returns the server address as host:port
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

func getEnvStr(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return intVal
}
