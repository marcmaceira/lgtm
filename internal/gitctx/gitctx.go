// Package gitctx extracts the git context (diffs, branches, templates)
// that gets injected into generation prompts.
package gitctx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Truncate caps s to n characters, appending a marker if it was cut.
func Truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + fmt.Sprintf("\n... [truncated, %d more chars]", len(s)-n)
}

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out.String(), nil
}

// CurrentBranch returns the current branch name, or "(detached)" if in a
// detached HEAD state.
func CurrentBranch() (string, error) {
	out, err := run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(out)
	if branch == "HEAD" {
		return "(detached)", nil
	}
	return branch, nil
}

// HasStagedChanges reports whether there is anything staged for commit.
func HasStagedChanges() (bool, error) {
	out, err := run("diff", "--staged", "--name-only")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// StagedSummary returns `git diff --staged --stat`.
func StagedSummary() (string, error) {
	return run("diff", "--staged", "--stat")
}

// StagedPatch returns `git diff --staged`.
func StagedPatch() (string, error) {
	return run("diff", "--staged")
}

// DefaultRemoteBranch tries to detect the remote's default branch (e.g.
// "main"), falling back to "" if it can't be determined.
func DefaultRemoteBranch() string {
	out, err := run("symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return ""
	}
	ref := strings.TrimSpace(out)
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

// CommitLog returns `git log base..HEAD --oneline`.
func CommitLog(base string) (string, error) {
	return run("log", base+"..HEAD", "--oneline")
}

// RangeDiffStat returns `git diff base..HEAD --stat`.
func RangeDiffStat(base string) (string, error) {
	return run("diff", base+"..HEAD", "--stat")
}

// RangeDiffPatch returns `git diff base..HEAD`.
func RangeDiffPatch(base string) (string, error) {
	return run("diff", base+"..HEAD")
}

// RepoRoot returns the top-level directory of the current git repo.
func RepoRoot() (string, error) {
	out, err := run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

var prTemplateCandidates = []string{
	".github/PULL_REQUEST_TEMPLATE.md",
	".github/pull_request_template.md",
	"docs/PULL_REQUEST_TEMPLATE.md",
	"PULL_REQUEST_TEMPLATE.md",
	".gitlab/merge_request_templates/Default.md",
}

// DetectPRTemplate looks for a repo-defined PR/MR template and returns its
// contents, or "" if none is found.
func DetectPRTemplate() string {
	root, err := RepoRoot()
	if err != nil {
		return ""
	}
	for _, rel := range prTemplateCandidates {
		p := filepath.Join(root, rel)
		if data, err := os.ReadFile(p); err == nil {
			return string(data)
		}
	}
	return ""
}
