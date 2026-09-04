package commands

import (
	"github.com/nodephone/nodephone-cli/internal/auth"
)

type LogoutCommand struct{}

func NewLogoutCommand() Command {
	return &LogoutCommand{}
}

func (c *LogoutCommand) Name() string {
	return "logout"
}

func (c *LogoutCommand) Description() string {
	return "Log out and remove local credentials"
}

func (c *LogoutCommand) Usage() string {
	return "nodephone logout"
}

func (c *LogoutCommand) Execute(ctx *Context, args []string) error {
	creds, _ := auth.LoadCredentials()
	serverURL := "localhost:8080"
	if creds != nil && creds.ServerURL != "" {
		serverURL = creds.ServerURL
	}

	if err := auth.ClearCredentials(); err != nil {
		return err
	}

	ctx.Printer.Success("Successfully logged out.")
	ctx.Printer.Println()
	ctx.Printer.Header("NodePhone CLI")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Server", 7, serverURL)
	ctx.Printer.TwoColumn("Status", 7, "Not authenticated")
	ctx.Printer.Println()

	return nil
}
