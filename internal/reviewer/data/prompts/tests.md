## Focus: Tests

Review the diff for **test quality issues** that give false confidence or leave critical behaviour unverified.

**Flag:**
- **Noop tests** — tests with no assertions, assertions that compare a value to itself, or `if err != nil { t.Log(err) }` instead of `t.Fatal`
- **Wrong assertions** — expected and actual swapped, inequality used where equality is needed, error checked but value ignored
- **Missing error path tests** — only the happy path is tested; invalid input, timeouts, and permission errors are not exercised
- **Brittle tests** — assertions on unstable values (timestamps, map order, goroutine scheduling) that will flake
- **Tests that test the mock** — the test asserts on behaviour provided entirely by the mock setup, proving nothing about production code

**Do not flag:**
- Missing tests for trivial getters, simple delegation, or auto-generated code
- Test helper style (table-driven vs. individual functions) — that is a preference, not a defect
