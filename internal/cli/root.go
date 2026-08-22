// Package cli implements lgtm's subcommands.
package cli

import (
	"fmt"
	"os"
)

const usage = `lgtm - AI-generated commit messages and PR descriptions

Usage:
  lgtm commit [flags]     Generate a commit message from staged changes and commit
  lgtm pr [flags]         Generate a PR title/body from the commit range and open it
  lgtm init               Write default config + prompt templates to ~/.config/lgtm
  lgtm config [flags]     Print or update the effective configuration

Global flags (commit/pr):
  --backend <claude|codex>   Override the configured backend
  --model <name>              Override the configured model
  --dry-run                   Print the generated content without acting on it

Config flags:
  --commit-instructions <text>  Replace commit instructions (empty clears them)
  --pr-instructions <text>      Replace PR instructions (empty clears them)
`

func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}

	switch args[0] {
	case "commit":
		return runCommit(args[1:])
	case "pr":
		return runPR(args[1:])
	case "init":
		return runInit(args[1:])
	case "config":
		return runConfig(args[1:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "lgtm: unknown command %q\n\n%s", args[0], usage)
		return 1
	}
}
