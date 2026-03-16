---
name: Go
description: Reviews code for Go-specific bugs and convention violations
scope: file
---

## Focus: Go

Review the diff for **Go-specific bugs and convention violations** based on Effective Go, the Go Code Review Comments wiki, and standard library patterns.

**Flag:**
- Errors assigned to `_` or returned without checking — every error must be handled or explicitly documented as safe to ignore
- Missing error wrapping — `fmt.Errorf("…: %w", err)` to preserve the chain
- Exported names that stutter with the package name (`http.HTTPClient` → `http.Client`)
- Interfaces declared in the implementing package instead of the consuming package
- Goroutines launched without a cancellation path — no `context.Context`, no `done` channel, no `sync.WaitGroup`
- Mutable package-level state that creates hidden coupling or race conditions
- Deferred calls inside loops — defers do not run until the function returns
- Slice/map nil vs. empty confusion — `var s []T` vs. `s := []T{}` used inconsistently in API contracts

**Do not flag:**
- Stylistic choices already enforced by `gofmt` or `go vet`
- Use of `panic` in `init` or test helpers where it is idiomatic
