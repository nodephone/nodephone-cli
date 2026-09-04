package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var validProjectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

// ConfigFileSpec defines the structure of nodephone.json
type ConfigFileSpec struct {
	SchemaVersion string          `json:"$schema,omitempty"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	Schema        SchemaConfig    `json:"schema"`
	Functions     FunctionsConfig `json:"functions"`
	Storage       StorageConfig   `json:"storage"`
}

type SchemaConfig struct {
	Path string `json:"path"`
}

type FunctionsConfig struct {
	Path string `json:"path"`
}

type StorageConfig struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

type InitCommand struct{}

func NewInitCommand() Command {
	return &InitCommand{}
}

func (c *InitCommand) Name() string {
	return "init"
}

func (c *InitCommand) Description() string {
	return "Initialize a new NodePhone project scaffold"
}

func (c *InitCommand) Usage() string {
	return "nodephone init <project-name> [--force]"
}

func (c *InitCommand) Execute(ctx *Context, args []string) error {
	var projectName string
	force := false

	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else if !strings.HasPrefix(arg, "-") && projectName == "" {
			projectName = arg
		}
	}

	if projectName == "" {
		return fmt.Errorf("missing project name. Usage: %s", c.Usage())
	}

	if err := ValidateProjectName(projectName); err != nil {
		return fmt.Errorf("invalid project name %q: %w", projectName, err)
	}

	targetDir := filepath.Join(ctx.Config.ProjectRoot, projectName)

	// Check directory status
	if err := checkTargetDirectory(targetDir, force); err != nil {
		return err
	}

	ctx.Printer.Header(fmt.Sprintf("Initializing NodePhone project %q...", projectName))
	ctx.Printer.Println()

	// 1. Create Directory Tree
	directories := []string{
		targetDir,
		filepath.Join(targetDir, "schema"),
		filepath.Join(targetDir, "functions", "hello"),
		filepath.Join(targetDir, "storage", "public"),
		filepath.Join(targetDir, "storage", "private"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	ctx.Printer.Success("Created directory structure")

	// 2. Generate nodephone.json
	configContent, err := generateNodephoneJSON(projectName)
	if err != nil {
		return fmt.Errorf("failed to generate nodephone.json: %w", err)
	}
	if err := writeFile(filepath.Join(targetDir, "nodephone.json"), configContent); err != nil {
		return fmt.Errorf("failed to write nodephone.json: %w", err)
	}
	ctx.Printer.Success("Generated nodephone.json")

	// 3. Generate .env.example
	envContent := generateEnvExample(projectName)
	if err := writeFile(filepath.Join(targetDir, ".env.example"), envContent); err != nil {
		return fmt.Errorf("failed to write .env.example: %w", err)
	}
	ctx.Printer.Success("Generated .env.example")

	// 4. Generate schema/001_initial.sql
	sqlContent := generateStarterSQL(projectName)
	if err := writeFile(filepath.Join(targetDir, "schema", "001_initial.sql"), sqlContent); err != nil {
		return fmt.Errorf("failed to write schema/001_initial.sql: %w", err)
	}
	ctx.Printer.Success("Created schema/001_initial.sql")

	// 5. Generate functions/hello/index.js
	jsContent := generateSampleFunction()
	if err := writeFile(filepath.Join(targetDir, "functions", "hello", "index.js"), jsContent); err != nil {
		return fmt.Errorf("failed to write functions/hello/index.js: %w", err)
	}
	ctx.Printer.Success("Created functions/hello/index.js")

	// 6. Generate storage .gitkeep files
	if err := writeFile(filepath.Join(targetDir, "storage", "public", ".gitkeep"), ""); err != nil {
		return fmt.Errorf("failed to write storage/public/.gitkeep: %w", err)
	}
	if err := writeFile(filepath.Join(targetDir, "storage", "private", ".gitkeep"), ""); err != nil {
		return fmt.Errorf("failed to write storage/private/.gitkeep: %w", err)
	}
	ctx.Printer.Success("Created storage directories (public/ & private/)")

	// 7. Generate README.md
	readmeContent := generateReadme(projectName)
	if err := writeFile(filepath.Join(targetDir, "README.md"), readmeContent); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}
	ctx.Printer.Success("Created README.md")

	ctx.Printer.Println()
	ctx.Printer.Header(fmt.Sprintf("Successfully initialized project %q!", projectName))
	ctx.Printer.Println()
	ctx.Printer.Println("Next steps:")
	ctx.Printer.Println(fmt.Sprintf("  cd %s", projectName))
	ctx.Printer.Println("  nodephone help")
	ctx.Printer.Println()

	return nil
}

// ValidateProjectName verifies project name follows allowed conventions
func ValidateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("project name cannot be empty")
	}

	if name == "." || name == ".." {
		return errors.New("project name cannot be '.' or '..'")
	}

	if !validProjectNameRegex.MatchString(name) {
		return errors.New("project name can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}

func checkTargetDirectory(dirPath string, force bool) error {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to inspect destination path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("target path %s exists and is not a directory", dirPath)
	}

	// Check if directory is non-empty
	f, err := os.Open(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read destination directory: %w", err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		// Directory contains files
		if !force {
			return fmt.Errorf("directory %q already exists and is not empty. Use --force to overwrite", dirPath)
		}
	} else if err != io.EOF {
		return fmt.Errorf("error reading directory content: %w", err)
	}

	return nil
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func generateNodephoneJSON(projectName string) (string, error) {
	cfg := ConfigFileSpec{
		SchemaVersion: "https://nodephone.dev/schema/config.v1.json",
		Name:          projectName,
		Version:       "0.1.0",
		Schema: SchemaConfig{
			Path: "./schema",
		},
		Functions: FunctionsConfig{
			Path: "./functions",
		},
		Storage: StorageConfig{
			Public:  "./storage/public",
			Private: "./storage/private",
		},
	}

	bytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes) + "\n", nil
}

func generateEnvExample(projectName string) string {
	return fmt.Sprintf(`# NodePhone Environment Configuration for %s

# Runtime Environment
NODEPHONE_ENV=development
NODEPHONE_PORT=8080

# Database Connection
DATABASE_URL=postgres://postgres:postgres@localhost:5432/%s?sslmode=disable

# Storage Configuration
STORAGE_PUBLIC_PATH=./storage/public
STORAGE_PRIVATE_PATH=./storage/private
`, projectName, projectName)
}

func generateStarterSQL(projectName string) string {
	return fmt.Sprintf(`-- Starter Database Migration for %s
-- Created by nodephone init

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`, projectName)
}

func generateSampleFunction() string {
	return `/**
 * Sample NodePhone Function: hello
 */
module.exports = async function handler(req, res) {
  const name = req.query?.name || 'Developer';

  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json'
    },
    body: {
      message: ` + "`" + `Hello, ${name}! Welcome to NodePhone.` + "`" + `,
      timestamp: new Date().toISOString()
    }
  };
};
`
}

func generateReadme(projectName string) string {
	return fmt.Sprintf(`# %s

A NodePhone cloud application.

## Directory Overview

- **nodephone.json**: Main application configuration file.
- **.env.example**: Environment variables template.
- **schema/**: SQL database migrations (001_initial.sql).
- **functions/**: Serverless JavaScript function handlers (hello/index.js).
- **storage/**: Public and private static asset storage.

## Getting Started

1. Copy .env.example to .env:
   %s
   cp .env.example .env
   %s

2. Run NodePhone commands:
   %s
   nodephone help
   %s
`, projectName, "```bash", "```", "```bash", "```")
}
