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

Optionally supply a custom review prompt template or additional reviewer prompts:

```sh
agent review --reviewers go,python --prompts-dir ./my-prompts --review-template ./my-template.md --assistant claude
```

### `fix`

Reads issues as JSON from stdin and applies fixes to the working tree.

```sh
agent fix --assistant <name> < issues.json
```

```sh
agent fix --assistant copilot < issues.json
```

Optionally supply a custom fix prompt template:

```sh
agent fix --fix-template ./my-fix-template.md --assistant claude < issues.json
```

### `loop`

Runs review and fix in a loop until no issues remain or the maximum number of attempts is reached.

```sh
agent loop --reviewers <list> --assistant <name> [--max-attempts <n>]
```

```sh
agent loop --reviewers go,security,tests --assistant claude --max-attempts 3
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
| `--assistant` | all | *(required)* | AI assistant to use |
| `--max-attempts` | `loop` | `5` | Maximum number of fix attempts |
| `--verbose` | all | `false` | Enable debug logging |
| `--review-template` | `review`, `loop` | *(built-in)* | Path to a custom review prompt template file |
| `--prompts-dir` | `review`, `loop` | *(built-in)* | Directory containing custom reviewer prompt files (`.md`) |
| `--fix-template` | `fix`, `loop` | *(built-in)* | Path to a custom fix prompt template file |

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

## Custom prompts

### Review prompt template (`--review-template`)

Supply a Markdown file as a replacement for the built-in scene-setting context. The file must contain the `{{prompt}}` and `{{diff}}` placeholders. The JSON output specification is **always appended automatically** — you do not need to include it in your template.

Example `my-review-template.md`:

```markdown
You are a strict reviewer. Focus only on correctness.

## Focus

{{prompt}}

## Code

{{diff}}
```

```sh
agent review --reviewers security --review-template ./my-review-template.md --assistant claude
```

### Custom reviewer prompts (`--prompts-dir`)

Point to a directory containing `.md` files. Each file name (without the extension) becomes a reviewer name. Custom prompts override built-in reviewers with the same name and can introduce entirely new reviewer types.

```
my-prompts/
  python.md      # new reviewer: "python"
  security.md    # overrides the built-in "security" reviewer
```

```sh
agent review --reviewers security,python --prompts-dir ./my-prompts --assistant claude
```

### Fix prompt template (`--fix-template`)

Supply a Markdown file as a replacement for the built-in fix prompt. The file must contain the `{{issues}}` and `{{diff}}` placeholders.

Example `my-fix-template.md`:

```markdown
Fix the following issues in the code.

## Issues

{{issues}}

## Diff

{{diff}}
```

```sh
agent fix --fix-template ./my-fix-template.md --assistant claude < issues.json
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
