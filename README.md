# agent

A CLI tool that uses AI assistants to review and fix code changes in a Git repository.

## How it works

`agent` diffs your current branch against the default branch (`main`/`master`), runs one or more specialized reviewers against that diff in parallel, and outputs a list of issues as JSON. Issues can then be fed back to the assistant to apply fixes automatically.

## Installation

```sh
go install ./cmd/agent
```

Requires Go 1.26+.

## Commands

### `review`

Runs the diff through the selected reviewers and prints issues as JSON to stdout.

```sh
agent review --reviewers <list> --assistant <name>
```

```sh
agent review --reviewers go,security --assistant claude | tee issues.json
```

### `fix`

Reads issues as JSON from stdin and applies fixes to the working tree.

```sh
agent fix --assistant <name> < issues.json
```

```sh
agent fix --assistant copilot < issues.json
```

### `loop`

Runs review and fix in a loop until no issues remain or the maximum number of attempts is reached.

```sh
agent loop --reviewers <list> --review-assistant <name> --fix-assistant <name> [--max-attempts <n>]
```

```sh
agent loop --reviewers go,security,tests --review-assistant claude --fix-assistant copilot --max-attempts 3
```

### `help`

Shows usage for a command.

```sh
agent help <command>
```

## Flags

| Flag | Commands | Default | Description |
|------|----------|---------|-------------|
| `--reviewers` | `review`, `loop` | *(required)* | Comma-separated list of reviewers |
| `--assistant` | `review`, `fix` | *(required)* | AI assistant to use |
| `--review-assistant` | `loop` | *(required)* | AI assistant to use for reviewing |
| `--fix-assistant` | `loop` | *(required)* | AI assistant to use for fixing |
| `--max-attempts` | `loop` | `5` | Maximum number of fix attempts |
| `--verbose` | all | `false` | Enable debug logging |

## Reviewers

Each reviewer focuses on a specific aspect of code quality. Available reviewers:

| Name | Focus |
|------|-------|
| `architecture` | Structural and design concerns |
| `cli` | CLI interface and flag usage |
| `duplication` | Repeated or redundant code |
| `go` | Go idioms and best practices |
| `security` | Security vulnerabilities |
| `tests` | Test coverage and quality |
| `unused` | Dead code and unused declarations |

Combine multiple reviewers with commas:

```sh
agent review --reviewers go,security,tests --assistant claude
```

## Assistants

| Name | Requires |
|------|----------|
| `claude` | [`claude` CLI](https://github.com/anthropics/claude-code) |
| `copilot` | [`copilot` CLI](https://github.com/github/gh-copilot) |

## Compose review and fix

`review` and `fix` are designed to be piped together:

```sh
agent review --reviewers go,security --assistant claude | agent fix --assistant copilot
```

Or save issues to a file for inspection before fixing:

```sh
agent review --reviewers go,security,tests --assistant copilot | tee issues.json
agent fix --assistant claude < issues.json
```

## Development

```sh
# Run tests
go test ./...

# Vet
go vet ./...

# Format
go fmt ./...
```
