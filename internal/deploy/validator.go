package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nodephone/nodephone-cli/internal/db"
	"github.com/nodephone/nodephone-cli/internal/functions"
)

// ValidateProject verifies project integrity before deployment
func ValidateProject(projectRoot string) *ValidationResult {
	res := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// 1. Check nodephone.json
	cfgPath := filepath.Join(projectRoot, "nodephone.json")
	cfgBytes, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		res.Valid = false
		res.Errors = append(res.Errors, "missing nodephone.json configuration file. Run 'nodephone init' to create project config.")
	} else if err != nil {
		res.Valid = false
		res.Errors = append(res.Errors, fmt.Sprintf("failed to read nodephone.json: %v", err))
	} else {
		var dummy map[string]any
		if err := json.Unmarshal(cfgBytes, &dummy); err != nil {
			res.Valid = false
			res.Errors = append(res.Errors, fmt.Sprintf("nodephone.json contains invalid JSON: %v", err))
		}
	}

	// 2. Check schema/ directory SQL files
	schemaDir := filepath.Join(projectRoot, "schema")
	if _, err := db.ReadLocalMigrations(schemaDir); err != nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("schema validation issue: %v", err))
	}

	// 3. Check functions/ directory manifests
	funcsDir := filepath.Join(projectRoot, "functions")
	if funcs, err := functions.ScanLocalFunctions(funcsDir); err == nil {
		for _, fn := range funcs {
			if err := fn.Manifest.Validate(); err != nil {
				res.Valid = false
				res.Errors = append(res.Errors, fmt.Sprintf("function %s manifest error: %v", fn.Name, err))
			}
		}
	}

	return res
}
