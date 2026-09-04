package commands

import (
	"github.com/nodephone/nodephone-cli/internal/auth"
)

type WhoamiCommand struct {
	authClient *auth.Client
}

func NewWhoamiCommand() Command {
	return &WhoamiCommand{
		authClient: auth.NewClient(),
	}
}

func (c *WhoamiCommand) Name() string {
	return "whoami"
}

func (c *WhoamiCommand) Description() string {
	return "Display current connected user account information"
}

func (c *WhoamiCommand) Usage() string {
	return "nodephone whoami"
}

func (c *WhoamiCommand) Execute(ctx *Context, args []string) error {
	creds, err := auth.LoadCredentials()
	if err != nil || creds == nil || creds.Email == "" {
		cfg, _ := auth.LoadServerConfig()
		serverURL := "http://localhost:8080"
		if cfg != nil && cfg.ServerURL != "" {
			serverURL = cfg.ServerURL
		}

		ctx.Printer.Header("NodePhone CLI")
		ctx.Printer.Println()
		ctx.Printer.TwoColumn("Server", 7, serverURL)
		ctx.Printer.TwoColumn("Status", 7, "Not authenticated")
		ctx.Printer.Println()
		ctx.Printer.Info("Run 'nodephone login' to connect to your NodePhone server.")
		return nil
	}

	// Auto-refresh token if expired
	if creds.IsExpired() {
		refreshed, err := c.authClient.RefreshToken(creds)
		if err == nil && refreshed != nil {
			creds = refreshed
			_ = auth.SaveCredentials(creds)
		}
	}

	ctx.Printer.Header("NodePhone CLI")
	ctx.Printer.Println()
	ctx.Printer.Header("Connected User")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Email", 7, creds.Email)
	ctx.Printer.TwoColumn("Server", 7, creds.ServerURL)
	ctx.Printer.TwoColumn("Status", 7, "Connected")
	ctx.Printer.Println()

	return nil
}
