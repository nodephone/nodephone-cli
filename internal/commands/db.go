package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/db"
)

type DBCommand struct {
	dbClient   *db.Client
	authClient *auth.Client
}

func NewDBCommand() Command {
	return &DBCommand{
		dbClient:   db.NewClient(),
		authClient: auth.NewClient(),
	}
}

func (c *DBCommand) Name() string {
	return "db"
}

func (c *DBCommand) Description() string {
	return "Manage database migrations and schema synchronization"
}

func (c *DBCommand) Usage() string {
	return "nodephone db <push|pull|status|reset> [--force]"
}

func (c *DBCommand) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp(ctx)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "push":
		return c.handlePush(ctx, subArgs)
	case "pull":
		return c.handlePull(ctx, subArgs)
	case "status":
		return c.handleStatus(ctx, subArgs)
	case "reset":
		return c.handleReset(ctx, subArgs)
	case "help", "--help", "-h":
		return c.showHelp(ctx)
	default:
		return fmt.Errorf("unknown db subcommand %q. Run 'nodephone db help' for usage", subcommand)
	}
}

func (c *DBCommand) showHelp(ctx *Context) error {
	ctx.Printer.Header("NodePhone Database Commands")
	ctx.Printer.Println()
	ctx.Printer.Println("Usage: " + c.Usage())
	ctx.Printer.Println()
	ctx.Printer.Println("Subcommands:")
	ctx.Printer.TwoColumn("push", 8, "Apply pending local SQL migrations to server")
	ctx.Printer.TwoColumn("pull", 8, "Download remote database schema into local schema/ directory")
	ctx.Printer.TwoColumn("status", 8, "Display local vs remote migration comparison")
	ctx.Printer.TwoColumn("reset", 8, "Reset development database and re-apply all migrations")
	return nil
}

func (c *DBCommand) getAuthContext() (string, string, error) {
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

	return serverURL, token, nil
}

func (c *DBCommand) handlePush(ctx *Context, args []string) error {
	serverURL, token, err := c.getAuthContext()
	if err != nil {
		return err
	}

	if err := c.authClient.PingServer(serverURL); err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", serverURL, err)
	}
	ctx.Printer.Success("Connected")

	schemaDir := filepath.Join(ctx.Config.ProjectRoot, "schema")
	localMigrations, err := db.ReadLocalMigrations(schemaDir)
	if err != nil {
		return fmt.Errorf("failed to read local migrations: %w", err)
	}

	remoteRecords, err := c.dbClient.GetStatus(serverURL, token)
	if err != nil {
		return fmt.Errorf("failed to get remote migration status: %w", err)
	}

	remoteAppliedMap := make(map[string]string)
	for _, rec := range remoteRecords {
		remoteAppliedMap[rec.Name] = rec.Checksum
	}

	var pending []db.MigrationFile
	for _, loc := range localMigrations {
		remoteChecksum, applied := remoteAppliedMap[loc.Name]
		if !applied {
			pending = append(pending, loc)
		} else if remoteChecksum != "" && remoteChecksum != loc.Checksum {
			ctx.Printer.Warn(fmt.Sprintf("Migration %s checksum mismatch! Local copy has been modified.", loc.Name))
		}
	}

	if len(pending) == 0 {
		ctx.Printer.Info("Database is up to date (0 pending migrations).")
		return nil
	}

	ctx.Printer.Success(fmt.Sprintf("%d pending migration(s)", len(pending)))

	appliedNames, err := c.dbClient.PushMigrations(serverURL, token, pending)
	if err != nil {
		return err
	}

	for _, name := range appliedNames {
		ctx.Printer.Success(fmt.Sprintf("Applied %s", name))
	}

	ctx.Printer.Println()
	ctx.Printer.Header("Database synchronized.")
	return nil
}

func (c *DBCommand) handlePull(ctx *Context, args []string) error {
	serverURL, token, err := c.getAuthContext()
	if err != nil {
		return err
	}

	if err := c.authClient.PingServer(serverURL); err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", serverURL, err)
	}
	ctx.Printer.Success("Connected")

	schemaDir := filepath.Join(ctx.Config.ProjectRoot, "schema")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		return fmt.Errorf("failed to create schema directory: %w", err)
	}

	remoteMigrations, err := c.dbClient.PullSchema(serverURL, token)
	if err != nil {
		return err
	}

	if len(remoteMigrations) == 0 {
		ctx.Printer.Info("No remote schema migrations found.")
		return nil
	}

	ctx.Printer.Success(fmt.Sprintf("Pulled %d schema migration(s) into schema/", len(remoteMigrations)))

	for _, m := range remoteMigrations {
		fileName := m.Name
		if !strings.HasSuffix(fileName, ".sql") {
			fileName = fileName + ".sql"
		}
		filePath := filepath.Join(schemaDir, fileName)
		if err := os.WriteFile(filePath, []byte(m.Content), 0644); err != nil {
			return fmt.Errorf("failed to save migration %s: %w", fileName, err)
		}
		ctx.Printer.Success(fmt.Sprintf("Saved %s", fileName))
	}

	ctx.Printer.Println()
	ctx.Printer.Header("Schema pull complete.")
	return nil
}

func (c *DBCommand) handleStatus(ctx *Context, args []string) error {
	serverURL, token, err := c.getAuthContext()
	if err != nil {
		return err
	}

	schemaDir := filepath.Join(ctx.Config.ProjectRoot, "schema")
	localMigrations, err := db.ReadLocalMigrations(schemaDir)
	if err != nil {
		return fmt.Errorf("failed to read local migrations: %w", err)
	}

	remoteRecords, err := c.dbClient.GetStatus(serverURL, token)
	if err != nil {
		return fmt.Errorf("failed to fetch remote migration status: %w", err)
	}

	localCount := len(localMigrations)
	remoteCount := len(remoteRecords)

	remoteAppliedMap := make(map[string]bool)
	for _, rec := range remoteRecords {
		remoteAppliedMap[rec.Name] = true
	}

	pendingCount := 0
	for _, loc := range localMigrations {
		if !remoteAppliedMap[loc.Name] {
			pendingCount++
		}
	}

	statusText := "Synced"
	if pendingCount > 0 {
		statusText = fmt.Sprintf("%d Pending Migration(s)", pendingCount)
	}

	ctx.Printer.Header("NodePhone Database Status")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Local Migrations", 18, fmt.Sprintf("%d", localCount))
	ctx.Printer.TwoColumn("Remote Migrations", 18, fmt.Sprintf("%d", remoteCount))
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Status", 18, statusText)
	ctx.Printer.Println()

	return nil
}

func (c *DBCommand) handleReset(ctx *Context, args []string) error {
	force := false
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		}
	}

	if !force {
		ctx.Printer.Warn("This will drop and rebuild the development database, destroying all local data.")
		ctx.Printer.Printf("Are you sure you want to proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			ctx.Printer.Info("Database reset cancelled.")
			return nil
		}
	}

	serverURL, token, err := c.getAuthContext()
	if err != nil {
		return err
	}

	if err := c.authClient.PingServer(serverURL); err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", serverURL, err)
	}
	ctx.Printer.Success("Connected")

	if err := c.dbClient.ResetDatabase(serverURL, token); err != nil {
		return fmt.Errorf("database reset failed: %w", err)
	}
	ctx.Printer.Success("Reset development database")

	// Re-apply local migrations after reset
	schemaDir := filepath.Join(ctx.Config.ProjectRoot, "schema")
	localMigrations, _ := db.ReadLocalMigrations(schemaDir)

	if len(localMigrations) > 0 {
		appliedNames, err := c.dbClient.PushMigrations(serverURL, token, localMigrations)
		if err != nil {
			return fmt.Errorf("failed to re-apply local migrations after reset: %w", err)
		}
		ctx.Printer.Success(fmt.Sprintf("Re-applied %d local migration(s)", len(appliedNames)))
	}

	ctx.Printer.Println()
	ctx.Printer.Header("Database reset complete.")
	return nil
}
