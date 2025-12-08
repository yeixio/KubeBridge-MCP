package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	ClusterName          string
	Kubeconfig          string
	EnableWriteOperations bool
	LogLevel            string
}

// LoadConfig loads configuration from environment variables and returns defaults
func LoadConfig() *Config {
	cfg := &Config{
		ClusterName:          getEnv("KUBEBRIDGE_CLUSTER_NAME", ""),
		Kubeconfig:          getKubeconfigPath(),
		EnableWriteOperations: getEnvBool("KUBEBRIDGE_ENABLE_WRITE", false),
		LogLevel:            getEnv("KUBEBRIDGE_LOG_LEVEL", "info"),
	}
	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// getKubeconfigPath returns the kubeconfig path from environment or default location
func getKubeconfigPath() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}


