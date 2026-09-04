package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientLoginAndPing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health", "/health":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/auth/login":
			var req LoginRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Email == "dev@nodephone.dev" && req.Password == "secret123" {
				json.NewEncoder(w).Encode(LoginResponse{
					UserID:       "usr_99",
					Email:        req.Email,
					AccessToken:  "mock-jwt-token",
					RefreshToken: "mock-refresh-token",
					ExpiresIn:    3600,
				})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(LoginResponse{Error: "invalid credentials"})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient()

	if err := client.PingServer(server.URL); err != nil {
		t.Fatalf("PingServer failed: %v", err)
	}

	creds, err := client.Login(server.URL, "dev@nodephone.dev", "secret123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if creds.UserID != "usr_99" || creds.AccessToken != "mock-jwt-token" {
		t.Errorf("unexpected credentials returned: %+v", creds)
	}

	// Invalid password
	_, err = client.Login(server.URL, "dev@nodephone.dev", "wrongpass")
	if err == nil {
		t.Error("expected login error for wrong password, got nil")
	}
}
