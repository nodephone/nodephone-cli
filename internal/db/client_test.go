package db

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDBClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/db/migrations":
			records := []MigrationRecord{
				{Name: "001_initial", Checksum: "abc", AppliedAt: "2026-09-04T12:00:00Z"},
			}
			json.NewEncoder(w).Encode(records)
		case "/api/v1/db/push":
			var req PushRequest
			json.NewDecoder(r.Body).Decode(&req)
			applied := make([]string, 0)
			for _, m := range req.Migrations {
				applied = append(applied, m.Name)
			}
			json.NewEncoder(w).Encode(PushResponse{Applied: applied, Success: true})
		case "/api/v1/db/pull":
			json.NewEncoder(w).Encode(PullResponse{
				Migrations: []MigrationFile{
					{Name: "001_initial", Content: "CREATE TABLE users (id INT);", Checksum: "abc"},
				},
			})
		case "/api/v1/db/reset":
			w.WriteHeader(http.StatusOK)
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
	if len(status) != 1 || status[0].Name != "001_initial" {
		t.Errorf("unexpected status response: %+v", status)
	}

	// 2. PushMigrations
	pending := []MigrationFile{{Name: "002_auth", Content: "CREATE TABLE auth (id INT);"}}
	applied, err := client.PushMigrations(server.URL, "token", pending)
	if err != nil {
		t.Fatalf("PushMigrations failed: %v", err)
	}
	if len(applied) != 1 || applied[0] != "002_auth" {
		t.Errorf("unexpected applied migrations: %+v", applied)
	}

	// 3. PullSchema
	pulled, err := client.PullSchema(server.URL, "token")
	if err != nil {
		t.Fatalf("PullSchema failed: %v", err)
	}
	if len(pulled) != 1 || pulled[0].Name != "001_initial" {
		t.Errorf("unexpected pulled migrations: %+v", pulled)
	}

	// 4. ResetDatabase
	if err := client.ResetDatabase(server.URL, "token"); err != nil {
		t.Fatalf("ResetDatabase failed: %v", err)
	}
}
