package app

import (
	"fmt"
	"io"

	"github.com/nodephone/nodephone-cli/internal/commands"
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/output"
)

// App is the CLI application kernel container.
type App struct {
	Config   *config.Config
	Printer  *output.Printer
	Registry *commands.Registry
}

// New creates and initializes a default App instance.
func New() (*App, error) {
	return NewWithWriters(nil, nil)
}

// NewWithWriters creates an App instance with optional custom output writers.
func NewWithWriters(out, errOut io.Writer) (*App, error) {
	var p *output.Printer
	if out != nil && errOut != nil {
		p = output.NewWithWriters(out, errOut)
	} else {
		p = output.New()
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	reg := commands.NewRegistry()

	// Register Core Commands
	initCmd := commands.NewInitCommand()
	loginCmd := commands.NewLoginCommand()
	logoutCmd := commands.NewLogoutCommand()
	whoamiCmd := commands.NewWhoamiCommand()
	dbCmd := commands.NewDBCommand()
	genCmd := commands.NewGenCommand()
	fnCmd := commands.NewFunctionsCommand()
	logsCmd := commands.NewLogsCommand()
	helpCmd := commands.NewHelpCommand()
	versionCmd := commands.NewVersionCommand()

	// Register Core Builtins & Placeholders
	reg.Register(initCmd)
	reg.Register(loginCmd)
	reg.Register(logoutCmd)
	reg.Register(whoamiCmd)
	reg.Register(dbCmd)
	reg.Register(genCmd)
	reg.Register(fnCmd)
	reg.Register(logsCmd)
	commands.RegisterPlaceholders(reg)
	reg.Register(helpCmd)
	reg.Register(versionCmd)

	app := &App{
		Config:   cfg,
		Printer:  p,
		Registry: reg,
	}

	return app, nil
}

// Run parses command line arguments and dispatches command execution.
// Returns an exit code (0 for success, 1 for failure).
func (a *App) Run(args []string) int {
	ctx := &commands.Context{
		Config:   a.Config,
		Printer:  a.Printer,
		Registry: a.Registry,
	}

	// Case 1: No arguments passed or explicit help flag/command
	if len(args) == 0 {
		return a.execCommand(ctx, "help", nil)
	}

	firstArg := args[0]

	switch firstArg {
	case "--help", "-h":
		return a.execCommand(ctx, "help", args[1:])
	case "version", "--version", "-v":
		return a.execCommand(ctx, "version", args[1:])
	case "deploy":
		return a.execCommand(ctx, "functions", append([]string{"deploy"}, args[1:]...))
	default:
		return a.execCommand(ctx, firstArg, args[1:])
	}
}

func (a *App) execCommand(ctx *commands.Context, name string, args []string) int {
	err := a.Registry.Execute(ctx, name, args)
	if err != nil {
		a.Printer.ErrorMsg(err.Error())
		return 1
	}
	return 0
}
