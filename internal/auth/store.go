package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigStore represents local user preferences stored in ~/.nodephone/config.json
type ConfigStore struct {
	ServerURL        string `json:"server_url"`
	ActiveWorkspace string `json:"active_workspace,omitempty"`
}

// GetConfigDir returns the path to ~/.nodephone/
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to locate home directory: %w", err)
	}
	dir := filepath.Join(home, ".nodephone")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return dir, nil
}

// SaveCredentials encrypts and writes credentials to ~/.nodephone/credentials.json
func SaveCredentials(creds *Credentials) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	encrypted, err := Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	credPath := filepath.Join(dir, "credentials.json")
	return os.WriteFile(credPath, encrypted, 0600)
}

// LoadCredentials reads and decrypts credentials from ~/.nodephone/credentials.json
func LoadCredentials() (*Credentials, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	credPath := filepath.Join(dir, "credentials.json")
	encrypted, err := os.ReadFile(credPath)
	if os.IsNotExist(err) {
		return nil, errors.New("not authenticated")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// ClearCredentials removes stored credentials file (logout)
func ClearCredentials() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	credPath := filepath.Join(dir, "credentials.json")
	err = os.Remove(credPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// SaveServerConfig writes server config settings to ~/.nodephone/config.json
func SaveServerConfig(serverURL string, workspace string) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	cfg := ConfigStore{
		ServerURL:        serverURL,
		ActiveWorkspace: workspace,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(dir, "config.json")
	return os.WriteFile(configPath, data, 0600)
}

// LoadServerConfig reads ~/.nodephone/config.json
func LoadServerConfig() (*ConfigStore, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &ConfigStore{ServerURL: "http://localhost:8080"}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg ConfigStore
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = "http://localhost:8080"
	}

	return &cfg, nil
}
