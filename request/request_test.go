package request

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"nabr/config"
)

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want []string
	}{
		{"single param", "/users/{id}", []string{"id"}},
		{"multiple params", "/users/{userId}/posts/{postId}", []string{"userId", "postId"}},
		{"no params", "/users/all", nil},
		{"empty string", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPathParams(tt.url)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("param[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		raw  bool
		want string
	}{
		{
			"pretty print",
			[]byte(`{"key":"value"}`),
			false,
			"{\n  \"key\": \"value\"\n}",
		},
		{
			"raw passthrough",
			[]byte(`{"key":"value"}`),
			true,
			`{"key":"value"}`,
		},
		{
			"invalid JSON fallback",
			[]byte(`not json`),
			false,
			"not json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatJSON(tt.body, tt.raw)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// echoHandler returns a JSON response containing the request details.
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

func TestExecute(t *testing.T) {
	srv := httptest.NewServer(echoHandler())
	defer srv.Close()

	t.Run("basic GET", func(t *testing.T) {
		cmd := config.Command{Method: "GET", URL: srv.URL + "/ping"}
		resp, err := Execute(cmd, nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		var echo map[string]interface{}
		json.Unmarshal(resp.Body, &echo)
		if echo["method"] != "GET" {
			t.Errorf("method = %v, want GET", echo["method"])
		}
	})

	t.Run("path param substitution", func(t *testing.T) {
		cmd := config.Command{Method: "GET", URL: srv.URL + "/users/{id}"}
		resp, err := Execute(cmd, map[string]string{"id": "42"})
		if err != nil {
			t.Fatal(err)
		}
		var echo map[string]interface{}
		json.Unmarshal(resp.Body, &echo)
		if echo["path"] != "/users/42" {
			t.Errorf("path = %v, want /users/42", echo["path"])
		}
	})

	t.Run("headers applied", func(t *testing.T) {
		cmd := config.Command{
			Method:  "GET",
			URL:     srv.URL + "/test",
			Headers: map[string]string{"X-Custom": "hello"},
		}
		resp, err := Execute(cmd, nil)
		if err != nil {
			t.Fatal(err)
		}
		var echo map[string]interface{}
		json.Unmarshal(resp.Body, &echo)
		headers := echo["headers"].(map[string]interface{})
		vals := headers["X-Custom"].([]interface{})
		if vals[0] != "hello" {
			t.Errorf("X-Custom = %v, want hello", vals[0])
		}
	})

	t.Run("query params applied", func(t *testing.T) {
		cmd := config.Command{
			Method:      "GET",
			URL:         srv.URL + "/search",
			QueryParams: map[string]string{"q": "test", "page": "1"},
		}
		resp, err := Execute(cmd, nil)
		if err != nil {
			t.Fatal(err)
		}
		var echo map[string]interface{}
		json.Unmarshal(resp.Body, &echo)
		query := echo["query"].(map[string]interface{})
		qVals := query["q"].([]interface{})
		if qVals[0] != "test" {
			t.Errorf("q = %v, want test", qVals[0])
		}
	})

	t.Run("body sent", func(t *testing.T) {
		cmd := config.Command{
			Method: "POST",
			URL:    srv.URL + "/data",
			Body:   `{"name":"nabr"}`,
		}
		resp, err := Execute(cmd, nil)
		if err != nil {
			t.Fatal(err)
		}
		var echo map[string]interface{}
		json.Unmarshal(resp.Body, &echo)
		if echo["body"] != `{"name":"nabr"}` {
			t.Errorf("body = %v, want {\"name\":\"nabr\"}", echo["body"])
		}
	})

	t.Run("combined request", func(t *testing.T) {
		cmd := config.Command{
			Method:      "POST",
			URL:         srv.URL + "/users/{id}/posts",
			Headers:     map[string]string{"Authorization": "Bearer tok"},
			Body:        `{"title":"hi"}`,
			QueryParams: map[string]string{"draft": "true"},
		}
		resp, err := Execute(cmd, map[string]string{"id": "7"})
		if err != nil {
			t.Fatal(err)
		}
		var echo map[string]interface{}
		json.Unmarshal(resp.Body, &echo)

		if echo["method"] != "POST" {
			t.Errorf("method = %v, want POST", echo["method"])
		}
		if echo["path"] != "/users/7/posts" {
			t.Errorf("path = %v, want /users/7/posts", echo["path"])
		}
		if echo["body"] != `{"title":"hi"}` {
			t.Errorf("body = %v", echo["body"])
		}
		headers := echo["headers"].(map[string]interface{})
		auth := headers["Authorization"].([]interface{})
		if auth[0] != "Bearer tok" {
			t.Errorf("Authorization = %v", auth[0])
		}
		query := echo["query"].(map[string]interface{})
		draft := query["draft"].([]interface{})
		if draft[0] != "true" {
			t.Errorf("draft = %v", draft[0])
		}
	})

	t.Run("connection refused", func(t *testing.T) {
		cmd := config.Command{Method: "GET", URL: "http://127.0.0.1:1/nope"}
		_, err := Execute(cmd, nil)
		if err == nil {
			t.Fatal("expected error for connection refused")
		}
	})
}
