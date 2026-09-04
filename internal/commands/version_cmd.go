package commands

import (
	"github.com/nodephone/nodephone-cli/internal/version"
)

type VersionCommand struct{}

func NewVersionCommand() Command {
	return &VersionCommand{}
}

func (c *VersionCommand) Name() string {
	return "version"
}

func (c *VersionCommand) Description() string {
	return "Display NodePhone CLI version information"
}

func (c *VersionCommand) Usage() string {
	return "nodephone version"
}

func (c *VersionCommand) Execute(ctx *Context, args []string) error {
	info := version.GetInfo()
	ctx.Printer.Println(info.String())
	return nil
}
