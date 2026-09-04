package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MigrationFile represents a local SQL migration script
type MigrationFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Content  string `json:"content"`
	Checksum string `json:"checksum"`
}

// MigrationRecord represents a remote applied migration record
type MigrationRecord struct {
	Name      string `json:"name"`
	Checksum  string `json:"checksum"`
	AppliedAt string `json:"applied_at"`
}

// StatusSummary represents local vs remote migration comparison
type StatusSummary struct {
	LocalCount  int               `json:"local_count"`
	RemoteCount int               `json:"remote_count"`
	Pending     []MigrationFile   `json:"pending"`
	Synced      bool              `json:"synced"`
	History     []MigrationRecord `json:"history"`
}

// PushRequest payload for push API
type PushRequest struct {
	Migrations []MigrationFile `json:"migrations"`
}

// PushResponse response from push API
type PushResponse struct {
	Applied []string `json:"applied"`
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
}

// PullResponse payload from pull API
type PullResponse struct {
	Migrations []MigrationFile `json:"migrations"`
}

// ComputeChecksum calculates SHA-256 hash of migration string content
func ComputeChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ReadLocalMigrations reads, sorts, and hashes SQL files from schema/ directory
func ReadLocalMigrations(schemaDir string) ([]MigrationFile, error) {
	info, err := os.Stat(schemaDir)
	if os.IsNotExist(err) {
		return []MigrationFile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access schema directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("schema path %s is not a directory", schemaDir)
	}

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema directory: %w", err)
	}

	var files []MigrationFile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		fullPath := filepath.Join(schemaDir, entry.Name())
		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		content := string(contentBytes)
		checksum := ComputeChecksum(content)

		// Base name without extension for clean identifier (e.g. 001_initial)
		baseName := strings.TrimSuffix(entry.Name(), ".sql")

		files = append(files, MigrationFile{
			Name:     baseName,
			Path:     fullPath,
			Content:  content,
			Checksum: checksum,
		})
	}

	// Sort migrations alphabetically / numerically by file name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}
