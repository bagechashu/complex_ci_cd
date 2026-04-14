package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	ServerPort     int
	ServerHost     string
	Environment    string

	// Database
	DatabasePath   string

	// Kubernetes
	KubeConfigPath string

	// Encryption
	EncryptionKey  string
}

func LoadConfig() *Config {
	return &Config{
		ServerPort:     getEnvInt("SERVER_PORT", 8080),
		ServerHost:     getEnvStr("SERVER_HOST", "0.0.0.0"),
		Environment:    getEnvStr("ENVIRONMENT", "development"),
		DatabasePath:   getEnvStr("DATABASE_PATH", "./release_control.db"),
		KubeConfigPath: getEnvStr("KUBECONFIG", os.ExpandEnv("$HOME/.kube/config")),
		EncryptionKey:  getEnvStr("ENCRYPTION_KEY", "default-key-change-in-production"),
	}
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
