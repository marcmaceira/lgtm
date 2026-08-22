package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"lgtm/internal/config"
)

func TestRunInitWritesEditableInstructionFields(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if code := runInit(nil); code != 0 {
		t.Fatalf("init returned %d", code)
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "lgtm", "config.json"))
	if err != nil {
		t.Fatal(err)
	}

	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"commitInstructions", "prInstructions"} {
		if value, ok := values[key]; !ok || value != "" {
			t.Fatalf("%s = %#v, present = %v; want an editable empty value", key, value, ok)
		}
	}
}

func TestRunConfigReplacesAndClearsCommitInstructions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if code := runConfig([]string{"--commit-instructions", "Use Conventional Commits"}); code != 0 {
		t.Fatalf("setting commit instructions returned %d", code)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.CommitInstructions; got != "Use Conventional Commits" {
		t.Fatalf("CommitInstructions = %q", got)
	}

	if code := runConfig([]string{"--commit-instructions", ""}); code != 0 {
		t.Fatalf("clearing commit instructions returned %d", code)
	}
	cfg, err = config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.CommitInstructions; got != "" {
		t.Fatalf("CommitInstructions = %q after clearing", got)
	}
}
