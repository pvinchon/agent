## Focus: CLI Design

Review the diff for command-line interface issues that **break user expectations or POSIX conventions**.

**Flag:**
- Errors printed to stdout instead of stderr, or normal output sent to stderr
- Missing or wrong exit codes — success on failure, zero on error, non-zero on `--help`
- `log.Fatal` or `os.Exit` called where deferred cleanup (file handles, temp files) would be skipped
- Required flags that should be positional arguments, or boolean flags with confusing defaults
- Missing usage output on invalid input — the user gets a raw error with no guidance
- Long-running operations with no way to interrupt cleanly (no signal handling, no context cancellation)

**Do not flag:**
- Flag libraries or frameworks the project has already committed to — review usage, not choice
- Subcommand structure preferences unless they cause ambiguity
