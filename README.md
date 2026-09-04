# NodePhone CLI (`nodephone-cli`)

The production-ready command line interface and developer kernel for **NodePhone**.

---

## 🚀 Overview

`nodephone` is the unified CLI entry point for NodePhone developer tools, database migrations, serverless function deployments, SDK generation, log streaming, and system diagnostics.

Built with **pure Go standard library** for zero external dependencies, ultra-fast performance, and cross-platform reliability.

---

## 📁 Repository Structure

```text
cli/
├── cmd/
│   └── nodephone/
│       └── main.go       # Minimal CLI entry point
├── dist/                 # Precompiled cross-platform v1.0.0 release binaries
├── internal/
│   ├── app/              # CLI kernel manager & argument dispatcher
│   ├── auth/             # AES-256-GCM encrypted credential storage
│   ├── commands/         # Modular command interface & registry
│   ├── config/           # Project & global configuration loader
│   ├── deploy/           # Production deployment pipeline & validator
│   ├── diagnostics/      # Log streamer & inspect telemetry
│   ├── init/             # Project scaffolding engine
│   ├── migrations/       # Database SQL migration manager & generator
│   ├── output/           # ANSI colored output abstraction
│   ├── serverless/       # Local function emulator & manager
│   ├── typesgen/         # OpenAPI 3.1 TypeScript code generator
│   └── version/          # Build metadata & version provider
├── pkg/                  # Exported library packages
├── go.mod                # Go module specification (github.com/nodephone/nodephone-cli)
├── CHANGELOG.md          # Version release history
├── CONTRIBUTING.md       # Development & contribution guidelines
├── LICENSE               # MIT License
├── README.md             # Project documentation
└── SECURITY.md           # Encryption specs & disclosure policy
```

---

## 🛠️ Installation & Building

### Precompiled Binaries

Download precompiled release binaries for your operating system from the [`dist/`](dist/) folder:

- **Windows (x64)**: [`dist/nodephone-windows-amd64.exe`](dist/nodephone-windows-amd64.exe)
- **Linux (x64)**: [`dist/nodephone-linux-amd64`](dist/nodephone-linux-amd64)
- **macOS (Intel)**: [`dist/nodephone-darwin-amd64`](dist/nodephone-darwin-amd64)
- **macOS (Apple Silicon)**: [`dist/nodephone-darwin-arm64`](dist/nodephone-darwin-arm64)

### Building from Source

Prerequisites: [Go](https://go.dev/doc/install) 1.20+

```bash
# Clone repository
git clone https://github.com/nodephone/nodephone-cli.git
cd nodephone-cli

# Compile binary
go build -o nodephone.exe ./cmd/nodephone

# Test installation
./nodephone.exe version
```

---

## 💻 Complete Command Reference

| Command | Subcommand / Flag | Description |
| :--- | :--- | :--- |
| `nodephone` | `help` | Display CLI header, usage guide, and command list |
| `nodephone version` | | Output CLI version, Go runtime, OS/Arch, and commit hash |
| `nodephone init` | `<app-name>` | Scaffold a new production-ready NodePhone project |
| `nodephone login` | | Authenticate with NodePhone server & store encrypted credentials |
| `nodephone logout` | | Remove local encrypted credentials |
| `nodephone whoami` | | Display active authenticated user and server URL |
| `nodephone db` | `push` | Apply pending SQL migrations to server |
| | `pull` | Introspect server schema & generate local SQL migration |
| | `status` | Check local vs remote database migration state |
| | `reset` | Reset remote database schema (requires confirmation) |
| `nodephone gen` | `types` | Generate TypeScript models (`types/`) from server OpenAPI 3.1 schema |
| | `api` | Alias for `gen types` |
| `nodephone functions`| `new <name>` | Create a new serverless function starter template |
| | `list` | List all local serverless functions |
| | `serve` | Run local serverless function emulator server |
| | `deploy [name]` | Deploy serverless function(s) to NodePhone server |
| | `logs <name>` | Fetch log output for specific function |
| | `delete <name>` | Remove function from connected NodePhone server |
| `nodephone logs` | `--follow` | Stream live server logs with optional continuous follow mode |
| `nodephone inspect` | `--json` | Inspect system telemetry, health, database, and storage state |
| | `realtime` | Inspect WebSocket realtime connection metrics |
| | `storage` | Inspect storage usage, bucket counts, and file stats |
| `nodephone deploy` | `--prod` | Execute full project production deployment pipeline |
| | `--dry-run` | Perform pre-deployment validation check without applying changes |
| | `status` | View deployment history and current release status |
| | `rollback` | Rollback target environment to previous deployment release |

---

## 🧪 Running Tests & Quality Audit

Run the test suite across all 11 internal packages:

```bash
go test -v ./...
go vet ./...
```

---

## 📄 License

Distributed under the MIT License. See [`LICENSE`](LICENSE) for details.