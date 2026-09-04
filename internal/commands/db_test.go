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
	"github.com/nodephone/nodephone-cli/internal/db"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestDBCommands(t *testing.T) {
	tempProjectDir := t.TempDir()
	schemaDir := filepath.Join(tempProjectDir, "schema")
	os.MkdirAll(schemaDir, 0755)

	// Create a local migration
	os.WriteFile(filepath.Join(schemaDir, "001_initial.sql"), []byte("CREATE TABLE users (id INT);"), 0644)

	// Mock DB HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health", "/health":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/db/migrations":
			json.NewEncoder(w).Encode([]db.MigrationRecord{})
		case "/api/v1/db/push":
			json.NewEncoder(w).Encode(db.PushResponse{
				Applied: []string{"001_initial"},
				Success: true,
			})
		case "/api/v1/db/pull":
			json.NewEncoder(w).Encode(db.PullResponse{
				Migrations: []db.MigrationFile{
					{Name: "001_initial.sql", Content: "CREATE TABLE users (id INT);"},
					{Name: "002_posts.sql", Content: "CREATE TABLE posts (id INT);"},
				},
			})
		case "/api/v1/db/reset":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Save mock server URL in config
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

	dbCmd := NewDBCommand()

	// 1. db status
	out.Reset()
	if err := dbCmd.Execute(ctx, []string{"status"}); err != nil {
		t.Fatalf("db status failed: %v", err)
	}
	if !strings.Contains(out.String(), "Local Migrations") || !strings.Contains(out.String(), "1 Pending") {
		t.Errorf("unexpected status output:\n%s", out.String())
	}

	// 2. db push
	out.Reset()
	if err := dbCmd.Execute(ctx, []string{"push"}); err != nil {
		t.Fatalf("db push failed: %v", err)
	}
	if !strings.Contains(out.String(), "Applied 001_initial") || !strings.Contains(out.String(), "Database synchronized") {
		t.Errorf("unexpected push output:\n%s", out.String())
	}

	// 3. db pull
	out.Reset()
	if err := dbCmd.Execute(ctx, []string{"pull"}); err != nil {
		t.Fatalf("db pull failed: %v", err)
	}
	if !strings.Contains(out.String(), "Pulled 2 schema migration(s)") || !strings.Contains(out.String(), "Saved 002_posts.sql") {
		t.Errorf("unexpected pull output:\n%s", out.String())
	}

	// Verify pull saved 002_posts.sql
	if _, err := os.Stat(filepath.Join(schemaDir, "002_posts.sql")); os.IsNotExist(err) {
		t.Error("expected 002_posts.sql to be created by db pull")
	}

	// 4. db reset with --force
	out.Reset()
	if err := dbCmd.Execute(ctx, []string{"reset", "--force"}); err != nil {
		t.Fatalf("db reset failed: %v", err)
	}
	if !strings.Contains(out.String(), "Reset development database") || !strings.Contains(out.String(), "Database reset complete") {
		t.Errorf("unexpected reset output:\n%s", out.String())
	}
}
