package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadValidConfig(t *testing.T) {
	yaml := `
commands:
  - name: get-user
    description: Get a user by ID
    method: GET
    url: https://api.example.com/users/{id}
    headers:
      Authorization: Bearer token123
    body: ""
    query_params:
      verbose: "true"
`
	cfg, err := Load(writeTempConfig(t, yaml))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(cfg.Commands))
	}
	c := cfg.Commands[0]
	if c.Name != "get-user" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Description != "Get a user by ID" {
		t.Errorf("description = %q", c.Description)
	}
	if c.Method != "GET" {
		t.Errorf("method = %q", c.Method)
	}
	if c.URL != "https://api.example.com/users/{id}" {
		t.Errorf("url = %q", c.URL)
	}
	if c.Headers["authorization"] != "Bearer token123" {
		t.Errorf("headers = %v", c.Headers)
	}
	if c.QueryParams["verbose"] != "true" {
		t.Errorf("query_params = %v", c.QueryParams)
	}
}

func TestLoadMinimalCommand(t *testing.T) {
	yaml := `
commands:
  - name: ping
    method: GET
    url: https://api.example.com/ping
`
	cfg, err := Load(writeTempConfig(t, yaml))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(cfg.Commands))
	}
	c := cfg.Commands[0]
	if c.Name != "ping" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Headers != nil {
		t.Errorf("headers = %v, want nil", c.Headers)
	}
	if c.QueryParams != nil {
		t.Errorf("query_params = %v, want nil", c.QueryParams)
	}
	if c.Body != "" {
		t.Errorf("body = %q, want empty", c.Body)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	path := writeTempConfig(t, `{{{invalid yaml:::`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	path := writeTempConfig(t, "")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Commands) != 0 {
		t.Errorf("got %d commands, want 0", len(cfg.Commands))
	}
}
