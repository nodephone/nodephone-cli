package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/generator"
)

type GenCommand struct {
	genClient  *generator.Client
	authClient *auth.Client
}

func NewGenCommand() Command {
	return &GenCommand{
		genClient:  generator.NewClient(),
		authClient: auth.NewClient(),
	}
}

func (c *GenCommand) Name() string {
	return "gen"
}

func (c *GenCommand) Description() string {
	return "Generate client models and TypeScript definitions from server OpenAPI spec"
}

func (c *GenCommand) Usage() string {
	return "nodephone gen <types|api> [--server <url>]"
}

func (c *GenCommand) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp(ctx)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "types", "api":
		return c.handleGen(ctx, subArgs)
	case "help", "--help", "-h":
		return c.showHelp(ctx)
	default:
		return fmt.Errorf("unknown gen subcommand %q. Run 'nodephone gen help' for usage", subcommand)
	}
}

func (c *GenCommand) showHelp(ctx *Context) error {
	ctx.Printer.Header("NodePhone Code & Type Generator")
	ctx.Printer.Println()
	ctx.Printer.Println("Usage: " + c.Usage())
	ctx.Printer.Println()
	ctx.Printer.Println("Subcommands:")
	ctx.Printer.TwoColumn("types", 8, "Generate TypeScript models and interfaces from OpenAPI spec")
	ctx.Printer.TwoColumn("api", 8, "Generate TypeScript API routes and client definitions")
	return nil
}

func (c *GenCommand) handleGen(ctx *Context, args []string) error {
	var serverURL string
	for i := 0; i < len(args); i++ {
		if (args[i] == "--server" || args[i] == "-s") && i+1 < len(args) {
			serverURL = args[i+1]
			break
		}
	}

	if serverURL == "" {
		creds, _ := auth.LoadCredentials()
		cfg, _ := auth.LoadServerConfig()
		if creds != nil && creds.ServerURL != "" {
			serverURL = creds.ServerURL
		} else if cfg != nil && cfg.ServerURL != "" {
			serverURL = cfg.ServerURL
		} else {
			serverURL = "http://localhost:8080"
		}
	}

	serverURL = auth.EnsureURLHasScheme(serverURL)

	ctx.Printer.Header("NodePhone Type Generator")
	ctx.Printer.Println()

	if err := c.authClient.PingServer(serverURL); err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", serverURL, err)
	}
	ctx.Printer.Success(fmt.Sprintf("Connected to %s", serverURL))

	specBytes, err := c.genClient.FetchOpenAPISpec(serverURL)
	if err != nil {
		return err
	}
	ctx.Printer.Success("Downloaded OpenAPI 3.1 specification")

	spec, err := generator.ParseOpenAPI(specBytes)
	if err != nil {
		return err
	}

	files, err := generator.GenerateTypeScript(spec)
	if err != nil {
		return fmt.Errorf("failed to generate TypeScript code: %w", err)
	}

	typesDir := filepath.Join(ctx.Config.ProjectRoot, "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		return fmt.Errorf("failed to create types directory: %w", err)
	}

	fileOrder := []string{"auth.ts", "database.ts", "storage.ts", "functions.ts", "api.ts", "index.ts"}
	for _, fileName := range fileOrder {
		content, exists := files[fileName]
		if !exists {
			continue
		}
		targetPath := filepath.Join(typesDir, fileName)
		if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", fileName, err)
		}
		ctx.Printer.Success(fmt.Sprintf("Generated types/%s", fileName))
	}

	ctx.Printer.Println()
	ctx.Printer.Header("TypeScript types successfully generated in types/")
	return nil
}
