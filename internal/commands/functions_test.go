package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/functions"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestFunctionsCommands(t *testing.T) {
	tempProjectDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health", "/health":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/functions":
			json.NewEncoder(w).Encode([]functions.RemoteFunction{})
		case "/api/v1/functions/deploy":
			json.NewEncoder(w).Encode(functions.DeployResponse{
				Deployed: []string{"hello"},
				Success:  true,
			})
		case "/api/v1/functions/hello":
			w.WriteHeader(http.StatusOK)
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

	fnCmd := NewFunctionsCommand()

	// 1. functions new hello
	out.Reset()
	if err := fnCmd.Execute(ctx, []string{"new", "hello"}); err != nil {
		t.Fatalf("functions new failed: %v", err)
	}
	if !strings.Contains(out.String(), "Created function \"hello\"") {
		t.Errorf("unexpected output from functions new:\n%s", out.String())
	}

	// Verify created files
	fnDir := filepath.Join(tempProjectDir, "functions", "hello")
	if _, err := os.Stat(filepath.Join(fnDir, "index.js")); os.IsNotExist(err) {
		t.Error("expected index.js to exist")
	}
	if _, err := os.Stat(filepath.Join(fnDir, "function.json")); os.IsNotExist(err) {
		t.Error("expected function.json to exist")
	}

	// 2. functions list
	out.Reset()
	if err := fnCmd.Execute(ctx, []string{"list"}); err != nil {
		t.Fatalf("functions list failed: %v", err)
	}
	if !strings.Contains(out.String(), "Local Functions") || !strings.Contains(out.String(), "hello") {
		t.Errorf("unexpected output from functions list:\n%s", out.String())
	}

	// 3. functions deploy
	out.Reset()
	if err := fnCmd.Execute(ctx, []string{"deploy"}); err != nil {
		t.Fatalf("functions deploy failed: %v", err)
	}
	if !strings.Contains(out.String(), "hello deployed") || !strings.Contains(out.String(), "synchronized") {
		t.Errorf("unexpected output from functions deploy:\n%s", out.String())
	}

	// 4. functions delete hello --force
	out.Reset()
	if err := fnCmd.Execute(ctx, []string{"delete", "hello", "--force"}); err != nil {
		t.Fatalf("functions delete failed: %v", err)
	}
	if !strings.Contains(out.String(), "Function \"hello\" successfully deleted") {
		t.Errorf("unexpected output from functions delete:\n%s", out.String())
	}

	// Verify folder deleted
	if _, err := os.Stat(fnDir); !os.IsNotExist(err) {
		t.Error("expected function directory to be deleted")
	}
}
