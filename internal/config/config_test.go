package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	os.Setenv("NODEPHONE_ENV", "production")
	defer os.Unsetenv("NODEPHONE_ENV")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected environment 'production', got '%s'", cfg.Environment)
	}
}
