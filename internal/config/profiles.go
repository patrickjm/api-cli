package config

import (
	"encoding/json"
	"errors"
	"os"
)

const DefaultProfile = "default"

type Profile struct {
	Secrets []string          `json:"secrets"`
	Env     map[string]string `json:"env,omitempty"`
}

type Profiles struct {
	Default  string             `json:"default"`
	Profiles map[string]Profile `json:"profiles"`
}

func LoadProfiles(path string) (*Profiles, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p := &Profiles{
				Default:  DefaultProfile,
				Profiles: map[string]Profile{DefaultProfile: {}},
			}
			return p, nil
		}
		return nil, err
	}
	var p Profiles
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if p.Profiles == nil {
		p.Profiles = map[string]Profile{}
	}
	if p.Default == "" {
		p.Default = DefaultProfile
	}
	if _, ok := p.Profiles[p.Default]; !ok {
		p.Profiles[p.Default] = Profile{}
	}
	return &p, nil
}

func SaveProfiles(path string, p *Profiles) error {
	if p.Default == "" {
		p.Default = DefaultProfile
	}
	if p.Profiles == nil {
		p.Profiles = map[string]Profile{}
	}
	if _, ok := p.Profiles[p.Default]; !ok {
		p.Profiles[p.Default] = Profile{}
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func ResolveProfile(p *Profiles, requested string) (string, error) {
	if requested != "" {
		if _, ok := p.Profiles[requested]; !ok {
			return "", errors.New("profile does not exist")
		}
		return requested, nil
	}
	if len(p.Profiles) == 1 {
		return p.Default, nil
	}
	return "", errors.New("multiple profiles exist; pass --profile")
}

func AddProfile(p *Profiles, name string) error {
	if name == "" {
		return errors.New("profile name is empty")
	}
	if p.Profiles == nil {
		p.Profiles = map[string]Profile{}
	}
	if _, ok := p.Profiles[name]; ok {
		return errors.New("profile already exists")
	}
	p.Profiles[name] = Profile{}
	return nil
}

func RemoveProfile(p *Profiles, name string) error {
	if name == "" {
		return errors.New("profile name is empty")
	}
	if name == p.Default {
		return errors.New("cannot remove default profile")
	}
	if _, ok := p.Profiles[name]; !ok {
		return errors.New("profile does not exist")
	}
	delete(p.Profiles, name)
	return nil
}

func UpsertSecret(p *Profiles, profile, key string) {
	if p.Profiles == nil {
		p.Profiles = map[string]Profile{}
	}
	prof := p.Profiles[profile]
	for _, existing := range prof.Secrets {
		if existing == key {
			p.Profiles[profile] = prof
			return
		}
	}
	prof.Secrets = append(prof.Secrets, key)
	p.Profiles[profile] = prof
}

func RemoveSecret(p *Profiles, profile, key string) {
	prof := p.Profiles[profile]
	if len(prof.Secrets) == 0 {
		return
	}
	out := prof.Secrets[:0]
	for _, existing := range prof.Secrets {
		if existing != key {
			out = append(out, existing)
		}
	}
	prof.Secrets = out
	p.Profiles[profile] = prof
}

func UpsertEnv(p *Profiles, profile, key, value string) {
	if p.Profiles == nil {
		p.Profiles = map[string]Profile{}
	}
	prof := p.Profiles[profile]
	if prof.Env == nil {
		prof.Env = map[string]string{}
	}
	prof.Env[key] = value
	p.Profiles[profile] = prof
}

func RemoveEnv(p *Profiles, profile, key string) {
	prof := p.Profiles[profile]
	if prof.Env == nil {
		return
	}
	delete(prof.Env, key)
	p.Profiles[profile] = prof
}
