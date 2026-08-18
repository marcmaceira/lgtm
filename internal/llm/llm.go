// Package llm shells out to a structured-output-capable CLI (claude or
// codex) to turn a prompt + JSON schema into a validated JSON result.
package llm

import (
	"context"
	"fmt"
)

// Backend generates structured JSON output from a prompt, constrained by
// the given JSON Schema. model may be empty to use the backend's default.
type Backend interface {
	Name() string
	Generate(ctx context.Context, prompt, jsonSchema, model string) ([]byte, error)
}

func New(name string) (Backend, error) {
	switch name {
	case "claude":
		return ClaudeBackend{}, nil
	case "codex":
		return CodexBackend{}, nil
	default:
		return nil, fmt.Errorf("unknown backend %q (want \"claude\" or \"codex\")", name)
	}
}
