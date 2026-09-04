package functions

import (
	"bufio"
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
			Timeout: 60 * time.Second,
		},
	}
}

type RemoteFunction struct {
	Name      string    `json:"name"`
	Runtime   string    `json:"runtime"`
	Checksum  string    `json:"checksum"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeployPayload struct {
	Functions []FunctionInfo `json:"functions"`
}

type DeployResponse struct {
	Deployed []string `json:"deployed"`
	Success  bool     `json:"success"`
	Message  string   `json:"message,omitempty"`
}

// ListRemoteFunctions queries backend for deployed serverless functions
func (c *Client) ListRemoteFunctions(serverURL, token string) ([]RemoteFunction, error) {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	endpoints := []string{
		serverURL + "/api/v1/functions",
		serverURL + "/functions",
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
		return nil, fmt.Errorf("failed to query remote functions from %s: %w", serverURL, lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status code %d when checking functions", lastResp.StatusCode)
	}

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, err
	}

	var funcs []RemoteFunction
	if err := json.Unmarshal(respBytes, &funcs); err != nil {
		return nil, fmt.Errorf("failed to parse functions response: %w", err)
	}

	return funcs, nil
}

// DeployFunctions uploads modified functions to backend server
func (c *Client) DeployFunctions(serverURL, token string, funcs []FunctionInfo) ([]string, error) {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	payload := DeployPayload{Functions: funcs}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoints := []string{
		serverURL + "/api/v1/functions/deploy",
		serverURL + "/functions/deploy",
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
		return nil, fmt.Errorf("failed to reach functions deploy endpoint: %w", lastErr)
	}
	defer lastResp.Body.Close()

	respBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, err
	}

	if lastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("function deployment failed (status %d)", lastResp.StatusCode)
	}

	var deployResp DeployResponse
	if err := json.Unmarshal(respBytes, &deployResp); err != nil {
		return nil, err
	}

	return deployResp.Deployed, nil
}

// DeleteFunction calls backend API to remove deployed function
func (c *Client) DeleteFunction(serverURL, token, name string) error {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	endpoints := []string{
		fmt.Sprintf("%s/api/v1/functions/%s", serverURL, name),
		fmt.Sprintf("%s/functions/%s", serverURL, name),
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		req, err := http.NewRequest("DELETE", ep, nil)
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
		return fmt.Errorf("failed to reach delete function endpoint: %w", lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK && lastResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete function failed (status %d)", lastResp.StatusCode)
	}

	return nil
}

// StreamLogs streams real-time logs for a function to stdout
func (c *Client) StreamLogs(serverURL, token, name string, out io.Writer) error {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	url := fmt.Sprintf("%s/api/v1/functions/%s/logs", serverURL, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to log stream for %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to stream logs (status %d)", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Fprintln(out, scanner.Text())
	}

	return scanner.Err()
}
