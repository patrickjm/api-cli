package provider

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ReadSource(source string, stdin io.Reader) ([]byte, error) {
	if source == "" || source == "-" {
		return io.ReadAll(stdin)
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, errors.New("failed to download provider")
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}

func SaveProvider(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, contents, 0o644)
}

func ListProviders(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".js") {
			out = append(out, strings.TrimSuffix(name, ".js"))
		}
	}
	return out, nil
}
