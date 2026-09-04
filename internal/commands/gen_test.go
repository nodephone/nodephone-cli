package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestGenCommand(t *testing.T) {
	tempProjectDir := t.TempDir()

	sampleOpenAPI := `{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {},
		"components": {
			"schemas": {
				"LoginRequest": {
					"type": "object",
					"properties": {
						"email": { "type": "string" }
					}
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health", "/health":
			w.WriteHeader(http.StatusOK)
		case "/docs/openapi.json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleOpenAPI))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	auth.SaveServerConfig(server.URL, "")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	ctx := &Context{
		Config:   &config.Config{ProjectRoot: tempProjectDir},
		Printer:  p,
		Registry: NewRegistry(),
	}

	genCmd := NewGenCommand()

	// Run nodephone gen types
	err := genCmd.Execute(ctx, []string{"types"})
	if err != nil {
		t.Fatalf("gen types failed: %v", err)
	}

	if !strings.Contains(out.String(), "Generated types/auth.ts") || !strings.Contains(out.String(), "TypeScript types successfully generated") {
		t.Errorf("unexpected output from gen types:\n%s", out.String())
	}

	// Check files created in types/
	typesDir := filepath.Join(tempProjectDir, "types")
	expectedFiles := []string{"auth.ts", "database.ts", "storage.ts", "functions.ts", "api.ts", "index.ts"}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(typesDir, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected generated file %s to exist", f)
		}
	}
}
