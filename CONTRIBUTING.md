# Contributing to NodePhone CLI

Thank you for your interest in contributing to the NodePhone CLI (`nodephone/cli`)!

---

## 🛠️ Development Principles

1. **Pure Go Standard Library Only**: The CLI must remain free of external package dependencies (no Cobra, Viper, or 3rd party Go modules). Use `flag`, `os`, `net/http`, `crypto/cipher`, `encoding/json`, etc.
2. **Cross-Platform Compatibility**: Every feature and command must work seamlessly across Windows, Linux, and macOS. Avoid platform-specific shell commands or unhandled path separators (always use `filepath.Join` or POSIX normalized slashes).
3. **Comprehensive Testing**: Write unit tests (`_test.go`) alongside new commands and packages. Keep unit tests isolated using temporary test directories (`t.TempDir()`).
4. **Simple Commit Messages**: Write clear, natural human commit messages (e.g., `Add database pull command`, `Fix Windows path handling`). Avoid bot/AI jargon and automated conventional commit prefix requirements.

---

## 🚀 Setting Up Local Development

### Requirements
- [Go](https://go.dev/doc/install) 1.20 or newer

### Building and Testing

```bash
# Clone the repository
git clone https://github.com/nodephone/nodephone-cli.git
cd nodephone-cli

# Run all unit tests across all internal packages
go test -v ./...

# Run static analysis check
go vet ./...

# Build local executable
go build -o nodephone.exe ./cmd/nodephone

# Verify built executable
./nodephone.exe version
```

---

## 📦 Cross-Platform Build Target Matrix

To compile release binaries for supported architectures:

```bash
# Windows (amd64)
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o dist/nodephone-windows-amd64.exe ./cmd/nodephone

# Linux (amd64)
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o dist/nodephone-linux-amd64 ./cmd/nodephone

# macOS Intel (amd64)
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o dist/nodephone-darwin-amd64 ./cmd/nodephone

# macOS Apple Silicon (arm64)
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o dist/nodephone-darwin-arm64 ./cmd/nodephone
```

---

## 📄 Pull Request Guidelines

1. Fork the repo and create your feature branch (`git checkout -b feature/amazing-feature`).
2. Ensure `go test ./...` and `go vet ./...` pass with 0 errors.
3. Push your branch and open a Pull Request against `main`.
