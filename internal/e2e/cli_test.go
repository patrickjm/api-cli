package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCLIWithHTTPBin(t *testing.T) {
	ctx := context.Background()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	req := testcontainers.ContainerRequest{
		Image:        "mccutchen/go-httpbin:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/get").WithPort("8080/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start container: %v", err)
	}
	defer func() {
		_ = container.Terminate(ctx)
	}()

	endpoint, err := container.Endpoint(ctx, "http")
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}

	bin := buildBinary(t)
	configDir := t.TempDir()

	scriptPath := filepath.Join(t.TempDir(), "provider.js")
	script := `export default {
  default: {
    run: (params) => {
      const base = env("BASE_URL");
      return fetch(base + params.path + "?q=" + params.q);
    },
  },
};`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	runCmd(t, bin, "-c", configDir, "install", scriptPath, "--name", "httpbin")
	runCmd(t, bin, "-c", configDir, "env", "set", "httpbin", "BASE_URL", endpoint)

	out := runCmd(t, bin, "-c", configDir, "--json", "httpbin", "-s", "path=/get", "-s", "q=hello")
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	args, ok := payload["args"].(map[string]any)
	if !ok {
		t.Fatalf("expected args object")
	}
	switch v := args["q"].(type) {
	case string:
		if v != "hello" {
			t.Fatalf("expected q=hello, got %v", v)
		}
	case []any:
		if len(v) == 0 || v[0] != "hello" {
			t.Fatalf("expected q=hello, got %v", v)
		}
	default:
		t.Fatalf("unexpected args type: %T", args["q"])
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	bin := filepath.Join(t.TempDir(), "api")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/api")
	cmd.Dir = repoRoot
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build error: %v\n%s", err, string(out))
	}
	return bin
}

func runCmd(t *testing.T, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, string(out))
	}
	return string(out)
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("repo root not found")
	return ""
}
