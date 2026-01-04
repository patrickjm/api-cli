package config

import "testing"

func TestProfilesDefaults(t *testing.T) {
	profiles, err := LoadProfiles("/nonexistent")
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}
	if profiles.Default != DefaultProfile {
		t.Fatalf("expected default profile %q, got %q", DefaultProfile, profiles.Default)
	}
	if _, ok := profiles.Profiles[DefaultProfile]; !ok {
		t.Fatalf("expected default profile entry")
	}
}

func TestProfilesResolve(t *testing.T) {
	profiles := &Profiles{Default: DefaultProfile, Profiles: map[string]Profile{DefaultProfile: {}}}
	resolved, err := ResolveProfile(profiles, "")
	if err != nil {
		t.Fatalf("ResolveProfile error: %v", err)
	}
	if resolved != DefaultProfile {
		t.Fatalf("expected %q, got %q", DefaultProfile, resolved)
	}
	if err := AddProfile(profiles, "work"); err != nil {
		t.Fatalf("AddProfile error: %v", err)
	}
	if _, err := ResolveProfile(profiles, ""); err == nil {
		t.Fatalf("expected error when multiple profiles exist")
	}
}

func TestProfilesEnv(t *testing.T) {
	profiles := &Profiles{Default: DefaultProfile, Profiles: map[string]Profile{DefaultProfile: {}}}
	UpsertEnv(profiles, DefaultProfile, "BASE_URL", "https://example.com")
	if profiles.Profiles[DefaultProfile].Env["BASE_URL"] != "https://example.com" {
		t.Fatalf("env value not set")
	}
	RemoveEnv(profiles, DefaultProfile, "BASE_URL")
	if _, ok := profiles.Profiles[DefaultProfile].Env["BASE_URL"]; ok {
		t.Fatalf("env value not removed")
	}
}
