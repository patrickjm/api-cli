package e2e

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCLIInsideContainer(t *testing.T) {
	ctx := context.Background()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	networkName := "api-cli-test-" + time.Now().UTC().Format("20060102150405.000000000")
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{Name: networkName},
	})
	if err != nil {
		t.Fatalf("create network: %v", err)
	}
	defer func() {
		_ = network.Remove(ctx)
	}()

	httpReq := testcontainers.ContainerRequest{
		Image:        "mccutchen/go-httpbin:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/get").WithPort("8080/tcp").WithStartupTimeout(20 * time.Second),
		Networks:     []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"httpbin"},
		},
	}
	httpbin, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: httpReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start httpbin: %v", err)
	}
	defer func() {
		_ = httpbin.Terminate(ctx)
	}()

	repoRoot := findRepoRoot(t)
	cliReq := testcontainers.ContainerRequest{
		Image:    "golang:1.24",
		Cmd:      []string{"sleep", "300"},
		Networks: []string{networkName},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(repoRoot, testcontainers.ContainerMountTarget("/repo")),
		),
		Env: map[string]string{
			"GOMODCACHE": "/tmp/gomodcache",
			"GOCACHE":    "/tmp/gocache",
			"GOPATH":     "/tmp/gopath",
		},
	}
	cli, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cliReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start cli container: %v", err)
	}
	defer func() {
		_ = cli.Terminate(ctx)
	}()

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
	if err := cli.CopyFileToContainer(ctx, scriptPath, "/tmp/provider.js", 0o644); err != nil {
		t.Fatalf("copy script: %v", err)
	}

	configDir := "/tmp/api-config"
	goBin := "/usr/local/go/bin/go"
	_ = execInContainer(t, ctx, cli, []string{"/bin/sh", "-lc", "mkdir -p " + configDir})
	_ = execInContainer(t, ctx, cli, []string{"/bin/sh", "-lc", "cd /repo && " + goBin + " run ./cmd/api -c " + configDir + " install /tmp/provider.js --name httpbin"})
	_ = execInContainer(t, ctx, cli, []string{"/bin/sh", "-lc", "cd /repo && " + goBin + " run ./cmd/api -c " + configDir + " env set httpbin BASE_URL http://httpbin:8080"})

	out := execInContainer(t, ctx, cli, []string{"/bin/sh", "-lc", "cd /repo && " + goBin + " run ./cmd/api -c " + configDir + " --json httpbin -s path=/get -s q=hello"})
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	args, ok := payload["args"].(map[string]any)
	if !ok {
		t.Fatalf("expected args object")
	}
	val := args["q"]
	switch v := val.(type) {
	case string:
		if v != "hello" {
			t.Fatalf("expected q=hello, got %v", v)
		}
	case []any:
		if len(v) == 0 || v[0] != "hello" {
			t.Fatalf("expected q=hello, got %v", v)
		}
	default:
		t.Fatalf("unexpected args type: %T", val)
	}
}

func execInContainer(t *testing.T, ctx context.Context, c testcontainers.Container, cmd []string) string {
	t.Helper()
	code, reader, err := c.Exec(ctx, cmd, tcexec.Multiplexed())
	if err != nil {
		t.Fatalf("exec error: %v", err)
	}
	data, _ := io.ReadAll(reader)
	out := strings.TrimSpace(string(data))
	if code != 0 {
		t.Fatalf("exec failed (%d): %s", code, out)
	}
	return out
}
