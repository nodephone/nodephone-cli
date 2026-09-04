package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeChecksum(t *testing.T) {
	content := "CREATE TABLE test (id INT);"
	hash1 := ComputeChecksum(content)
	hash2 := ComputeChecksum(content)

	if hash1 == "" || hash1 != hash2 {
		t.Errorf("expected deterministic hash output, got hash1: %s, hash2: %s", hash1, hash2)
	}

	differentHash := ComputeChecksum("CREATE TABLE test2 (id INT);")
	if hash1 == differentHash {
		t.Error("expected different hash for different SQL content")
	}
}

func TestReadLocalMigrations(t *testing.T) {
	tempDir := t.TempDir()
	schemaDir := filepath.Join(tempDir, "schema")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("failed to create temp schema dir: %v", err)
	}

	// Write mock migration files
	file1 := filepath.Join(schemaDir, "002_auth.sql")
	file2 := filepath.Join(schemaDir, "001_initial.sql")
	file3 := filepath.Join(schemaDir, "README.txt") // Non-SQL file

	os.WriteFile(file1, []byte("CREATE TABLE auth (id INT);"), 0644)
	os.WriteFile(file2, []byte("CREATE TABLE users (id INT);"), 0644)
	os.WriteFile(file3, []byte("Ignore this file"), 0644)

	migrations, err := ReadLocalMigrations(schemaDir)
	if err != nil {
		t.Fatalf("unexpected error reading migrations: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 SQL migrations, got %d", len(migrations))
	}

	// Verify sorting (001_initial should come before 002_auth)
	if migrations[0].Name != "001_initial" {
		t.Errorf("expected first migration to be '001_initial', got %s", migrations[0].Name)
	}
	if migrations[1].Name != "002_auth" {
		t.Errorf("expected second migration to be '002_auth', got %s", migrations[1].Name)
	}

	if migrations[0].Checksum == "" || migrations[1].Checksum == "" {
		t.Error("expected checksums to be populated")
	}
}
