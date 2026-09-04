package diagnostics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nodephone/nodephone-cli/internal/output"
)

func TestDiagnosticsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/inspect":
			json.NewEncoder(w).Encode(ServerStatus{
				Version:   "v1.0.0",
				Status:    "Running",
				Uptime:    "1d 2h",
				CPU:       "5%",
				Memory:    "120 MB",
				Storage:   "1.0 GB",
				Database:  "20 MB",
				Users:     10,
				Realtime:  "3 clients",
				Functions: 5,
			})
		case "/api/v1/inspect/realtime":
			json.NewEncoder(w).Encode(RealtimeMetrics{
				ConnectedUsers: 10,
				ActiveRooms:    2,
				PresenceCount:  5,
				MessagesPerSec: "50 msgs/sec",
			})
		case "/api/v1/inspect/storage":
			json.NewEncoder(w).Encode(StorageAnalytics{
				PublicBucketSize:  "500 MB",
				PrivateBucketSize: "200 MB",
				TotalFiles:        50,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient()

	// 1. GetInspectStatus
	status, err := client.GetInspectStatus(server.URL, "token")
	if err != nil {
		t.Fatalf("GetInspectStatus failed: %v", err)
	}
	if status.Version != "v1.0.0" || status.Status != "Running" {
		t.Errorf("unexpected status response: %+v", status)
	}

	// 2. GetRealtimeMetrics
	rt, err := client.GetRealtimeMetrics(server.URL, "token")
	if err != nil {
		t.Fatalf("GetRealtimeMetrics failed: %v", err)
	}
	if rt.ConnectedUsers != 10 || rt.ActiveRooms != 2 {
		t.Errorf("unexpected realtime response: %+v", rt)
	}

	// 3. GetStorageAnalytics
	st, err := client.GetStorageAnalytics(server.URL, "token")
	if err != nil {
		t.Fatalf("GetStorageAnalytics failed: %v", err)
	}
	if st.TotalFiles != 50 {
		t.Errorf("unexpected storage analytics: %+v", st)
	}
}

func TestPrintFormattedLog(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	p := output.NewWithWriters(out, errOut)
	p.SetColorEnabled(false)

	client := NewClient()
	client.PrintFormattedLog(LogEntry{
		Level:    "INFO",
		Message:  "Server started",
		Duration: "10ms",
	}, p)

	if !strings.Contains(out.String(), "[INFO] Server started (10ms)") {
		t.Errorf("unexpected log output: %s", out.String())
	}
}
