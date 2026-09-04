package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestAuthCommandsFlow(t *testing.T) {
	tempHome := t.TempDir()
	os.Setenv("USERPROFILE", tempHome)
	os.Setenv("HOME", tempHome)

	// Mock Auth HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health", "/health":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/auth/login":
			var req auth.LoginRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Email == "user@nodephone.dev" && req.Password == "password123" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(auth.LoginResponse{
					UserID:       "usr_42",
					Email:        req.Email,
					AccessToken:  "mock-access-token",
					RefreshToken: "mock-refresh-token",
					ExpiresIn:    3600,
				})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	ctx := &Context{
		Config:   &config.Config{ProjectRoot: tempHome},
		Printer:  p,
		Registry: NewRegistry(),
	}

	loginCmd := NewLoginCommand()
	whoamiCmd := NewWhoamiCommand()
	logoutCmd := NewLogoutCommand()

	// 1. whoami before login -> Not authenticated
	out.Reset()
	if err := whoamiCmd.Execute(ctx, nil); err != nil {
		t.Fatalf("whoami failed: %v", err)
	}
	if !strings.Contains(out.String(), "Not authenticated") {
		t.Errorf("expected 'Not authenticated' before login, got:\n%s", out.String())
	}

	// 2. login command with flags
	out.Reset()
	loginArgs := []string{"--server", server.URL, "--email", "user@nodephone.dev", "--password", "password123"}
	if err := loginCmd.Execute(ctx, loginArgs); err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if !strings.Contains(out.String(), "user@nodephone.dev") || !strings.Contains(out.String(), "Connected") {
		t.Errorf("expected successful login output, got:\n%s", out.String())
	}

	// 3. whoami after login -> Connected
	out.Reset()
	if err := whoamiCmd.Execute(ctx, nil); err != nil {
		t.Fatalf("whoami failed after login: %v", err)
	}
	if !strings.Contains(out.String(), "user@nodephone.dev") || !strings.Contains(out.String(), "Connected") {
		t.Errorf("expected connected state in whoami output, got:\n%s", out.String())
	}

	// 4. logout command
	out.Reset()
	if err := logoutCmd.Execute(ctx, nil); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if !strings.Contains(out.String(), "Successfully logged out") {
		t.Errorf("expected successful logout output, got:\n%s", out.String())
	}

	// 5. whoami after logout -> Not authenticated
	out.Reset()
	if err := whoamiCmd.Execute(ctx, nil); err != nil {
		t.Fatalf("whoami failed after logout: %v", err)
	}
	if !strings.Contains(out.String(), "Not authenticated") {
		t.Errorf("expected 'Not authenticated' after logout, got:\n%s", out.String())
	}
}
