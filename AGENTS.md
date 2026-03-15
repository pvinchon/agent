# AGENTS.md

This file provides instructions and context for OpenAI Codex and other AI coding agents working in this repository.

## Project Overview

`agent` is a zero-dependency Go CLI tool that uses AI assistants (Claude, Copilot) to review and fix code changes in a Git repository. It diffs the current branch against the default branch, runs specialized reviewers in parallel, outputs issues as JSON, and can automatically apply fixes.

## Build, Test, and Lint Commands

```sh
# Install the CLI
go install ./cmd/agent

# Run all tests
go test ./...
go test -race ./...    # used in CI

# Vet (must pass in CI)
go vet ./...

# Format (must produce no diff in CI)
go fmt ./...

# Tidy dependencies (must produce no diff in CI)
go mod tidy
```

Requires Go 1.26+. There are **no external dependencies** — standard library only.

## Repository Structure

```
cmd/agent/          Entry point + command implementations (review, fix, loop, help)
internal/
  assistant/        AI assistant abstraction: Assistant interface, Claude, Copilot
  reviewer/         Code review engine + 7 embedded reviewer prompts
  fixer/            Automated fix applicator
  git/              Git operations (diff, branch detection)
  x/
    io/             PrefixWriter utility
    log/            slog debug-level helper + --verbose flag
    strings/        StripMarkdownFence for LLM output
    sync/           Generic Parallel() for concurrent fan-out
    ux/             CLI spinner
```

## Architecture

### Adding a New Reviewer

Reviewers are loaded automatically from embedded Markdown files. To add one, create a new file:

```
internal/reviewer/data/prompts/<name>.md
```

The file becomes the `Focus` section of the review prompt. No Go code changes needed. The reviewer name is the filename without `.md`. Existing files in that directory show the expected format.

### Adding a New AI Assistant

Implement the `Assistant` interface in `internal/assistant/`:

```go
type Assistant interface {
    Command(prompt string) *exec.Cmd
}
```

Then register the new type in the `assistantByName` map in `internal/assistant/assistant.go`. See `claude.go` and `copilot.go` for minimal examples.

### Prompt Templates

- `internal/reviewer/data/prompt_template.md` — review prompt, uses `{{prompt}}` and `{{diff}}` placeholders
- `internal/fixer/data/prompt_template.md` — fix prompt, uses `{{issues}}` and `{{diff}}` placeholders

Placeholders are replaced with `strings.NewReplacer`. Templates are embedded with `//go:embed`.

### Parallelism

Reviewers run concurrently via `internal/x/sync.Parallel()` — a generic fan-out using `sync.WaitGroup`. Results and errors are collected after all goroutines complete.

## Code Conventions

- **Zero external dependencies**: standard library only. Do not add third-party modules.
- **Logging**: use `log/slog` (`slog.Debug`, `slog.Info`). The `--verbose` flag enables debug output.
- **Error wrapping**: `fmt.Errorf("context: %w", err)` — always wrap to preserve the chain.
- **Embedding**: use `//go:embed` for all prompt and template files.
- **Interface placement**: define interfaces in the consuming package, not the implementing package.
- **No mutable package-level state**: package-level maps/vars are initialized once at startup.
- **CLI flags**: use `flag.FlagSet` per command; never call the global `flag` package directly.

## Testing Conventions

- Standard library `testing` package only — no third-party frameworks.
- Tests in `_test.go` files in the same package (white-box testing).
- Table-driven tests for parameterized cases.
- Mock `assistant.Assistant` using a `fakeAssistant` struct with a function field:

```go
type fakeAssistant struct {
    fn func(string) *exec.Cmd
}
func (f *fakeAssistant) Command(prompt string) *exec.Cmd { return f.fn(prompt) }

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }
```

- `t.Fatalf` for setup failures (stops the test); `t.Errorf` for assertion failures (continues).
- Run `go test -race ./...` to detect data races before submitting changes.

## CLI Design

- Subcommands: `review`, `fix`, `loop`, `help`.
- `review` writes JSON issues to stdout; `fix` reads JSON issues from stdin.
- They are designed to be piped: `agent review ... | agent fix ...`
- Error messages go to `os.Stderr`; structured output goes to `os.Stdout`.
