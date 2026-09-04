package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func newTestContext() (*Context, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	cfg := &config.Config{
		Environment: "test",
	}

	reg := NewRegistry()

	ctx := &Context{
		Config:   cfg,
		Printer:  p,
		Registry: reg,
	}

	return ctx, out, errOut
}

func TestRegistry(t *testing.T) {
	ctx, out, _ := newTestContext()

	helpCmd := NewHelpCommand()
	versionCmd := NewVersionCommand()

	ctx.Registry.Register(helpCmd)
	ctx.Registry.Register(versionCmd)

	if len(ctx.Registry.List()) != 2 {
		t.Errorf("expected 2 registered commands, got %d", len(ctx.Registry.List()))
	}

	err := ctx.Registry.Execute(ctx, "version", nil)
	if err != nil {
		t.Fatalf("unexpected error executing version command: %v", err)
	}

	if !strings.Contains(out.String(), "NodePhone CLI") {
		t.Errorf("expected version output in stdout, got:\n%s", out.String())
	}
}

func TestHelpCommandOutput(t *testing.T) {
	ctx, out, _ := newTestContext()

	ctx.Registry.Register(NewInitCommand())
	RegisterPlaceholders(ctx.Registry)

	helpCmd := NewHelpCommand()
	ctx.Registry.Register(helpCmd)

	err := helpCmd.Execute(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected help command error: %v", err)
	}

	outputStr := out.String()
	expectedSubstrings := []string{
		"NodePhone CLI",
		"Version",
		"Available Commands",
		"init",
		"login",
		"db",
		"gen",
		"logs",
		"deploy",
		"inspect",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(outputStr, sub) {
			t.Errorf("help output missing expected substring %q. Output was:\n%s", sub, outputStr)
		}
	}
}

func TestPlaceholderExecution(t *testing.T) {
	ctx, out, _ := newTestContext()
	RegisterPlaceholders(ctx.Registry)

	err := ctx.Registry.Execute(ctx, "login", nil)
	if err != nil {
		t.Fatalf("unexpected error executing login placeholder: %v", err)
	}

	if !strings.Contains(out.String(), "login") {
		t.Errorf("expected placeholder text for login command, got:\n%s", out.String())
	}
}
