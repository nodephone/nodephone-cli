package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nodephone/nodephone-cli/internal/auth"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetStatus queries remote migration status from NodePhone server
func (c *Client) GetStatus(serverURL, token string) ([]MigrationRecord, error) {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	endpoints := []string{
		serverURL + "/api/v1/db/migrations",
		serverURL + "/db/migrations",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("GET", ep, nil)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

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
		return nil, fmt.Errorf("failed to query database migrations from %s: %w", serverURL, lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d when checking migrations", lastResp.StatusCode)
	}

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, err
	}

	var records []MigrationRecord
	if err := json.Unmarshal(respBytes, &records); err != nil {
		return nil, fmt.Errorf("failed to parse migration response: %w", err)
	}

	return records, nil
}

// PushMigrations sends pending migrations to server for execution
func (c *Client) PushMigrations(serverURL, token string, pending []MigrationFile) ([]string, error) {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	payload := PushRequest{Migrations: pending}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoints := []string{
		serverURL + "/api/v1/db/push",
		serverURL + "/db/push",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("POST", ep, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

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
		return nil, fmt.Errorf("failed to reach push migration endpoint: %w", lastErr)
	}
	defer lastResp.Body.Close()

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, err
	}

	if lastResp.StatusCode != http.StatusOK {
		var errResp PushResponse
		if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("migration push failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("migration push failed (status %d)", lastResp.StatusCode)
	}

	var pushResp PushResponse
	if err := json.Unmarshal(respBytes, &pushResp); err != nil {
		return nil, err
	}

	return pushResp.Applied, nil
}

// PullSchema fetches remote schema migration files from server
func (c *Client) PullSchema(serverURL, token string) ([]MigrationFile, error) {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	endpoints := []string{
		serverURL + "/api/v1/db/pull",
		serverURL + "/db/pull",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("GET", ep, nil)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

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
		return nil, fmt.Errorf("failed to reach pull schema endpoint: %w", lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("schema pull failed (status %d)", lastResp.StatusCode)
	}

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, err
	}

	var pullResp PullResponse
	if err := json.Unmarshal(respBytes, &pullResp); err != nil {
		return nil, err
	}

	return pullResp.Migrations, nil
}

// ResetDatabase sends reset request to backend server
func (c *Client) ResetDatabase(serverURL, token string) error {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	endpoints := []string{
		serverURL + "/api/v1/db/reset",
		serverURL + "/db/reset",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("POST", ep, nil)
		if err != nil {
			return err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

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
		return fmt.Errorf("failed to reach reset database endpoint: %w", lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK {
		return fmt.Errorf("database reset failed (status %d)", lastResp.StatusCode)
	}

	return nil
}
