# Copilot Instructions

This repository contains `agent` — a zero-dependency Go CLI tool that uses AI assistants (Claude, Copilot) to review and fix code changes in a Git repository.

## Project Summary

`agent` diffs the current branch against the default branch, runs specialized reviewers in parallel, outputs issues as JSON to stdout, and can automatically apply fixes by feeding those issues back to an AI assistant.

## Build, Test, and Lint

```sh
go install ./cmd/agent   # build and install
go test -race ./...      # run all tests (race detector on in CI)
go vet ./...             # static analysis
go fmt ./...             # format (CI verifies no diff)
go mod tidy              # tidy deps (CI verifies no diff)
```

Requires Go 1.26+. **No external dependencies** — standard library only.

## Repository Layout

| Path | Purpose |
|------|---------|
| `cmd/agent/` | CLI entry point and command implementations (`review`, `fix`, `loop`, `help`) |
| `internal/assistant/` | `Assistant` interface + Claude and Copilot implementations |
| `internal/reviewer/` | Review engine + 7 embedded reviewer prompts |
| `internal/fixer/` | Fix engine with embedded prompt template |
| `internal/git/` | Git diff and branch helpers |
| `internal/x/sync/` | Generic `Parallel()` for concurrent fan-out |
| `internal/x/strings/` | `StripMarkdownFence` for cleaning LLM output |
| `internal/x/log/` | `--verbose` flag and `slog` debug-level helper |
| `internal/x/io/` | `PrefixWriter` for debug output |
| `internal/x/ux/` | CLI spinner |

## Key Extension Points

### Adding a Reviewer

Create a Markdown file at `internal/reviewer/data/prompts/<name>.md`. It is automatically embedded and registered — no Go code changes needed. The reviewer name equals the filename without `.md`.

### Adding an AI Assistant

Implement the `Assistant` interface and register it in `assistantByName`:

```go
// internal/assistant/assistant.go
type Assistant interface {
    Command(prompt string) *exec.Cmd
}
var assistantByName = map[string]Assistant{
    "claude":  &Claude{},
    "copilot": &Copilot{},
    // add new entry here
}
```

See `claude.go` and `copilot.go` for minimal one-method implementations.

## Code Style and Conventions

- **No external dependencies** — stdlib only; do not add third-party modules.
- **Error wrapping**: `fmt.Errorf("context: %w", err)` to preserve the error chain.
- **Logging**: `log/slog` with `slog.Debug`/`slog.Info`. Debug output is gated behind `--verbose`.
- **Embedding**: use `//go:embed` for all prompt templates and static assets.
- **Interfaces in consuming packages**: define interfaces where they are used, not in the implementing package.
- **CLI flags**: always use per-command `flag.FlagSet`; never use the global `flag` package.
- **JSON flow**: `review` encodes issues to stdout; `fix` decodes issues from stdin — they pipe together.

## Testing Conventions

- Standard library `testing` only.
- Table-driven tests for multiple cases.
- Mock `assistant.Assistant` with a local `fakeAssistant` struct:

```go
type fakeAssistant struct{ fn func(string) *exec.Cmd }
func (f *fakeAssistant) Command(p string) *exec.Cmd { return f.fn(p) }

func echoCmd(out string) *exec.Cmd { return exec.Command("echo", out) }
func failCmd() *exec.Cmd          { return exec.Command("false") }
```

- `t.Fatalf` for fatal setup failures; `t.Errorf` for non-fatal assertions.
