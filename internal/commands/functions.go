package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/functions"
)

type FunctionsCommand struct {
	fnClient   *functions.Client
	authClient *auth.Client
}

func NewFunctionsCommand() Command {
	return &FunctionsCommand{
		fnClient:   functions.NewClient(),
		authClient: auth.NewClient(),
	}
}

func (c *FunctionsCommand) Name() string {
	return "functions"
}

func (c *FunctionsCommand) Description() string {
	return "Manage, serve, and deploy serverless functions"
}

func (c *FunctionsCommand) Usage() string {
	return "nodephone functions <new|list|serve|deploy|delete|logs> [args]"
}

func (c *FunctionsCommand) Execute(ctx *Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp(ctx)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "new":
		return c.handleNew(ctx, subArgs)
	case "list":
		return c.handleList(ctx, subArgs)
	case "serve":
		return c.handleServe(ctx, subArgs)
	case "deploy":
		return c.handleDeploy(ctx, subArgs)
	case "delete":
		return c.handleDelete(ctx, subArgs)
	case "logs":
		return c.handleLogs(ctx, subArgs)
	case "help", "--help", "-h":
		return c.showHelp(ctx)
	default:
		return fmt.Errorf("unknown functions subcommand %q. Run 'nodephone functions help' for usage", subcommand)
	}
}

func (c *FunctionsCommand) showHelp(ctx *Context) error {
	ctx.Printer.Header("NodePhone Functions Commands")
	ctx.Printer.Println()
	ctx.Printer.Println("Usage: " + c.Usage())
	ctx.Printer.Println()
	ctx.Printer.Println("Subcommands:")
	ctx.Printer.TwoColumn("new <name>", 14, "Create a new serverless function with starter template")
	ctx.Printer.TwoColumn("list", 14, "Display local and remote serverless functions")
	ctx.Printer.TwoColumn("serve", 14, "Start local development server with hot reload and live logs")
	ctx.Printer.TwoColumn("deploy", 14, "Deploy changed functions to NodePhone server")
	ctx.Printer.TwoColumn("logs <name>", 14, "Stream real-time logs for a function")
	ctx.Printer.TwoColumn("delete <name>", 14, "Remove a serverless function locally and remotely")
	return nil
}

func (c *FunctionsCommand) getAuthContext() (string, string, error) {
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

func (c *FunctionsCommand) handleNew(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("function name is required. Usage: nodephone functions new <function-name>")
	}

	fnName := strings.TrimSpace(args[0])
	if fnName == "" {
		return fmt.Errorf("invalid function name")
	}

	fnDir := filepath.Join(ctx.Config.ProjectRoot, "functions", fnName)
	if err := os.MkdirAll(fnDir, 0755); err != nil {
		return fmt.Errorf("failed to create function directory: %w", err)
	}

	// 1. Write index.js
	jsPath := filepath.Join(fnDir, "index.js")
	if err := os.WriteFile(jsPath, []byte(functions.StarterTemplate(fnName)), 0644); err != nil {
		return fmt.Errorf("failed to write index.js: %w", err)
	}

	// 2. Write function.json
	manifest := functions.DefaultManifest(fnName)
	manifestStr, err := functions.MarshalManifest(manifest)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %w", err)
	}
	manifestPath := filepath.Join(fnDir, "function.json")
	if err := os.WriteFile(manifestPath, []byte(manifestStr), 0644); err != nil {
		return fmt.Errorf("failed to write function.json: %w", err)
	}

	ctx.Printer.Success(fmt.Sprintf("Created function %q in functions/%s/", fnName, fnName))
	ctx.Printer.Println()
	ctx.Printer.Println("Files created:")
	ctx.Printer.Println(fmt.Sprintf("  functions/%s/index.js", fnName))
	ctx.Printer.Println(fmt.Sprintf("  functions/%s/function.json", fnName))
	ctx.Printer.Println()
	ctx.Printer.Println("To serve locally:")
	ctx.Printer.Println("  nodephone functions serve")
	ctx.Printer.Println()

	return nil
}

func (c *FunctionsCommand) handleList(ctx *Context, args []string) error {
	funcsDir := filepath.Join(ctx.Config.ProjectRoot, "functions")
	localFuncs, err := functions.ScanLocalFunctions(funcsDir)
	if err != nil {
		return fmt.Errorf("failed to scan local functions: %w", err)
	}

	serverURL, token, _ := c.getAuthContext()
	remoteFuncs, _ := c.fnClient.ListRemoteFunctions(serverURL, token)

	localCount := len(localFuncs)
	remoteCount := len(remoteFuncs)

	ctx.Printer.Header("NodePhone Serverless Functions")
	ctx.Printer.Println()
	ctx.Printer.TwoColumn("Local", 8, fmt.Sprintf("%d", localCount))
	ctx.Printer.TwoColumn("Remote", 8, fmt.Sprintf("%d", remoteCount))
	ctx.Printer.Println()

	if localCount > 0 {
		ctx.Printer.Header("Local Functions:")
		for _, fn := range localFuncs {
			ctx.Printer.Println("  " + fn.Name)
		}
		ctx.Printer.Println()
	}

	return nil
}

func (c *FunctionsCommand) handleServe(ctx *Context, args []string) error {
	funcsDir := filepath.Join(ctx.Config.ProjectRoot, "functions")
	server := functions.NewServer(funcsDir, 8080, ctx.Printer)

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return server.Start(cancelCtx)
}

func (c *FunctionsCommand) handleDeploy(ctx *Context, args []string) error {
	serverURL, token, err := c.getAuthContext()
	if err != nil {
		return err
	}

	if err := c.authClient.PingServer(serverURL); err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", serverURL, err)
	}
	ctx.Printer.Success("Connected")

	funcsDir := filepath.Join(ctx.Config.ProjectRoot, "functions")
	localFuncs, err := functions.ScanLocalFunctions(funcsDir)
	if err != nil {
		return fmt.Errorf("failed to scan local functions: %w", err)
	}

	if len(localFuncs) == 0 {
		ctx.Printer.Info("No local functions found in functions/")
		return nil
	}

	remoteFuncs, _ := c.fnClient.ListRemoteFunctions(serverURL, token)
	remoteMap := make(map[string]string)
	for _, rf := range remoteFuncs {
		remoteMap[rf.Name] = rf.Checksum
	}

	var toDeploy []functions.FunctionInfo
	for _, lf := range localFuncs {
		if remoteChecksum, exists := remoteMap[lf.Name]; !exists || remoteChecksum != lf.Checksum {
			toDeploy = append(toDeploy, lf)
		}
	}

	if len(toDeploy) == 0 {
		ctx.Printer.Info("All functions are up to date (0 pending deployments).")
		return nil
	}

	deployedNames, err := c.fnClient.DeployFunctions(serverURL, token, toDeploy)
	if err != nil {
		return err
	}

	for _, name := range deployedNames {
		ctx.Printer.Success(fmt.Sprintf("%s deployed", name))
	}

	ctx.Printer.Println()
	ctx.Printer.Header(fmt.Sprintf("%d functions synchronized.", len(deployedNames)))
	return nil
}

func (c *FunctionsCommand) handleDelete(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("function name is required. Usage: nodephone functions delete <function-name>")
	}

	fnName := args[0]
	force := false

	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		}
	}

	if !force {
		ctx.Printer.Warn(fmt.Sprintf("Are you sure you want to delete function %q? [y/N]: ", fnName))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			ctx.Printer.Info("Deletion cancelled.")
			return nil
		}
	}

	serverURL, token, _ := c.getAuthContext()
	_ = c.fnClient.DeleteFunction(serverURL, token, fnName)

	// Remove local folder
	fnDir := filepath.Join(ctx.Config.ProjectRoot, "functions", fnName)
	_ = os.RemoveAll(fnDir)

	ctx.Printer.Success(fmt.Sprintf("Function %q successfully deleted.", fnName))
	return nil
}

func (c *FunctionsCommand) handleLogs(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("function name is required. Usage: nodephone functions logs <function-name>")
	}

	fnName := args[0]
	serverURL, token, err := c.getAuthContext()
	if err != nil {
		return err
	}

	ctx.Printer.Header(fmt.Sprintf("Streaming logs for function %q...", fnName))
	ctx.Printer.Println()

	return c.fnClient.StreamLogs(serverURL, token, fnName, os.Stdout)
}
