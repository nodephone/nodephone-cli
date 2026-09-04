package generator

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleOpenAPIJSON = `{
  "openapi": "3.1.0",
  "info": { "title": "NodePhone API", "version": "1.0.0" },
  "paths": {
    "/auth/login": {
      "post": {
        "summary": "User Login",
        "operationId": "loginUser",
        "responses": { "200": { "description": "Successful login" } }
      }
    }
  },
  "components": {
    "schemas": {
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": { "type": "string", "description": "User email address" },
          "password": { "type": "string", "description": "Account password" }
        }
      },
      "LoginResponse": {
        "type": "object",
        "required": ["access_token", "refresh_token"],
        "properties": {
          "access_token": { "type": "string" },
          "refresh_token": { "type": "string" }
        }
      },
      "UserRole": {
        "type": "string",
        "enum": ["admin", "developer", "guest"]
      }
    }
  }
}`

func TestParseAndGenerateTypeScript(t *testing.T) {
	spec, err := ParseOpenAPI([]byte(sampleOpenAPIJSON))
	if err != nil {
		t.Fatalf("failed to parse OpenAPI spec: %v", err)
	}

	files, err := GenerateTypeScript(spec)
	if err != nil {
		t.Fatalf("failed to generate TypeScript: %v", err)
	}

	expectedFiles := []string{"auth.ts", "database.ts", "storage.ts", "functions.ts", "api.ts", "index.ts"}
	for _, f := range expectedFiles {
		if _, exists := files[f]; !exists {
			t.Errorf("expected generated file %s to exist", f)
		}
	}

	// Check auth.ts content
	authCode := files["auth.ts"]
	if !strings.Contains(authCode, "export interface LoginRequest") {
		t.Errorf("auth.ts missing LoginRequest interface: %s", authCode)
	}
	if !strings.Contains(authCode, "email: string;") || !strings.Contains(authCode, "password: string;") {
		t.Errorf("auth.ts missing LoginRequest properties: %s", authCode)
	}
	if !strings.Contains(authCode, `export type UserRole = "admin" | "developer" | "guest";`) {
		t.Errorf("auth.ts missing UserRole enum type: %s", authCode)
	}

	// Check index.ts content
	indexCode := files["index.ts"]
	if !strings.Contains(indexCode, "export * from './auth';") {
		t.Errorf("index.ts missing re-exports: %s", indexCode)
	}
}

func TestFetchOpenAPISpec(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/docs/openapi.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleOpenAPIJSON))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient()
	data, err := client.FetchOpenAPISpec(server.URL)
	if err != nil {
		t.Fatalf("FetchOpenAPISpec failed: %v", err)
	}

	if !strings.Contains(string(data), "NodePhone API") {
		t.Errorf("unexpected spec content: %s", string(data))
	}
}
