---
name: Unused Code
description: Reviews dead code introduced or left behind
scope: project
---
## Focus: Unused Code

Review the diff for **dead code introduced or left behind** that adds noise without serving any purpose.

**Flag:**
- Variables or parameters declared and never read — including `err` that is assigned but not checked
- Functions, methods, or types defined in the diff that are never called or referenced
- Imports added but not used (if the language allows it to compile)
- Unreachable code — statements after an unconditional `return`, `panic`, or `os.Exit`; branches guarded by conditions that are always true or always false
- Commented-out code blocks — either delete or explain with a TODO why they exist

**Do not flag:**
- Exported symbols that may be used by external consumers you cannot see in the diff
- Interface methods required by a contract even if not called directly
- `_` used intentionally to discard values (e.g. `_ = flag.String(...)` for side effects)
