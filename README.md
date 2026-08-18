# lgtm

AI-generated commit messages and PR descriptions, driven by whatever
structured-output-capable coding CLI you already have installed and
authenticated — [Claude Code](https://claude.com/product/claude-code) or
[Codex CLI](https://github.com/openai/codex).

`lgtm` reads your staged diff (or the diff against a base branch for PRs),
renders a prompt, and asks the backend to return a small JSON object
(`{subject, body}` or `{title, body}`) — then runs the actual `git commit`
or `git push` + `gh pr create` itself. The model never touches your working
tree or executes commands.

## Install

Requires Go 1.26+, and either the `claude` or `codex` CLI on your `PATH`
(logged in). PR creation also requires the `gh` CLI, authenticated.

```sh
git clone <this repo> ~/Projects/lgtm
cd ~/Projects/lgtm
make install        # or: go install .
```

This installs the `lgtm` binary to `$(go env GOPATH)/bin` (typically
`~/go/bin`). Make sure that directory is on your `PATH`:

```sh
export PATH="$HOME/go/bin:$PATH"
```

## Usage

```sh
git add -A
lgtm commit                       # generates a message and commits
lgtm commit --dry-run             # just prints the message
lgtm commit --backend codex       # use Codex instead of Claude for this run

lgtm pr --base main               # generates title/body, pushes, opens a PR
lgtm pr --draft --dry-run         # preview without pushing or opening anything
```

Flags available on both `commit` and `pr`:

| Flag         | Description                                      |
|--------------|---------------------------------------------------|
| `--backend`  | `claude` or `codex` (overrides config default)     |
| `--model`    | Model name override                                |
| `--dry-run`  | Print the generated content, don't act on it       |

`pr` also takes `--base <branch>` (defaults to the repo's detected remote
`HEAD`, or `defaultBase` in config) and `--draft`.

## Configuration

`lgtm init` writes `~/.config/lgtm/config.json` plus editable prompt
templates (`commit.tmpl`, `pr.tmpl`) to the same directory.

```jsonc
{
  "backend": "claude",              // default backend: "claude" or "codex"
  "claude": { "model": "sonnet" },
  "codex": { "model": "" },         // "" lets codex use your account's default
  "defaultBase": "main",
  "commitInstructions": "",         // free-text rules appended to the commit prompt
  "prInstructions": "",             // free-text rules appended to the PR prompt
  "limits": {                       // char caps on injected context, to bound prompt size
    "commitSummary": 6000,
    "commitPatch": 40000,
    "prSummary": 20000,
    "prDiffStat": 20000,
    "prPatch": 60000
  }
}
```

Edit `~/.config/lgtm/commit.tmpl` or `pr.tmpl` to fully customize the prompt
(e.g. enforce Conventional Commits, change required PR sections). If a repo
has a `.github/PULL_REQUEST_TEMPLATE.md` (or the other common template
locations), `lgtm pr` detects it automatically and asks the model to follow
its structure instead of the default Summary/Testing sections.

## How it works

- **Claude backend**: `claude -p --output-format json --json-schema <schema> --tools ""`,
  prompt piped via stdin. `--tools ""` disables all tool access, so the call
  never needs `--dangerously-skip-permissions`.
- **Codex backend**: `codex exec --output-schema <file> --sandbox read-only -o <file>`,
  prompt piped via stdin, result read back from the output file.

Both backends are plain `exec.Command` calls — no SDKs, no network calls
made directly by `lgtm` itself, no external Go dependencies.

## Project layout

```
main.go                    entrypoint
internal/cli/               subcommands: commit, pr, init, config
internal/gitctx/            git context extraction (diffs, branches, PR templates)
internal/prompt/            prompt templates + rendering
internal/llm/               backend abstraction (claude, codex)
internal/schema/            JSON Schemas + result types
internal/config/            ~/.config/lgtm/config.json load/save
```
