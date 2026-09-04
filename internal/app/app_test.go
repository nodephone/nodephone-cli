package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestAppNoArgs(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	application, err := NewWithWriters(out, errOut)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	exitCode := application.Run([]string{})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for no args (default help), got %d", exitCode)
	}

	if !strings.Contains(out.String(), "NodePhone CLI") {
		t.Errorf("expected stdout to contain header, got:\n%s", out.String())
	}
}

func TestAppVersionFlag(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	application, err := NewWithWriters(out, errOut)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	exitCode := application.Run([]string{"--version"})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for --version, got %d", exitCode)
	}

	if !strings.Contains(out.String(), "Version : v1.0.0") {
		t.Errorf("expected stdout to contain version, got:\n%s", out.String())
	}
}

func TestAppUnknownCommand(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	application, err := NewWithWriters(out, errOut)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	exitCode := application.Run([]string{"nonexistent"})
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for unknown command, got %d", exitCode)
	}

	if !strings.Contains(errOut.String(), "unknown command") {
		t.Errorf("expected stderr to contain 'unknown command', got:\n%s", errOut.String())
	}
}
