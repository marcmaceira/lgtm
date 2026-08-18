# lgtm — agent guidance

Guidance for AI coding agents working in this repository.

## Commands

```sh
make build      # go build -o bin/lgtm .
make install    # go install . -> $(go env GOPATH)/bin/lgtm
make fmt        # gofmt -l . (reports only; run `gofmt -w .` to fix)
make vet        # go vet ./...
make test       # go test ./...
```

No tests exist yet. Once added, run a single one with `go test ./internal/<pkg> -run '<TestName>' -v`.

Go 1.26, zero external dependencies (`go.mod` has no `require` block) — keep it that way; everything is stdlib plus `exec.Command`.

## Architecture

`lgtm` is a CLI that generates commit messages and PR descriptions by shelling out to an
already-installed coding CLI (`claude` or `codex`) in structured-output mode. The pipeline is
the same for both subcommands:

1. `internal/cli` parses flags, loads config, resolves backend + model (flag > config).
2. `internal/gitctx` shells out to `git` to collect context (staged diff/stat, `base..HEAD`
   range diff, branch names, repo PR template).
3. `internal/prompt` renders that context into a `text/template` prompt.
4. `internal/llm` pipes the prompt to `claude`/`codex` via stdin with a JSON Schema from
   `internal/schema`, and returns raw JSON bytes.
5. `internal/cli` unmarshals into `schema.CommitResult`/`schema.PRResult` and runs the
   real side-effecting commands itself: `git commit`, or `git push` + `gh pr create`.

Key invariant: **the model never touches the working tree or executes commands.** The Claude
backend passes `--tools ""` and the Codex backend passes `--sandbox read-only`, so neither
needs permission bypass flags. Preserve this when editing `internal/llm`.

### Backends

`llm.Backend` is a 2-method interface (`Name`, `Generate`). The two implementations differ in
how structured output comes back:

- `claude.go`: `claude -p --output-format json --json-schema <inline schema> --tools ""`.
  Reads stdout, unwraps the `claudeEnvelope` (`is_error`, `result`, `structured_output`),
  preferring `structured_output` and falling back to `result`.
- `codex.go`: `codex exec --output-schema <tempfile> -o <tempfile> --sandbox read-only -`.
  Schema and result both go through temp files; the result is read back from disk, not stdout.

Adding a backend means implementing `Backend`, registering it in `llm.New`, and adding a
`BackendConfig` field + `backendModel()` case (`internal/cli/commit.go`) for its model default.

### Configuration and prompt overrides

`~/.config/lgtm/` holds `config.json` plus optional `commit.tmpl` / `pr.tmpl`. `config.Load()`
starts from `config.Default()` and unmarshals the file over it, so missing fields fall back to
defaults rather than zero values — add new config fields to `Default()` too.

`prompt.load()` reads `configDir/<name>.tmpl` if it exists and otherwise uses the built-in
constant in `internal/prompt/templates.go`. The built-ins are the source of truth that
`lgtm init` writes to disk, so template changes there only affect users who haven't
customized (or who re-run `init` after deleting their copies). Any new field added to
`CommitData`/`PRData` must also be referenced in the built-in template to have an effect.

Prompt size is bounded by `config.Limits` (per-section character caps) applied via
`gitctx.Truncate` at the call site in `internal/cli`, not inside `gitctx`.

### CLI conventions

Dispatch is a hand-rolled `switch` in `internal/cli/root.go` — no cobra. Each subcommand is a
`run*(args []string) int` returning the process exit code; `main.go` passes it to `os.Exit`.
Note `runConfig` lives in `internal/cli/init.go`, not a `config.go`.
Errors go to stderr prefixed `lgtm: `, and generated content goes to stdout so `--dry-run`
output is pipeable. When adding a subcommand, update both the `switch` and the `usage` const.
