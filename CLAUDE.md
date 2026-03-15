# CLAUDE.md

This file provides instructions and context for Claude when working in this repository.

## Project Overview

`agent` is a zero-dependency Go CLI tool that uses AI assistants (Claude, Copilot) to review and fix code changes in a Git repository. It diffs the current branch against the default branch, runs specialized reviewers in parallel, outputs issues as JSON, and can automatically apply fixes.

## Build, Test, and Lint Commands

```sh
# Install the CLI
go install ./cmd/agent

# Run all tests (use -race in CI)
go test ./...
go test -race ./...

# Vet
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

### Reviewers

Each reviewer is an embedded Markdown file under `internal/reviewer/data/prompts/`. The filename (without `.md`) is the reviewer name.

| Name | Focus |
|------|-------|
| `architecture` | Structural and design concerns |
| `cli` | CLI interface and flag usage |
| `docs` | Documentation drift — `CLAUDE.md` out of sync with the code |
| `duplication` | Repeated or redundant code |
| `go` | Go idioms and best practices |
| `security` | Security vulnerabilities |
| `tests` | Test coverage and quality |
| `unused` | Dead code and unused declarations |

Run any combination with `--reviewers`:

```sh
agent review --reviewers go,security,docs --assistant claude
```

### Adding a New Reviewer

Create a Markdown file at `internal/reviewer/data/prompts/<name>.md`. It is automatically embedded and registered — no Go code changes are needed. The reviewer name is the filename without `.md`. See existing files for the prompt format.

### Adding a New AI Assistant

Implement the `Assistant` interface in `internal/assistant/`:

```go
type Assistant interface {
    Command(prompt string) *exec.Cmd
}
```

Then register it in the `assistantByName` map in `internal/assistant/assistant.go`. See `claude.go` and `copilot.go` for examples — each is a single struct with a `Command` method that returns an `*exec.Cmd`.

### Prompt Templates

- **Review prompt**: `internal/reviewer/data/prompt_template.md` — uses `{{prompt}}` and `{{diff}}` placeholders
- **Fix prompt**: `internal/fixer/data/prompt_template.md` — uses `{{issues}}` and `{{diff}}` placeholders

Templates are embedded with `//go:embed` and rendered with `strings.NewReplacer`.

### Parallelism

All reviewers run concurrently via `internal/x/sync.Parallel()` — a generic function using `sync.WaitGroup`. Results and errors are collected after all goroutines complete.

## Code Conventions

- **Zero external dependencies**: use only the standard library.
- **Logging**: use `log/slog` with debug level (`slog.Debug`). The `--verbose` flag enables debug output.
- **Error wrapping**: always use `fmt.Errorf("context: %w", err)` to preserve the error chain.
- **Embedding**: use `//go:embed` for all prompt and template files.
- **Interfaces at the consumer**: define interfaces (e.g., `assistant.Assistant`) in the package that uses them, not the implementing package.
- **No mutable globals**: package-level state (e.g., `reviewersByName`, `assistantByName`) is initialized once and never mutated.

## Testing Conventions

- Use Go's standard `testing` package — no third-party test frameworks.
- Tests are in `_test.go` files in the same package (white-box testing).
- Use **table-driven tests** (`tests := []struct{...}{}`) for parameterized cases.
- Mock the `assistant.Assistant` interface with a `fakeAssistant` struct:

```go
type fakeAssistant struct {
    fn func(string) *exec.Cmd
}

func (f *fakeAssistant) Command(prompt string) *exec.Cmd {
    return f.fn(prompt)
}

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }
```

- Use `t.Fatalf` (stops the test) for setup failures, `t.Errorf` (continues) for assertion failures.
- Avoid mocking `os/exec` globally — instead pass `*exec.Cmd` via the interface.

## CLI Design

- All flags are defined with `flag.FlagSet` per command, never as global `flag` package calls.
- Flag-parsing helpers live in `flag.go` files within each package (e.g., `internal/reviewer/flag.go`).
- Error messages are written to `os.Stderr`; structured output (JSON issues) goes to `os.Stdout`.
- `review` outputs JSON to stdout so it can be piped directly to `fix`.
