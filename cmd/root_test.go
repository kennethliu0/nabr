package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		input   string
		wantK   string
		wantV   string
		wantOk  bool
	}{
		{"foo=bar", "foo", "bar", true},
		{"key=val=ue", "key", "val=ue", true},
		{"key=", "key", "", true},
		{"=value", "", "", false},
		{"noequals", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			k, v, ok := parseKeyValue(tt.input)
			if ok != tt.wantOk || k != tt.wantK || v != tt.wantV {
				t.Errorf("parseKeyValue(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.input, k, v, ok, tt.wantK, tt.wantV, tt.wantOk)
			}
		})
	}
}

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func echoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		resp := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   r.URL.Query(),
			"headers": r.Header,
			"body":    string(body),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func TestRegisterCommand(t *testing.T) {
	srv := httptest.NewServer(echoHandler())
	defer srv.Close()

	t.Run("path param flags exist and are required", func(t *testing.T) {
		yaml := `
commands:
  - name: get-user
    method: GET
    url: ` + srv.URL + `/users/{id}
`
		root := newRootCmd(writeTestConfig(t, yaml))
		// Find the subcommand
		sub, _, err := root.Find([]string{"get-user"})
		if err != nil {
			t.Fatal(err)
		}
		f := sub.Flags().Lookup("id")
		if f == nil {
			t.Fatal("expected --id flag")
		}
		// Required flags have an annotation
		ann := f.Annotations
		if _, ok := ann["cobra_annotation_bash_completion_one_required_flag"]; !ok {
			t.Error("expected --id to be required")
		}
	})

	t.Run("path param substitution end-to-end", func(t *testing.T) {
		yaml := `
commands:
  - name: get-user
    method: GET
    url: ` + srv.URL + `/users/{id}
`
		root := newRootCmd(writeTestConfig(t, yaml))
		root.SetArgs([]string{"get-user", "--id", "42"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("query params merge and override", func(t *testing.T) {
		yaml := `
commands:
  - name: search
    method: GET
    url: ` + srv.URL + `/search
    query_params:
      page: "1"
      limit: "10"
`
		root := newRootCmd(writeTestConfig(t, yaml))
		root.SetArgs([]string{"search", "-q", "page=2", "-q", "extra=yes"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("headers merge and override", func(t *testing.T) {
		yaml := `
commands:
  - name: authed
    method: GET
    url: ` + srv.URL + `/test
    headers:
      Authorization: Bearer old
`
		root := newRootCmd(writeTestConfig(t, yaml))
		root.SetArgs([]string{"authed", "-H", "Authorization=Bearer new", "-H", "X-Extra=val"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("body override", func(t *testing.T) {
		yaml := `
commands:
  - name: create
    method: POST
    url: ` + srv.URL + `/data
    body: '{"old":"body"}'
`
		root := newRootCmd(writeTestConfig(t, yaml))
		root.SetArgs([]string{"create", "-b", `{"new":"body"}`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid query param format", func(t *testing.T) {
		yaml := `
commands:
  - name: bad
    method: GET
    url: ` + srv.URL + `/test
`
		root := newRootCmd(writeTestConfig(t, yaml))
		root.SetArgs([]string{"bad", "-q", "badformat"})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error for invalid query param format")
		}
	})
}
