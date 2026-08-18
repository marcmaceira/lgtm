package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"lgtm/internal/config"
	"lgtm/internal/gitctx"
	"lgtm/internal/llm"
	"lgtm/internal/prompt"
	"lgtm/internal/schema"
)

func runPR(args []string) int {
	fs := flag.NewFlagSet("lgtm pr", flag.ContinueOnError)
	backendFlag := fs.String("backend", "", "backend to use: claude or codex")
	modelFlag := fs.String("model", "", "model override")
	baseFlag := fs.String("base", "", "base branch (default: detected remote HEAD or config default)")
	draft := fs.Bool("draft", false, "open the PR as a draft")
	dryRun := fs.Bool("dry-run", false, "print the generated title/body without opening a PR")
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

	base := *baseFlag
	if base == "" {
		base = gitctx.DefaultRemoteBranch()
	}
	if base == "" {
		base = cfg.DefaultBase
	}

	head, err := gitctx.CurrentBranch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	if head == base {
		fmt.Fprintf(os.Stderr, "lgtm: current branch %q is the same as base %q\n", head, base)
		return 1
	}

	commitLog, err := gitctx.CommitLog(base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	diffStat, err := gitctx.RangeDiffStat(base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	diffPatch, err := gitctx.RangeDiffPatch(base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	configDir, _ := config.Dir()
	p, err := prompt.PR(configDir, prompt.PRData{
		Base:         base,
		Head:         head,
		CommitLog:    gitctx.Truncate(commitLog, cfg.Limits.PRSummary),
		DiffStat:     gitctx.Truncate(diffStat, cfg.Limits.PRDiffStat),
		DiffPatch:    gitctx.Truncate(diffPatch, cfg.Limits.PRPatch),
		RepoTemplate: gitctx.DetectPRTemplate(),
		Instructions: cfg.PRInstructions,
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

	fmt.Fprintf(os.Stderr, "lgtm: generating PR content with %s...\n", backend.Name())
	out, err := backend.Generate(context.Background(), p, schema.PR, model)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	var result schema.PRResult
	if err := json.Unmarshal(out, &result); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: unparseable model output:", err, string(out))
		return 1
	}

	if *dryRun {
		fmt.Printf("Title: %s\n\n%s\n", result.Title, result.Body)
		return 0
	}

	push := exec.Command("git", "push", "-u", "origin", head)
	push.Stdout = os.Stdout
	push.Stderr = os.Stderr
	if err := push.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: git push failed:", err)
		return 1
	}

	ghArgs := []string{"pr", "create", "--base", base, "--head", head, "--title", result.Title, "--body", result.Body}
	if *draft {
		ghArgs = append(ghArgs, "--draft")
	}
	gh := exec.Command("gh", ghArgs...)
	gh.Stdout = os.Stdout
	gh.Stderr = os.Stderr
	if err := gh.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: gh pr create failed:", err)
		return 1
	}
	return 0
}
