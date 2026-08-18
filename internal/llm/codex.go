package llm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CodexBackend struct{}

func (CodexBackend) Name() string { return "codex" }

func (CodexBackend) Generate(ctx context.Context, prompt, jsonSchema, model string) ([]byte, error) {
	schemaFile, err := os.CreateTemp("", "lgtm-schema-*.json")
	if err != nil {
		return nil, err
	}
	defer os.Remove(schemaFile.Name())
	if _, err := schemaFile.WriteString(jsonSchema); err != nil {
		schemaFile.Close()
		return nil, err
	}
	schemaFile.Close()

	outFile, err := os.CreateTemp("", "lgtm-out-*.json")
	if err != nil {
		return nil, err
	}
	defer os.Remove(outFile.Name())
	outFile.Close()

	args := []string{
		"exec",
		"--output-schema", schemaFile.Name(),
		"--sandbox", "read-only",
		"--skip-git-repo-check",
		"-o", outFile.Name(),
	}
	if model != "" {
		args = append(args, "-m", model)
	}
	args = append(args, "-")

	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("codex: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		return nil, fmt.Errorf("codex: reading output: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, fmt.Errorf("codex: produced no output: %s", strings.TrimSpace(stdout.String()))
	}
	return data, nil
}
