package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
	Error        string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// EnsureURLHasScheme prepends http:// if missing
func EnsureURLHasScheme(serverURL string) string {
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		return "http://localhost:8080"
	}
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		return "http://" + serverURL
	}
	return serverURL
}

// PingServer checks if server is reachable at health endpoint
func (c *Client) PingServer(serverURL string) error {
	serverURL = EnsureURLHasScheme(serverURL)

	endpoints := []string{
		serverURL + "/api/v1/health",
		serverURL + "/health",
		serverURL + "/",
	}

	var lastErr error
	for _, ep := range endpoints {
		resp, err := c.httpClient.Get(ep)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
			lastErr = fmt.Errorf("server responded with status code %d", resp.StatusCode)
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("server unreachable at %s: %w", serverURL, lastErr)
}

// Login authenticates user against backend API
func (c *Client) Login(serverURL, email, password string) (*Credentials, error) {
	serverURL = EnsureURLHasScheme(serverURL)

	if err := c.PingServer(serverURL); err != nil {
		return nil, err
	}

	payload := LoginRequest{
		Email:    email,
		Password: password,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoints := []string{
		serverURL + "/api/v1/auth/login",
		serverURL + "/auth/login",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("POST", ep, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			continue
		}

		lastResp = resp
		break
	}

	if lastResp == nil {
		return nil, fmt.Errorf("failed to contact authentication endpoint at %s: %w", serverURL, lastErr)
	}
	defer lastResp.Body.Close()

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if lastResp.StatusCode != http.StatusOK {
		var errResp LoginResponse
		if json.Unmarshal(respBytes, &errResp) == nil {
			if errResp.Error != "" {
				return nil, errors.New(errResp.Error)
			}
			if errResp.Message != "" {
				return nil, errors.New(errResp.Message)
			}
		}
		return nil, fmt.Errorf("authentication failed (status %d)", lastResp.StatusCode)
	}

	var authResp LoginResponse
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return nil, errors.New("server returned empty access token")
	}

	expiresIn := authResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600 // Default 1 hour
	}

	creds := &Credentials{
		UserID:       authResp.UserID,
		Email:        email,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ServerURL:    serverURL,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	return creds, nil
}

// RefreshToken attempts to renew access token using refresh token
func (c *Client) RefreshToken(creds *Credentials) (*Credentials, error) {
	if creds == nil || creds.RefreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	serverURL := EnsureURLHasScheme(creds.ServerURL)

	payload := RefreshRequest{
		RefreshToken: creds.RefreshToken,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoints := []string{
		serverURL + "/api/v1/auth/refresh",
		serverURL + "/auth/refresh",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("POST", ep, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			continue
		}

		lastResp = resp
		break
	}

	if lastResp == nil {
		return nil, fmt.Errorf("failed to contact token refresh endpoint: %w", lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (status %d)", lastResp.StatusCode)
	}

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, err
	}

	var authResp LoginResponse
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		return nil, err
	}

	expiresIn := authResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	creds.AccessToken = authResp.AccessToken
	if authResp.RefreshToken != "" {
		creds.RefreshToken = authResp.RefreshToken
	}
	creds.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return creds, nil
}
