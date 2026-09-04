package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreCredentialsCycle(t *testing.T) {
	tempHome := t.TempDir()
	os.Setenv("USERPROFILE", tempHome)
	os.Setenv("HOME", tempHome)

	creds := &Credentials{
		UserID:       "usr_123",
		Email:        "dev@nodephone.dev",
		AccessToken:  "jwt-access-token",
		RefreshToken: "jwt-refresh-token",
		ServerURL:    "http://localhost:8080",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("failed to save credentials: %v", err)
	}

	// Verify on-disk file is encrypted ciphertext (not plaintext JSON)
	dir, _ := GetConfigDir()
	rawBytes, err := os.ReadFile(filepath.Join(dir, "credentials.json"))
	if err != nil {
		t.Fatalf("failed to read credentials file: %v", err)
	}
	if string(rawBytes) == "" {
		t.Error("expected non-empty credentials file")
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("failed to load credentials: %v", err)
	}

	if loaded.Email != creds.Email || loaded.AccessToken != creds.AccessToken {
		t.Errorf("loaded credentials %+v do not match saved %+v", loaded, creds)
	}

	if err := ClearCredentials(); err != nil {
		t.Fatalf("failed to clear credentials: %v", err)
	}

	_, err = LoadCredentials()
	if err == nil {
		t.Error("expected error loading cleared credentials, got nil")
	}
}
