---
name: Duplication
description: Reviews repeated logic that will diverge over time
scope: project
---
## Focus: Duplication

Review the diff for **repeated logic that will diverge over time**, creating bugs when one copy is updated and the other is not.

**Flag:**
- Two or more blocks with near-identical structure that differ only in names or literals — extract a function or loop
- Parallel data structures that always change together (e.g. two slices indexed in lockstep) — unify into a struct
- Repeated multi-step sequences (open, do, close) that should share a helper

**Do not flag:**
- Two lines that happen to look similar but serve unrelated purposes
- Test cases that repeat setup — test readability often benefits from explicit repetition
- Code that would require a complex generic or interface to deduplicate — the cure is worse than the disease
