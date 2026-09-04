package commands

import (
	"fmt"
	"strings"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/diagnostics"
)

type LogsCommand struct {
	diagClient *diagnostics.Client
}

func NewLogsCommand() Command {
	return &LogsCommand{
		diagClient: diagnostics.NewClient(),
	}
}

func (c *LogsCommand) Name() string {
	return "logs"
}

func (c *LogsCommand) Description() string {
	return "Stream live runtime logs from NodePhone Server or deployed functions"
}

func (c *LogsCommand) Usage() string {
	return "nodephone logs [service|function-name] [--follow|-f] [--json] [--level <INFO|DB|WS|FUNC|ERROR>]"
}

func (c *LogsCommand) Execute(ctx *Context, args []string) error {
	follow := false
	jsonMode := false
	levelFilter := ""
	target := ""

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--follow" || arg == "-f":
			follow = true
		case arg == "--json":
			jsonMode = true
		case (arg == "--level" || arg == "-l") && i+1 < len(args):
			levelFilter = args[i+1]
			i++
		case !strings.HasPrefix(arg, "-") && target == "":
			target = arg
		}
	}

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

	if target != "" && levelFilter == "" {
		// If a specific function or service target was passed, set level filter or header
		ctx.Printer.Header(fmt.Sprintf("Streaming logs for %q...", target))
	} else {
		ctx.Printer.Header("NodePhone Live Server Logs")
	}
	ctx.Printer.Println()

	return c.diagClient.StreamServerLogs(serverURL, token, follow, levelFilter, jsonMode, ctx.Printer)
}
