package commands

import (
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/output"
)

// Context provides shared runtime context to commands.
type Context struct {
	Config   *config.Config
	Printer  *output.Printer
	Registry *Registry
}

// Command is the interface every CLI command must implement.
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(ctx *Context, args []string) error
}
