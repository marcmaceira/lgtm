package llm

import (
	"errors"
	"testing"
)

func TestClaudeRunErrorUsesJSONErrorResult(t *testing.T) {
	stdout := []byte(`{"is_error":true,"result":"Not logged in · Please run /login"}`)

	err := claudeRunError(errors.New("exit status 1"), stdout, nil)

	const want = "claude: generation failed: Not logged in · Please run /login"
	if err.Error() != want {
		t.Fatalf("claudeRunError() = %q, want %q", err, want)
	}
}

func TestClaudeRunErrorFallsBackToStderr(t *testing.T) {
	err := claudeRunError(errors.New("exit status 1"), nil, []byte("request failed\n"))

	const want = "claude: exit status 1: request failed"
	if err.Error() != want {
		t.Fatalf("claudeRunError() = %q, want %q", err, want)
	}
}

func TestClaudeRunErrorFallsBackToRawStdout(t *testing.T) {
	err := claudeRunError(errors.New("exit status 1"), []byte("unexpected failure\n"), nil)

	const want = "claude: exit status 1: unexpected failure"
	if err.Error() != want {
		t.Fatalf("claudeRunError() = %q, want %q", err, want)
	}
}
