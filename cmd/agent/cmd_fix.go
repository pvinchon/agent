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
"github.com/pvinchon/agent/internal/fixer"
"github.com/pvinchon/agent/internal/git"
"github.com/pvinchon/agent/internal/reviewer"
xlog "github.com/pvinchon/agent/internal/x/log"
)

func fixUsage(fs *flag.FlagSet) {
fmt.Fprint(os.Stderr, `usage: agent fix [flags]

Reads issues as JSON from stdin and applies fixes.

Flags:
`)
fs.PrintDefaults()
}

func fixFlags(fs *flag.FlagSet) (mustAssistant func() assistant.Assistant, resolveLog func() *slog.Logger, fixTemplate func() string) {
return assistant.FlagSet(fs), xlog.FlagSet(fs), fixer.TemplateFlagSet(fs)
}

func runFix(args []string, r io.Reader) {
fs := flag.NewFlagSet("fix", flag.ExitOnError)
fs.Usage = func() { fixUsage(fs) }
mustAssistant, resolveLog, fixTemplate := fixFlags(fs)
fs.Parse(args)

slog.SetDefault(resolveLog())
fix(mustAssistant(), fixTemplate(), r)
}

func fix(a assistant.Assistant, fixTemplate string, r io.Reader) {
var issues []reviewer.Issue
if err := json.NewDecoder(r).Decode(&issues); err != nil {
log.Fatal(err)
}

diff, err := git.DiffWithDefault()
if err != nil {
log.Fatal(err)
}

log.Println("Fixing issues")
if err := fixer.Fix(issues, diff, fixTemplate, a); err != nil {
log.Fatal(err)
}
}
