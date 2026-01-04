package runtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type echoResponse struct {
	Method string            `json:"method"`
	Header map[string]string `json:"header"`
	Body   map[string]any    `json:"body"`
}

func TestExecuteFetchOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		resp := echoResponse{
			Method: r.Method,
			Header: map[string]string{
				"X-Test": r.Header.Get("X-Test"),
			},
			Body: body,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	script := `export default { default: { run: (params) => {
  return fetch(params.base + "/echo", { method: "POST", headers: { "X-Test": "ok" }, body: { value: params.value } });
} } }`

	res, err := Execute([]byte(script), ExecOptions{
		Provider: "test",
		Profile:  "default",
		Command:  "default",
		Params: map[string]string{
			"base":  server.URL,
			"value": "hello",
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	var payload echoResponse
	if err := json.Unmarshal([]byte(res.JSON), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Method != "POST" {
		t.Fatalf("expected POST, got %s", payload.Method)
	}
	if payload.Header["X-Test"] != "ok" {
		t.Fatalf("expected header to be set")
	}
	if payload.Body["value"] != "hello" {
		t.Fatalf("expected body value to be hello")
	}
}

func TestExecuteFetchUrlAndEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"path": r.URL.Path,
			"env":  r.Header.Get("X-Env"),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	script := `export default { default: { run: (params) => {
  return fetch(params.base + "/ping", { headers: { "X-Env": env("TOKEN") } });
} } }`

	res, err := Execute([]byte(script), ExecOptions{
		Provider: "test",
		Profile:  "default",
		Command:  "default",
		Params: map[string]string{
			"base": server.URL,
		},
		Env: map[string]string{
			"TOKEN": "abc123",
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(res.JSON), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["path"] != "/ping" {
		t.Fatalf("expected /ping, got %s", payload["path"])
	}
	if payload["env"] != "abc123" {
		t.Fatalf("expected env header")
	}
}

func TestExecuteFetchWithOptionsObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"method": r.Method}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	script := `export default { default: { run: (params) => {
  return fetch({ url: params.base + "/ping", method: "PUT" });
} } }`

	res, err := Execute([]byte(script), ExecOptions{
		Provider: "test",
		Profile:  "default",
		Command:  "default",
		Params: map[string]string{
			"base": server.URL,
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(res.JSON), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["method"] != "PUT" {
		t.Fatalf("expected PUT, got %s", payload["method"])
	}
}
