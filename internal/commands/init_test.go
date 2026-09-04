package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestValidateProjectName(t *testing.T) {
	validNames := []string{"my-app", "my_app", "app123", "DemoApp"}
	for _, name := range validNames {
		if err := ValidateProjectName(name); err != nil {
			t.Errorf("expected valid project name %q, got error: %v", name, err)
		}
	}

	invalidNames := []string{"", "my app", "my/app", "..", "app*", "app?", "app>"}
	for _, name := range invalidNames {
		if err := ValidateProjectName(name); err == nil {
			t.Errorf("expected error for invalid project name %q, but got nil", name)
		}
	}
}

func TestInitCommandScaffold(t *testing.T) {
	tempDir := t.TempDir()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	ctx := &Context{
		Config: &config.Config{
			ProjectRoot: tempDir,
		},
		Printer:  p,
		Registry: NewRegistry(),
	}

	initCmd := NewInitCommand()
	err := initCmd.Execute(ctx, []string{"test-app"})
	if err != nil {
		t.Fatalf("unexpected error executing init command: %v", err)
	}

	appDir := filepath.Join(tempDir, "test-app")

	expectedFiles := []string{
		"nodephone.json",
		".env.example",
		filepath.Join("schema", "001_initial.sql"),
		filepath.Join("functions", "hello", "index.js"),
		filepath.Join("storage", "public", ".gitkeep"),
		filepath.Join("storage", "private", ".gitkeep"),
		"README.md",
	}

	for _, relPath := range expectedFiles {
		fullPath := filepath.Join(appDir, relPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist, but it was not created", relPath)
		}
	}

	// Verify nodephone.json content
	cfgBytes, err := os.ReadFile(filepath.Join(appDir, "nodephone.json"))
	if err != nil {
		t.Fatalf("failed to read nodephone.json: %v", err)
	}
	if !strings.Contains(string(cfgBytes), `"name": "test-app"`) {
		t.Errorf("nodephone.json does not contain project name: %s", string(cfgBytes))
	}

	// Verify SQL content
	sqlBytes, err := os.ReadFile(filepath.Join(appDir, "schema", "001_initial.sql"))
	if err != nil {
		t.Fatalf("failed to read 001_initial.sql: %v", err)
	}
	if !strings.Contains(string(sqlBytes), "CREATE TABLE IF NOT EXISTS users") {
		t.Errorf("001_initial.sql missing starter schema: %s", string(sqlBytes))
	}

	// Verify function content
	jsBytes, err := os.ReadFile(filepath.Join(appDir, "functions", "hello", "index.js"))
	if err != nil {
		t.Fatalf("failed to read index.js: %v", err)
	}
	if !strings.Contains(string(jsBytes), "module.exports = async function handler") {
		t.Errorf("index.js missing handler export: %s", string(jsBytes))
	}
}

func TestInitCommandOverwriteProtection(t *testing.T) {
	tempDir := t.TempDir()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	ctx := &Context{
		Config: &config.Config{
			ProjectRoot: tempDir,
		},
		Printer:  p,
		Registry: NewRegistry(),
	}

	initCmd := NewInitCommand()

	// Initial run
	if err := initCmd.Execute(ctx, []string{"existing-app"}); err != nil {
		t.Fatalf("initial run failed: %v", err)
	}

	// Second run without --force should fail
	err := initCmd.Execute(ctx, []string{"existing-app"})
	if err == nil {
		t.Error("expected error when initializing over existing non-empty directory without --force, got nil")
	}

	// Second run with --force should succeed
	err = initCmd.Execute(ctx, []string{"existing-app", "--force"})
	if err != nil {
		t.Errorf("expected success with --force flag, got: %v", err)
	}
}
