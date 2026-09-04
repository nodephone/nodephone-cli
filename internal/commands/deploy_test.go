package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/deploy"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestDeployCommandFlow(t *testing.T) {
	tempProjectDir := t.TempDir()

	// Create valid nodephone.json
	os.WriteFile(filepath.Join(tempProjectDir, "nodephone.json"), []byte(`{"name":"test-app"}`), 0644)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health", "/health":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/deploy/status":
			json.NewEncoder(w).Encode(deploy.DeployStatus{
				ReleaseID: "rel_100",
				Version:   "v1.0.0",
				Env:       "production",
				Health:    "Healthy",
			})
		case "/api/v1/deploy":
			json.NewEncoder(w).Encode(deploy.DeployStatus{
				ReleaseID: "rel_101",
				Version:   "v1.0.0",
				Env:       "production",
				Health:    "Healthy",
			})
		case "/api/v1/deploy/rollback":
			json.NewEncoder(w).Encode(deploy.RollbackResult{
				PreviousReleaseID: "rel_100",
				Status:            "Success",
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	auth.SaveServerConfig(server.URL, "")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	ctx := &Context{
		Config:   &config.Config{ProjectRoot: tempProjectDir},
		Printer:  p,
		Registry: NewRegistry(),
	}

	depCmd := NewDeployCommand()

	// 1. deploy status
	out.Reset()
	if err := depCmd.Execute(ctx, []string{"status"}); err != nil {
		t.Fatalf("deploy status failed: %v", err)
	}
	if !strings.Contains(out.String(), "rel_100") || !strings.Contains(out.String(), "production") {
		t.Errorf("unexpected output from deploy status:\n%s", out.String())
	}

	// 2. deploy --dry-run
	out.Reset()
	if err := depCmd.Execute(ctx, []string{"--dry-run"}); err != nil {
		t.Fatalf("deploy --dry-run failed: %v", err)
	}
	if !strings.Contains(out.String(), "[DRY RUN]") || !strings.Contains(out.String(), "Dry-run complete") {
		t.Errorf("unexpected output from deploy --dry-run:\n%s", out.String())
	}

	// 3. deploy --prod
	out.Reset()
	if err := depCmd.Execute(ctx, []string{"--prod"}); err != nil {
		t.Fatalf("deploy --prod failed: %v", err)
	}
	if !strings.Contains(out.String(), "Deployment successful") {
		t.Errorf("unexpected output from deploy --prod:\n%s", out.String())
	}

	// 4. deploy rollback --force
	out.Reset()
	if err := depCmd.Execute(ctx, []string{"rollback", "--force"}); err != nil {
		t.Fatalf("deploy rollback failed: %v", err)
	}
	if !strings.Contains(out.String(), "Restored previous release") || !strings.Contains(out.String(), "rel_100") {
		t.Errorf("unexpected output from deploy rollback:\n%s", out.String())
	}
}
