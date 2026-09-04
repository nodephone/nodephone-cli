package functions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// FunctionInfo represents a local serverless function item
type FunctionInfo struct {
	Name       string           `json:"name"`
	Path       string           `json:"path"`
	Manifest   FunctionManifest `json:"manifest"`
	Code       string           `json:"code"`
	Checksum   string           `json:"checksum"`
	IsDeployed bool             `json:"is_deployed"`
}

// ComputeChecksum calculates SHA-256 hash of code + manifest content
func ComputeChecksum(code string, manifest JSONRaw) string {
	hasher := sha256.New()
	hasher.Write([]byte(code))
	hasher.Write([]byte(manifest))
	return hex.EncodeToString(hasher.Sum(nil))
}

type JSONRaw string

// ScanLocalFunctions reads all valid function directories under functionsDir
func ScanLocalFunctions(functionsDir string) ([]FunctionInfo, error) {
	info, err := os.Stat(functionsDir)
	if os.IsNotExist(err) {
		return []FunctionInfo{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access functions directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", functionsDir)
	}

	entries, err := os.ReadDir(functionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read functions directory: %w", err)
	}

	var results []FunctionInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fnName := entry.Name()
		fnDir := filepath.Join(functionsDir, fnName)

		// Read manifest function.json if exists, else fallback default
		manifestPath := filepath.Join(fnDir, "function.json")
		manifest := DefaultManifest(fnName)

		if manifestBytes, err := os.ReadFile(manifestPath); err == nil {
			_ = json.Unmarshal(manifestBytes, &manifest)
		}

		// Read entrypoint code file (e.g. index.js)
		codePath := filepath.Join(fnDir, manifest.Entrypoint)
		codeBytes, err := os.ReadFile(codePath)
		if err != nil {
			// Skip if entrypoint code file missing
			continue
		}

		code := string(codeBytes)
		manifestRaw, _ := json.Marshal(manifest)
		checksum := ComputeChecksum(code, JSONRaw(manifestRaw))

		results = append(results, FunctionInfo{
			Name:     fnName,
			Path:     fnDir,
			Manifest: manifest,
			Code:     code,
			Checksum: checksum,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}
