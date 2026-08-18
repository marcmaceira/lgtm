package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"lgtm/internal/config"
	"lgtm/internal/gitctx"
	"lgtm/internal/llm"
	"lgtm/internal/prompt"
	"lgtm/internal/schema"
)

func runCommit(args []string) int {
	fs := flag.NewFlagSet("lgtm commit", flag.ContinueOnError)
	backendFlag := fs.String("backend", "", "backend to use: claude or codex")
	modelFlag := fs.String("model", "", "model override")
	dryRun := fs.Bool("dry-run", false, "print the generated message without committing")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: loading config:", err)
		return 1
	}
	backendName := cfg.Backend
	if *backendFlag != "" {
		backendName = *backendFlag
	}
	model := backendModel(cfg, backendName)
	if *modelFlag != "" {
		model = *modelFlag
	}

	staged, err := gitctx.HasStagedChanges()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	if !staged {
		fmt.Fprintln(os.Stderr, "lgtm: nothing staged — run `git add` first")
		return 1
	}

	branch, err := gitctx.CurrentBranch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	summary, err := gitctx.StagedSummary()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	patch, err := gitctx.StagedPatch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	configDir, _ := config.Dir()
	p, err := prompt.Commit(configDir, prompt.CommitData{
		Branch:        branch,
		StagedSummary: gitctx.Truncate(summary, cfg.Limits.CommitSummary),
		StagedPatch:   gitctx.Truncate(patch, cfg.Limits.CommitPatch),
		Instructions:  cfg.CommitInstructions,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: rendering prompt:", err)
		return 1
	}

	backend, err := llm.New(backendName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "lgtm: generating commit message with %s...\n", backend.Name())
	out, err := backend.Generate(context.Background(), p, schema.Commit, model)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	var result schema.CommitResult
	if err := json.Unmarshal(out, &result); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: unparseable model output:", err, string(out))
		return 1
	}

	message := result.Subject
	if strings.TrimSpace(result.Body) != "" {
		message += "\n\n" + result.Body
	}

	if *dryRun {
		fmt.Println(message)
		return 0
	}

	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: git commit failed:", err)
		return 1
	}
	return 0
}

func backendModel(cfg config.Config, backend string) string {
	switch backend {
	case "codex":
		return cfg.Codex.Model
	default:
		return cfg.Claude.Model
	}
}
