package functions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFunctionManifest(t *testing.T) {
	manifest := DefaultManifest("hello")
	if err := manifest.Validate(); err != nil {
		t.Fatalf("expected valid manifest, got error: %v", err)
	}

	jsonStr, err := MarshalManifest(manifest)
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}

	if !json.Valid([]byte(jsonStr)) {
		t.Error("expected valid JSON output from MarshalManifest")
	}
}

func TestScanLocalFunctions(t *testing.T) {
	tempDir := t.TempDir()
	funcsDir := filepath.Join(tempDir, "functions")

	// Create function directories
	fn1Dir := filepath.Join(funcsDir, "hello")
	fn2Dir := filepath.Join(funcsDir, "cleanup")
	os.MkdirAll(fn1Dir, 0755)
	os.MkdirAll(fn2Dir, 0755)

	os.WriteFile(filepath.Join(fn1Dir, "index.js"), []byte("module.exports = async function() {};"), 0644)
	os.WriteFile(filepath.Join(fn2Dir, "index.js"), []byte("module.exports = async function() {};"), 0644)

	results, err := ScanLocalFunctions(funcsDir)
	if err != nil {
		t.Fatalf("ScanLocalFunctions failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(results))
	}

	// Verify sorting (cleanup before hello)
	if results[0].Name != "cleanup" || results[1].Name != "hello" {
		t.Errorf("expected sorted functions, got: %s, %s", results[0].Name, results[1].Name)
	}

	if results[0].Checksum == "" || results[1].Checksum == "" {
		t.Error("expected checksums to be calculated")
	}
}

func TestFunctionsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/functions":
			json.NewEncoder(w).Encode([]RemoteFunction{
				{Name: "hello", Runtime: "nodejs18", Checksum: "abc"},
			})
		case "/api/v1/functions/deploy":
			json.NewEncoder(w).Encode(DeployResponse{
				Deployed: []string{"hello", "cleanup"},
				Success:  true,
			})
		case "/api/v1/functions/hello":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient()

	// 1. ListRemoteFunctions
	list, err := client.ListRemoteFunctions(server.URL, "token")
	if err != nil {
		t.Fatalf("ListRemoteFunctions failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "hello" {
		t.Errorf("unexpected list output: %+v", list)
	}

	// 2. DeployFunctions
	deployed, err := client.DeployFunctions(server.URL, "token", []FunctionInfo{{Name: "hello"}})
	if err != nil {
		t.Fatalf("DeployFunctions failed: %v", err)
	}
	if len(deployed) != 2 {
		t.Errorf("unexpected deployed output: %+v", deployed)
	}

	// 3. DeleteFunction
	if err := client.DeleteFunction(server.URL, "token", "hello"); err != nil {
		t.Fatalf("DeleteFunction failed: %v", err)
	}
}
