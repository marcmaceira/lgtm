// Package config loads and saves lgtm's user configuration.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type BackendConfig struct {
	Model string `json:"model,omitempty"`
}

type Config struct {
	// Backend selects which CLI to shell out to: "claude" or "codex".
	Backend string `json:"backend"`

	Claude BackendConfig `json:"claude"`
	Codex  BackendConfig `json:"codex"`

	// CommitInstructions/PRInstructions are free-text style rules appended
	// to the generation prompts (e.g. "use Conventional Commits").
	CommitInstructions string `json:"commitInstructions,omitempty"`
	PRInstructions     string `json:"prInstructions,omitempty"`

	// DefaultBase is the base branch used for `lgtm pr` when not overridden
	// with --base and no remote HEAD can be detected.
	DefaultBase string `json:"defaultBase,omitempty"`

	// Size caps (in characters) for context sections injected into prompts.
	Limits Limits `json:"limits"`
}

type Limits struct {
	CommitSummary int `json:"commitSummary"`
	CommitPatch   int `json:"commitPatch"`
	PRSummary     int `json:"prSummary"`
	PRDiffStat    int `json:"prDiffStat"`
	PRPatch       int `json:"prPatch"`
}

func Default() Config {
	return Config{
		Backend:     "claude",
		Claude:      BackendConfig{Model: "sonnet"},
		Codex:       BackendConfig{Model: "gpt-5-codex"},
		DefaultBase: "main",
		Limits: Limits{
			CommitSummary: 6000,
			CommitPatch:   40000,
			PRSummary:     20000,
			PRDiffStat:    20000,
			PRPatch:       60000,
		},
	}
}

// Dir returns ~/.config/lgtm, creating it if necessary.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "lgtm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads ~/.config/lgtm/config.json, falling back to defaults for any
// field that is absent or if the file doesn't exist yet.
func Load() (Config, error) {
	cfg := Default()
	p, err := path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes cfg to ~/.config/lgtm/config.json.
func Save(cfg Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
