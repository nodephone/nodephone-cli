package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrinterColorDisabled(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	p := NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	text := p.Green("Hello World")
	if text != "Hello World" {
		t.Errorf("expected plain text when color disabled, got %q", text)
	}

	p.Header("Test Header")
	if !strings.Contains(out.String(), "Test Header") {
		t.Errorf("expected header output, got %q", out.String())
	}
}

func TestPrinterColorEnabled(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	p := NewWithWriters(out, errOut)
	p.SetColorEnabled(true)

	text := p.Green("Hello World")
	if !strings.Contains(text, GreenSeq) || !strings.Contains(text, Reset) {
		t.Errorf("expected ANSI escape sequences in colored text, got %q", text)
	}
}
