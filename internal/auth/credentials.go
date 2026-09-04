package auth

import (
	"time"
)

// Credentials holds authenticated session details.
type Credentials struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ServerURL    string    `json:"server_url"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired checks if the access token has expired or is about to expire.
func (c *Credentials) IsExpired() bool {
	if c.AccessToken == "" {
		return true
	}
	// Consider token expired 30 seconds ahead to account for network latency
	return time.Now().Add(30 * time.Second).After(c.ExpiresAt)
}
