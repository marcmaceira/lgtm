// Package prompt renders the commit-message and PR-content prompt
// templates, allowing per-repo/per-user overrides on disk.
package prompt

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const defaultCommitTemplate = `You write concise git commit messages.
Return a JSON object with keys: subject, body.
Rules:
- subject must be imperative, <= 72 chars, and no trailing period
- body can be empty string or short bullet points
- capture the primary user-visible or developer-visible change
{{- if .Instructions}}
Additional instructions: {{.Instructions}}
{{- end}}

Branch: {{.Branch}}

Staged files:
{{.StagedSummary}}

Staged patch:
{{.StagedPatch}}
`

const defaultPRTemplate = `You write source control change request content.
Return a JSON object with keys: title, body.
Rules:
- title should be concise and specific
{{- if .RepoTemplate}}
- follow the repository change request template structure below: fill in
  its sections, drop HTML comments, keep its markdown structure
{{- else}}
- body must be markdown and include headings '## Summary' and '## Testing'
{{- end}}
{{- if .Instructions}}
Additional instructions: {{.Instructions}}
{{- end}}
{{- if .RepoTemplate}}

Repository change request template:
{{.RepoTemplate}}
{{- end}}

Base branch: {{.Base}}
Head branch: {{.Head}}

Commits:
{{.CommitLog}}

Diff stat:
{{.DiffStat}}

Diff patch:
{{.DiffPatch}}
`

type CommitData struct {
	Branch        string
	StagedSummary string
	StagedPatch   string
	Instructions  string
}

type PRData struct {
	Base         string
	Head         string
	CommitLog    string
	DiffStat     string
	DiffPatch    string
	RepoTemplate string
	Instructions string
}

// load returns the override template at configDir/name if present,
// otherwise falls back to def.
func load(configDir, name, def string) (string, error) {
	if configDir != "" {
		p := filepath.Join(configDir, name)
		if data, err := os.ReadFile(p); err == nil {
			return string(data), nil
		}
	}
	return def, nil
}

func render(tmplText string, data any) (string, error) {
	t, err := template.New("prompt").Parse(tmplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n") + "\n", nil
}

// Commit renders the commit-message prompt, using configDir/commit.tmpl
// as an override if present.
func Commit(configDir string, data CommitData) (string, error) {
	tmpl, err := load(configDir, "commit.tmpl", defaultCommitTemplate)
	if err != nil {
		return "", err
	}
	return render(tmpl, data)
}

// PR renders the PR-content prompt, using configDir/pr.tmpl as an
// override if present.
func PR(configDir string, data PRData) (string, error) {
	tmpl, err := load(configDir, "pr.tmpl", defaultPRTemplate)
	if err != nil {
		return "", err
	}
	return render(tmpl, data)
}

// WriteDefaults writes the built-in templates to configDir so the user has
// a starting point to customize (used by `lgtm init`).
func WriteDefaults(configDir string) error {
	if err := os.WriteFile(filepath.Join(configDir, "commit.tmpl"), []byte(defaultCommitTemplate), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, "pr.tmpl"), []byte(defaultPRTemplate), 0o644)
}
