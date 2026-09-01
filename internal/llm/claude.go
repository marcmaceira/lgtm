package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type ClaudeBackend struct{}

func (ClaudeBackend) Name() string { return "claude" }

type claudeEnvelope struct {
	IsError          bool            `json:"is_error"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output"`
}

func (ClaudeBackend) Generate(ctx context.Context, prompt, jsonSchema, model string) ([]byte, error) {
	args := []string{
		"-p",
		"--output-format", "json",
		"--json-schema", jsonSchema,
		"--tools", "",
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, claudeRunError(err, stdout.Bytes(), stderr.Bytes())
	}

	var env claudeEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		return nil, fmt.Errorf("claude: unparseable response: %w", err)
	}
	if env.IsError {
		return nil, fmt.Errorf("claude: generation failed: %s", env.Result)
	}
	if len(env.StructuredOutput) > 0 {
		return env.StructuredOutput, nil
	}
	return []byte(env.Result), nil
}

func claudeRunError(runErr error, stdout, stderr []byte) error {
	var env claudeEnvelope
	if err := json.Unmarshal(stdout, &env); err == nil && env.IsError && strings.TrimSpace(env.Result) != "" {
		return fmt.Errorf("claude: generation failed: %s", strings.TrimSpace(env.Result))
	}

	detail := strings.TrimSpace(string(stderr))
	if detail == "" {
		detail = strings.TrimSpace(string(stdout))
	}
	if detail == "" {
		return fmt.Errorf("claude: %w", runErr)
	}
	return fmt.Errorf("claude: %w: %s", runErr, detail)
}
