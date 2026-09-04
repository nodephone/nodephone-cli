package deploy

import (
	"time"

	"github.com/nodephone/nodephone-cli/internal/db"
	"github.com/nodephone/nodephone-cli/internal/functions"
)

// DeployPlan holds diff calculations before executing deployment
type DeployPlan struct {
	Environment        string                   `json:"environment"`
	FunctionsToDeploy  []functions.FunctionInfo `json:"functions_to_deploy"`
	PendingMigrations  []db.MigrationFile       `json:"pending_migrations"`
	ConfigChanged      bool                     `json:"config_changed"`
	ConfigKeysModified int                      `json:"config_keys_modified"`
}

// DeployStatus holds active release details
type DeployStatus struct {
	ReleaseID  string    `json:"release_id"`
	Version    string    `json:"version"`
	Env        string    `json:"environment"`
	DeployedAt time.Time `json:"deployed_at"`
	Health     string    `json:"health"`
	ActiveURL  string    `json:"active_url"`
}

// RollbackResult holds release rollback outcome
type RollbackResult struct {
	PreviousReleaseID string    `json:"previous_release_id"`
	RestoredAt        time.Time `json:"restored_at"`
	Status            string    `json:"status"`
	Message           string    `json:"message"`
}

// ValidationResult holds project pre-deploy check results
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}
