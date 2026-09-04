package functions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// FunctionManifest represents function.json specification
type FunctionManifest struct {
	Name        string            `json:"name"`
	Runtime     string            `json:"runtime"`
	Entrypoint  string            `json:"entrypoint"`
	Timeout     int               `json:"timeout,omitempty"` // in seconds
	Environment map[string]string `json:"environment,omitempty"`
}

// DefaultManifest returns a standard function.json manifest for a new function
func DefaultManifest(name string) FunctionManifest {
	return FunctionManifest{
		Name:       name,
		Runtime:    "nodejs18",
		Entrypoint: "index.js",
		Timeout:    10,
		Environment: map[string]string{
			"NODE_ENV": "development",
		},
	}
}

// ValidateManifest checks manifest parameters
func (m *FunctionManifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("function name cannot be empty")
	}
	if strings.TrimSpace(m.Entrypoint) == "" {
		return errors.New("function entrypoint cannot be empty")
	}
	return nil
}

// StarterTemplate returns sample index.js content for a new function
func StarterTemplate(name string) string {
	return fmt.Sprintf(`/**
 * NodePhone Function: %s
 * Created via 'nodephone functions new %s'
 */
module.exports = async function handler(req, res) {
  const queryName = req.query?.name || 'Developer';

  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json'
    },
    body: {
      message: `+"`"+`Hello ${queryName} from %s function!`+"`"+`,
      timestamp: new Date().toISOString()
    }
  };
};
`, name, name, name)
}

// MarshalManifest returns pretty-printed JSON of function.json
func MarshalManifest(m FunctionManifest) (string, error) {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes) + "\n", nil
}
