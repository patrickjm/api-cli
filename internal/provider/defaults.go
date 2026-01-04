package provider

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/patrickjm/api-cli/internal/config"
)

//go:embed defaults/*.js
var defaultProviders embed.FS

func EnsureDefaults(base string) error {
	entries, err := fs.Glob(defaultProviders, "defaults/*.js")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := strings.TrimSuffix(filepath.Base(entry), ".js")
		target := config.ProviderPath(base, name)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		contents, err := fs.ReadFile(defaultProviders, entry)
		if err != nil {
			return err
		}
		if err := SaveProvider(target, contents); err != nil {
			return err
		}
	}
	return nil
}
