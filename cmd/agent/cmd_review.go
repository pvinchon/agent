package main

import (
"encoding/json"
"flag"
"fmt"
"io"
"log"
"log/slog"
"os"

"github.com/pvinchon/agent/internal/assistant"
"github.com/pvinchon/agent/internal/git"
"github.com/pvinchon/agent/internal/reviewer"
xlog "github.com/pvinchon/agent/internal/x/log"
)

func reviewUsage(fs *flag.FlagSet) {
fmt.Fprint(os.Stderr, `usage: agent review [flags]

Reviews the current diff and outputs issues as JSON to stdout.

Flags:
`)
fs.PrintDefaults()
}

func reviewFlags(fs *flag.FlagSet) (mustReviewers func() []reviewer.Reviewer, reviewTemplate func() string, mustAssistant func() assistant.Assistant, resolveLog func() *slog.Logger) {
return reviewer.FlagSet(fs), reviewer.TemplateFlagSet(fs), assistant.FlagSet(fs), xlog.FlagSet(fs)
}

func runReview(args []string, w io.Writer) {
fs := flag.NewFlagSet("review", flag.ExitOnError)
fs.Usage = func() { reviewUsage(fs) }
mustReviewers, reviewTemplate, mustAssistant, resolveLog := reviewFlags(fs)
fs.Parse(args)

slog.SetDefault(resolveLog())
review(mustReviewers(), reviewTemplate(), mustAssistant(), w)
}

func review(reviewers []reviewer.Reviewer, reviewTemplate string, a assistant.Assistant, w io.Writer) {
diff, err := git.DiffWithDefault()
if err != nil {
log.Fatal(err)
}

log.Println("Reviewing difference")
issues, errs := reviewer.Review(reviewers, diff, reviewTemplate, a)
if len(errs) != 0 {
for _, err := range errs {
fmt.Fprintln(os.Stderr, err)
}
os.Exit(1)
}

if err := json.NewEncoder(w).Encode(issues); err != nil {
log.Fatal(err)
}
}
