package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/diagnostics"
)

type InspectCommand struct {
	diagClient *diagnostics.Client
	authClient *auth.Client
}

func NewInspectCommand() Command {
	return &InspectCommand{
		diagClient: diagnostics.NewClient(),
		authClient: auth.NewClient(),
	}
}

func (c *InspectCommand) Name() string {
	return "inspect"
}

func (c *InspectCommand) Description() string {
	return "Inspect NodePhone Server health, realtime, and storage analytics"
}

func (c *InspectCommand) Usage() string {
	return "nodephone inspect [realtime|storage] [--json]"
}

func (c *InspectCommand) Execute(ctx *Context, args []string) error {
	jsonMode := false
	subcommand := ""

	for _, arg := range args {
		if arg == "--json" {
			jsonMode = true
		} else if !strings.HasPrefix(arg, "-") && subcommand == "" {
			subcommand = arg
		}
	}

	serverURL, token := c.getAuthContext()

	switch subcommand {
	case "realtime":
		return c.handleRealtime(ctx, serverURL, token, jsonMode)
	case "storage":
		return c.handleStorage(ctx, serverURL, token, jsonMode)
	case "", "status":
		return c.handleDefault(ctx, serverURL, token, jsonMode)
	default:
		return fmt.Errorf("unknown inspect subcommand %q. Usage: %s", subcommand, c.Usage())
	}
}

func (c *InspectCommand) getAuthContext() (string, string) {
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

	return serverURL, token
}

func (c *InspectCommand) handleDefault(ctx *Context, serverURL, token string, jsonMode bool) error {
	status, err := c.diagClient.GetInspectStatus(serverURL, token)
	if err != nil {
		return err
	}

	if jsonMode {
		bytes, _ := json.MarshalIndent(status, "", "  ")
		ctx.Printer.Println(string(bytes))
		return nil
	}

	ctx.Printer.Header("NodePhone Server")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Version", 12, status.Version)
	ctx.Printer.TwoColumn("Status", 12, status.Status)
	ctx.Printer.TwoColumn("Uptime", 12, status.Uptime)
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("CPU", 12, status.CPU)
	ctx.Printer.TwoColumn("Memory", 12, status.Memory)
	ctx.Printer.TwoColumn("Storage", 12, status.Storage)
	ctx.Printer.TwoColumn("Database", 12, status.Database)
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Users", 12, fmt.Sprintf("%d", status.Users))
	ctx.Printer.TwoColumn("Realtime", 12, status.Realtime)
	ctx.Printer.TwoColumn("Functions", 12, fmt.Sprintf("%d", status.Functions))
	ctx.Printer.Println()

	return nil
}

func (c *InspectCommand) handleRealtime(ctx *Context, serverURL, token string, jsonMode bool) error {
	rt, err := c.diagClient.GetRealtimeMetrics(serverURL, token)
	if err != nil {
		return err
	}

	if jsonMode {
		bytes, _ := json.MarshalIndent(rt, "", "  ")
		ctx.Printer.Println(string(bytes))
		return nil
	}

	ctx.Printer.Header("NodePhone Realtime Analytics")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Connected Users", 18, fmt.Sprintf("%d", rt.ConnectedUsers))
	ctx.Printer.TwoColumn("Active Rooms", 18, fmt.Sprintf("%d", rt.ActiveRooms))
	ctx.Printer.TwoColumn("Presence Count", 18, fmt.Sprintf("%d", rt.PresenceCount))
	ctx.Printer.TwoColumn("Messages Rate", 18, rt.MessagesPerSec)
	ctx.Printer.Println()

	return nil
}

func (c *InspectCommand) handleStorage(ctx *Context, serverURL, token string, jsonMode bool) error {
	st, err := c.diagClient.GetStorageAnalytics(serverURL, token)
	if err != nil {
		return err
	}

	if jsonMode {
		bytes, _ := json.MarshalIndent(st, "", "  ")
		ctx.Printer.Println(string(bytes))
		return nil
	}

	ctx.Printer.Header("NodePhone Storage Analytics")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Public Bucket", 18, st.PublicBucketSize)
	ctx.Printer.TwoColumn("Private Bucket", 18, st.PrivateBucketSize)
	ctx.Printer.TwoColumn("Total Files", 18, fmt.Sprintf("%d", st.TotalFiles))
	ctx.Printer.Println()

	if len(st.LargestFiles) > 0 {
		ctx.Printer.Header("Largest Files:")
		for _, f := range st.LargestFiles {
			ctx.Printer.Println(fmt.Sprintf("  - %s (%s) [%s]", f.Name, f.Size, f.Bucket))
		}
		ctx.Printer.Println()
	}

	if len(st.RecentUploads) > 0 {
		ctx.Printer.Header("Recent Uploads:")
		for _, f := range st.RecentUploads {
			ctx.Printer.Println(fmt.Sprintf("  - %s (%s) [%s]", f.Name, f.Size, f.Bucket))
		}
		ctx.Printer.Println()
	}

	return nil
}
