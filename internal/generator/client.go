package generator

import (
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
			Timeout: 20 * time.Second,
		},
	}
}

// FetchOpenAPISpec downloads the OpenAPI 3.1 JSON document from server
func (c *Client) FetchOpenAPISpec(serverURL string) ([]byte, error) {
	serverURL = auth.EnsureURLHasScheme(serverURL)

	endpoints := []string{
		serverURL + "/docs/openapi.json",
		serverURL + "/api/v1/docs/openapi.json",
		serverURL + "/openapi.json",
	}

	var lastResp *http.Response
	var lastErr error

	for _, ep := range endpoints {
		resp, err := c.httpClient.Get(ep)
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
		return nil, fmt.Errorf("failed to download OpenAPI spec from %s: %w", serverURL, lastErr)
	}
	defer lastResp.Body.Close()

	if lastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status code %d when fetching OpenAPI spec", lastResp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI response: %w", err)
	}

	return bodyBytes, nil
}
