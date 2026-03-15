## Focus: Docs

Review the diff for **documentation drift** — changes that make `CLAUDE.md` inaccurate or incomplete compared to the code in the repository.

**Flag:**
- A new package added under `internal/` that is not mentioned in `CLAUDE.md`
- A new reviewer prompt added under `internal/reviewer/data/prompts/` that is not listed in the Reviewers table in `CLAUDE.md`
- A new AI assistant registered in `assistantByName` (in `internal/assistant/assistant.go`) that is not listed in `CLAUDE.md`
- A new command added to `cmd/agent/main.go` that is not described in `CLAUDE.md`
- A code convention, test pattern, or extension point that changed but is not reflected in `CLAUDE.md`

**Do not flag:**
- Cosmetic or wording differences that do not affect accuracy
- Missing documentation for unexported symbols or internal implementation details
