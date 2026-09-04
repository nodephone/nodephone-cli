# Changelog

All notable changes to the NodePhone CLI (`nodephone/cli`) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-09-04

### Initial Production Release

#### Added
- **Core CLI Kernel (`nodephone`, `version`, `help`)**
  - Dependency-free Go CLI framework built with standard library `flag` and `os/exec`.
  - ANSI color-coded formatting for standard, warning, and error outputs (`internal/output`).
  - Runtime environment detection (OS, Arch, Go version, Git commit context).

- **Project Initialization (`nodephone init`)**
  - Scaffolds complete starter projects (`nodephone.json`, `.env.example`, `schema/001_initial.sql`, `functions/hello/index.js`, `storage/public`, `storage/private`).
  - Interactive/flag prompt confirmation before overwriting existing project directories.

- **Authentication & Local Credentials (`nodephone login`, `logout`, `whoami`)**
  - AES-256-GCM encrypted local credential storage in `~/.nodephone/credentials.json`.
  - Token refresh and server connectivity verification against NodePhone endpoints.

- **Database Synchronization (`nodephone db push`, `db pull`, `db status`, `db reset`)**
  - Migration tracker and runner applying SQL migration files sequentially.
  - Interactive reset protection and schema introspection generator.

- **Type & SDK Generator (`nodephone gen types`, `gen api`)**
  - Schema-to-TypeScript type generator converting OpenAPI 3.1 definitions into TypeScript interfaces (`types/api.ts`, `types/database.ts`, etc.).

- **Serverless Functions Workflow (`nodephone functions new`, `list`, `serve`, `deploy`, `logs`, `delete`)**
  - Local runtime server supporting Node.js functions with hot-reload simulation.
  - Deployment, log inspection, and deletion controls.

- **Diagnostics & Monitoring (`nodephone inspect`, `logs --follow`)**
  - Diagnostic metrics streaming, JSON output mode (`--json`), subsystem breakdown (`realtime`, `storage`).
  - Real-time log streamer with filter options.

- **Production Deployment Pipeline (`nodephone deploy`, `--prod`, `--dry-run`, `status`, `rollback`)**
  - Integrity validation, function archive build, pending migration execution, health verification, and rollback safety.

- **Cross-Platform Pre-built Binaries**
  - Precompiled binaries for Windows (`amd64`), Linux (`amd64`), and macOS (`amd64`, `arm64 Apple Silicon`) in `dist/`.
