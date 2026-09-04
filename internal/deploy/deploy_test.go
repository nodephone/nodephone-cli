package deploy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateProject(t *testing.T) {
	tempDir := t.TempDir()

	// Missing nodephone.json
	res := ValidateProject(tempDir)
	if res.Valid {
		t.Error("expected validation to fail for missing nodephone.json")
	}

	// Create valid nodephone.json
	os.WriteFile(filepath.Join(tempDir, "nodephone.json"), []byte(`{"name":"test-app"}`), 0644)
	res = ValidateProject(tempDir)
	if !res.Valid {
		t.Errorf("expected project validation to pass, got errors: %v", res.Errors)
	}
}

func TestDeployClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/deploy/status":
			json.NewEncoder(w).Encode(DeployStatus{
				ReleaseID: "rel_123",
				Version:   "v1.0.0",
				Env:       "production",
				Health:    "Healthy",
			})
		case "/api/v1/deploy":
			json.NewEncoder(w).Encode(DeployStatus{
				ReleaseID: "rel_124",
				Version:   "v1.0.0",
				Env:       "production",
				Health:    "Healthy",
			})
		case "/api/v1/deploy/rollback":
			json.NewEncoder(w).Encode(RollbackResult{
				PreviousReleaseID: "rel_123",
				Status:            "Success",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient()

	// 1. GetStatus
	status, err := client.GetStatus(server.URL, "token")
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status.ReleaseID != "rel_123" || status.Env != "production" {
		t.Errorf("unexpected deploy status: %+v", status)
	}

	// 2. ExecuteDeploy
	plan := &DeployPlan{Environment: "production"}
	newStatus, err := client.ExecuteDeploy(server.URL, "token", plan)
	if err != nil {
		t.Fatalf("ExecuteDeploy failed: %v", err)
	}
	if newStatus.ReleaseID != "rel_124" {
		t.Errorf("unexpected execute deploy status: %+v", newStatus)
	}

	// 3. TriggerRollback
	rb, err := client.TriggerRollback(server.URL, "token")
	if err != nil {
		t.Fatalf("TriggerRollback failed: %v", err)
	}
	if rb.PreviousReleaseID != "rel_123" {
		t.Errorf("unexpected rollback result: %+v", rb)
	}
}
