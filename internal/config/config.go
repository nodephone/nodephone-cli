package config

import (
	"os"
	"path/filepath"
)

// Config represents global CLI configuration.
type Config struct {
	ConfigDir   string `json:"configDir"`
	ProjectRoot string `json:"projectRoot"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// Load loads or creates CLI configuration from default paths and environment variables.
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".nodephone")

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}

	cfg := &Config{
		ConfigDir:   configDir,
		ProjectRoot: workDir,
		Environment: getEnvOrDefault("NODEPHONE_ENV", "development"),
		Debug:       os.Getenv("NODEPHONE_DEBUG") == "true" || os.Getenv("NODEPHONE_DEBUG") == "1",
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
