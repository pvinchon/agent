package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/fixer"
	"github.com/pvinchon/agent/internal/prompt"
	"github.com/pvinchon/agent/internal/reviewer"
	xlog "github.com/pvinchon/agent/internal/x/log"
)

func loopUsage(fs *flag.FlagSet) {
	fmt.Fprint(os.Stderr, `usage: agent loop [flags]

Runs review and fix in a loop until no issues remain or max attempts is reached.

Flags:
`)
	fs.PrintDefaults()
}

func loopFlags(fs *flag.FlagSet) (
	mustReviewers        func() []reviewer.Reviewer,
	mustReviewerTemplate func() prompt.Prompt,
	mustReviewAssistant  func() assistant.Assistant,
	mustFixAssistant     func() assistant.Assistant,
	mustFixerTemplate    func() prompt.Prompt,
	resolveLog           func() *slog.Logger,
	maxAttempts          *int,
) {
	maxAttempts = fs.Int("max-attempts", 5, "maximum number of fix attempts")
	mustReviewers, mustReviewerTemplate = reviewer.FlagSet(fs)
	mustReviewAssistant = assistant.FlagSet(fs, "review")
	mustFixAssistant = assistant.FlagSet(fs, "fix")
	mustFixerTemplate = fixer.FlagSet(fs)
	resolveLog = xlog.FlagSet(fs)
	return
}

func runLoop(args []string) {
	fs := flag.NewFlagSet("loop", flag.ExitOnError)
	fs.Usage = func() { loopUsage(fs) }
	mustReviewers, _, mustReviewAssistant, mustFixAssistant, mustFixerTemplate, resolveLog, maxAttempts := loopFlags(fs)
	fs.Parse(args)

	slog.SetDefault(resolveLog())

	reviewers := mustReviewers()
	ra := mustReviewAssistant()
	fa := mustFixAssistant()
	fixTmpl := mustFixerTemplate()

	for attempt := 1; attempt <= *maxAttempts; attempt++ {
		fmt.Printf("Attempt %d/%d\n", attempt, *maxAttempts)

		var buf bytes.Buffer
		review(reviewers, ra, &buf)

		var issues []reviewer.Issue
		if err := json.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&issues); err != nil {
			log.Fatal(err)
		}

		if len(issues) == 0 {
			fmt.Println("No issues found")
			return
		}

		log.Printf("Fixing %d issue(s)", len(issues))
		fix(fa, bytes.NewReader(buf.Bytes()), fixTmpl)
	}

	fmt.Printf("Reached max attempts (%d)\n", *maxAttempts)
}
