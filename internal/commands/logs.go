package commands

import (
	"fmt"
	"os"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/functions"
)

type LogsCommand struct {
	fnClient *functions.Client
}

func NewLogsCommand() Command {
	return &LogsCommand{
		fnClient: functions.NewClient(),
	}
}

func (c *LogsCommand) Name() string {
	return "logs"
}

func (c *LogsCommand) Description() string {
	return "Stream runtime logs from deployed services or functions"
}

func (c *LogsCommand) Usage() string {
	return "nodephone logs <service-name>"
}

func (c *LogsCommand) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("service/function name is required. Usage: %s", c.Usage())
	}

	target := args[0]
	creds, _ := auth.LoadCredentials()
	cfg, _ := auth.LoadServerConfig()

	serverURL := "http://localhost:8080"
	if creds != nil && creds.ServerURL != "" {
		serverURL = creds.ServerURL
	} else if cfg != nil && cfg.ServerURL != "" {
		serverURL = cfg.ServerURL
	}

	token := ""
	if creds != nil {
		token = creds.AccessToken
	}

	ctx.Printer.Header(fmt.Sprintf("Streaming logs for %q...", target))
	ctx.Printer.Println()

	return c.fnClient.StreamLogs(serverURL, token, target, os.Stdout)
}
