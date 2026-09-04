package output

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Standard ANSI Escape Codes
const (
	Reset     = "\033[0m"
	BoldSeq   = "\033[1m"
	DimSeq    = "\033[2m"
	RedSeq    = "\033[31m"
	GreenSeq  = "\033[32m"
	YellowSeq = "\033[33m"
	BlueSeq   = "\033[34m"
	MagentaSeq= "\033[35m"
	CyanSeq   = "\033[36m"
	GraySeq   = "\033[90m"
)

// Printer wraps terminal output streams with styling rules.
type Printer struct {
	out        io.Writer
	errOut     io.Writer
	colorEnabled bool
}

// New creates a new Printer instance writing to standard outputs.
func New() *Printer {
	return NewWithWriters(os.Stdout, os.Stderr)
}

// NewWithWriters creates a Printer with custom writers (useful for testing or redirection).
func NewWithWriters(out, errOut io.Writer) *Printer {
	// Respect NO_COLOR env standard (https://no-color.org)
	noColor := os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"
	
	return &Printer{
		out:          out,
		errOut:       errOut,
		colorEnabled: !noColor,
	}
}

// SetColorEnabled manually overrides color settings.
func (p *Printer) SetColorEnabled(enabled bool) {
	p.colorEnabled = enabled
}

// IsColorEnabled returns whether color output is active.
func (p *Printer) IsColorEnabled() bool {
	return p.colorEnabled
}

// Colorize wraps text in ANSI escape sequence if colors are enabled.
func (p *Printer) Colorize(seq, text string) string {
	if !p.colorEnabled || seq == "" {
		return text
	}
	return seq + text + Reset
}

// Styling helpers
func (p *Printer) Bold(text string) string   { return p.Colorize(BoldSeq, text) }
func (p *Printer) Dim(text string) string    { return p.Colorize(DimSeq, text) }
func (p *Printer) Red(text string) string    { return p.Colorize(RedSeq, text) }
func (p *Printer) Green(text string) string  { return p.Colorize(GreenSeq, text) }
func (p *Printer) Yellow(text string) string { return p.Colorize(YellowSeq, text) }
func (p *Printer) Cyan(text string) string   { return p.Colorize(CyanSeq, text) }
func (p *Printer) Gray(text string) string   { return p.Colorize(GraySeq, text) }

// Println writes to standard output with newline.
func (p *Printer) Println(a ...any) {
	fmt.Fprintln(p.out, a...)
}

// Printf writes formatted output to standard output.
func (p *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(p.out, format, a...)
}

// Errorln writes to error output with newline.
func (p *Printer) Errorln(a ...any) {
	fmt.Fprintln(p.errOut, a...)
}

// Errorf writes formatted output to error output.
func (p *Printer) Errorf(format string, a ...any) {
	fmt.Fprintf(p.errOut, format, a...)
}

// Header prints a prominent bold title.
func (p *Printer) Header(title string) {
	p.Println(p.Bold(title))
}

// Section prints a section header.
func (p *Printer) Section(title string) {
	p.Println(p.Bold(title))
}

// Success prints a green success message.
func (p *Printer) Success(msg string) {
	p.Println(p.Green("✔ " + msg))
}

// ErrorMsg prints a red error message.
func (p *Printer) ErrorMsg(msg string) {
	p.Errorln(p.Red("✖ " + msg))
}

// Warn prints a yellow warning message.
func (p *Printer) Warn(msg string) {
	p.Println(p.Yellow("⚠ " + msg))
}

// Info prints a cyan info message.
func (p *Printer) Info(msg string) {
	p.Println(p.Cyan("ℹ " + msg))
}

// TwoColumn prints key-value formatted output with aligned padding.
func (p *Printer) TwoColumn(key string, keyWidth int, val string) {
	padding := ""
	if len(key) < keyWidth {
		padding = strings.Repeat(" ", keyWidth-len(key))
	}
	p.Printf("%s%s : %s\n", key, padding, val)
}
