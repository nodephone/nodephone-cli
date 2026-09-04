package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nodephone/nodephone-cli/internal/auth"
)

type LoginCommand struct {
	authClient *auth.Client
}

func NewLoginCommand() Command {
	return &LoginCommand{
		authClient: auth.NewClient(),
	}
}

func (c *LoginCommand) Name() string {
	return "login"
}

func (c *LoginCommand) Description() string {
	return "Authenticate with NodePhone server"
}

func (c *LoginCommand) Usage() string {
	return "nodephone login [--server <url>] [--email <email>] [--password <pass>]"
}

func (c *LoginCommand) Execute(ctx *Context, args []string) error {
	var serverURL, email, password string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case (arg == "--server" || arg == "-s") && i+1 < len(args):
			serverURL = args[i+1]
			i++
		case (arg == "--email" || arg == "-e") && i+1 < len(args):
			email = args[i+1]
			i++
		case (arg == "--password" || arg == "-p") && i+1 < len(args):
			password = args[i+1]
			i++
		}
	}

	reader := bufio.NewReader(os.Stdin)

	if serverURL == "" {
		cfg, _ := auth.LoadServerConfig()
		defaultServer := "http://localhost:8080"
		if cfg != nil && cfg.ServerURL != "" {
			defaultServer = cfg.ServerURL
		}

		ctx.Printer.Printf("Server URL [%s]: ", defaultServer)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			serverURL = defaultServer
		} else {
			serverURL = input
		}
	}

	serverURL = auth.EnsureURLHasScheme(serverURL)

	if email == "" {
		ctx.Printer.Printf("Email: ")
		input, _ := reader.ReadString('\n')
		email = strings.TrimSpace(input)
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if password == "" {
		ctx.Printer.Printf("Password: ")
		input, _ := reader.ReadString('\n')
		password = strings.TrimSpace(input)
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	ctx.Printer.Println()
	ctx.Printer.Info(fmt.Sprintf("Connecting to %s...", serverURL))

	creds, err := c.authClient.Login(serverURL, email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := auth.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	if err := auth.SaveServerConfig(serverURL, ""); err != nil {
		return fmt.Errorf("failed to save server config: %w", err)
	}

	ctx.Printer.Success(fmt.Sprintf("Server %s is reachable", serverURL))
	ctx.Printer.Success(fmt.Sprintf("Authenticated successfully as %s", email))
	ctx.Printer.Println()
	ctx.Printer.Header("Connected User")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Email", 7, creds.Email)
	ctx.Printer.TwoColumn("Server", 7, creds.ServerURL)
	ctx.Printer.TwoColumn("Status", 7, "Connected")
	ctx.Printer.Println()

	return nil
}
