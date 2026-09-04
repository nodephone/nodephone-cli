package commands

import (
	"fmt"
	"github.com/nodephone/nodephone-cli/internal/version"
)

type HelpCommand struct{}

func NewHelpCommand() Command {
	return &HelpCommand{}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Display help and available commands"
}

func (c *HelpCommand) Usage() string {
	return "nodephone help [command]"
}

func (c *HelpCommand) Execute(ctx *Context, args []string) error {
	if len(args) > 0 {
		cmdName := args[0]
		if cmd, exists := ctx.Registry.Get(cmdName); exists {
			ctx.Printer.Header(fmt.Sprintf("Command: %s", cmd.Name()))
			ctx.Printer.Println()
			ctx.Printer.Println("Description : " + cmd.Description())
			ctx.Printer.Println("Usage       : " + cmd.Usage())
			return nil
		}
		return fmt.Errorf("unknown command %q for help", cmdName)
	}

	info := version.GetInfo()

	ctx.Printer.Header("NodePhone CLI")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Version", 7, info.Version)
	ctx.Printer.TwoColumn("Go", 7, info.GoVersion)
	ctx.Printer.TwoColumn("OS", 7, info.OS)
	ctx.Printer.Println()
	ctx.Printer.Header("Available Commands")
	ctx.Printer.Println()

	commands := ctx.Registry.List()
	for _, cmd := range commands {
		// Output command name cleanly
		ctx.Printer.Println(cmd.Name())
	}

	return nil
}
