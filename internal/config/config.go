package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	ProvidersDirName = "providers"
	ProfilesDirName  = "profiles"
)

func BaseDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if env := os.Getenv("API_CONFIG_DIR"); env != "" {
		return env, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "api"), nil
}

func EnsureLayout(base string) error {
	if base == "" {
		return errors.New("config dir is empty")
	}
	if err := os.MkdirAll(filepath.Join(base, ProvidersDirName), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(base, ProfilesDirName), 0o755); err != nil {
		return err
	}
	return nil
}

func ProvidersDir(base string) string {
	return filepath.Join(base, ProvidersDirName)
}

func ProfilesDir(base string) string {
	return filepath.Join(base, ProfilesDirName)
}

func ProviderPath(base, provider string) string {
	return filepath.Join(ProvidersDir(base), provider+".js")
}

func ProviderProfilesPath(base, provider string) string {
	return filepath.Join(ProfilesDir(base), provider+".json")
}
