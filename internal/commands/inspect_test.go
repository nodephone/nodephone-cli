package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/auth"
	"github.com/nodephone/nodephone-cli/internal/config"
	"github.com/nodephone/nodephone-cli/internal/diagnostics"
	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestInspectAndLogsCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/inspect":
			json.NewEncoder(w).Encode(diagnostics.ServerStatus{
				Version:   "v1.0.0",
				Status:    "Running",
				Uptime:    "2d 14h",
				CPU:       "8%",
				Memory:    "146 MB",
				Storage:   "1.8 GB",
				Database:  "42 MB",
				Users:     28,
				Realtime:  "6 clients",
				Functions: 12,
			})
		case "/api/v1/inspect/realtime":
			json.NewEncoder(w).Encode(diagnostics.RealtimeMetrics{
				ConnectedUsers: 28,
				ActiveRooms:    6,
				PresenceCount:  15,
				MessagesPerSec: "125 msgs/sec",
			})
		case "/api/v1/inspect/storage":
			json.NewEncoder(w).Encode(diagnostics.StorageAnalytics{
				PublicBucketSize:  "1.2 GB",
				PrivateBucketSize: "600 MB",
				TotalFiles:        142,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	auth.SaveServerConfig(server.URL, "")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	ctx := &Context{
		Config:   &config.Config{ProjectRoot: t.TempDir()},
		Printer:  p,
		Registry: NewRegistry(),
	}

	inspectCmd := NewInspectCommand()
	logsCmd := NewLogsCommand()

	// 1. inspect default
	out.Reset()
	if err := inspectCmd.Execute(ctx, nil); err != nil {
		t.Fatalf("inspect default failed: %v", err)
	}
	if !strings.Contains(out.String(), "NodePhone Server") || !strings.Contains(out.String(), "146 MB") {
		t.Errorf("unexpected output from inspect:\n%s", out.String())
	}

	// 2. inspect --json
	out.Reset()
	if err := inspectCmd.Execute(ctx, []string{"--json"}); err != nil {
		t.Fatalf("inspect --json failed: %v", err)
	}
	if !strings.Contains(out.String(), `"memory": "146 MB"`) {
		t.Errorf("unexpected JSON output from inspect:\n%s", out.String())
	}

	// 3. inspect realtime
	out.Reset()
	if err := inspectCmd.Execute(ctx, []string{"realtime"}); err != nil {
		t.Fatalf("inspect realtime failed: %v", err)
	}
	if !strings.Contains(out.String(), "Connected Users") || !strings.Contains(out.String(), "28") {
		t.Errorf("unexpected output from inspect realtime:\n%s", out.String())
	}

	// 4. inspect storage
	out.Reset()
	if err := inspectCmd.Execute(ctx, []string{"storage"}); err != nil {
		t.Fatalf("inspect storage failed: %v", err)
	}
	if !strings.Contains(out.String(), "Public Bucket") || !strings.Contains(out.String(), "1.2 GB") {
		t.Errorf("unexpected output from inspect storage:\n%s", out.String())
	}

	// 5. logs command
	out.Reset()
	if err := logsCmd.Execute(ctx, []string{"hello"}); err != nil {
		t.Fatalf("logs failed: %v", err)
	}
	if !strings.Contains(out.String(), "Server started") || !strings.Contains(out.String(), "[INFO]") {
		t.Errorf("unexpected output from logs:\n%s", out.String())
	}
}
