package reviewer

import (
"flag"
"fmt"
"maps"
"os"
)

// FlagSet registers --reviewers and --prompts-dir flags on fs and returns a
// function that resolves the chosen reviewers after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() []Reviewer {
names := fs.String("reviewers", "", "Comma-separated list of reviewers to use: "+reviewerNames)
promptsDir := fs.String("prompts-dir", "", "Directory containing additional or replacement reviewer prompt files (.md)")
return func() []Reviewer {
if *names == "" {
fs.Usage()
fmt.Fprintln(os.Stderr, "error: --reviewers is required")
os.Exit(2)
}
registry := reviewersByName
if *promptsDir != "" {
custom, err := LoadPromptsDir(*promptsDir)
if err != nil {
fs.Usage()
fmt.Fprintln(os.Stderr, "error:", err)
os.Exit(2)
}
merged := make(map[string]Reviewer, len(reviewersByName)+len(custom))
maps.Copy(merged, reviewersByName)
maps.Copy(merged, custom)
registry = merged
}
r, err := resolveFrom(registry, *names)
if err != nil {
fs.Usage()
fmt.Fprintln(os.Stderr, "error:", err)
os.Exit(2)
}
return r
}
}

// TemplateFlagSet registers a --review-template flag on fs and returns a
// function that returns the template content after fs.Parse() has been called.
// When no path is provided the returned string is empty, which causes the
// built-in template to be used.
func TemplateFlagSet(fs *flag.FlagSet) func() string {
path := fs.String("review-template", "", "Path to a custom review prompt template file (must contain {{prompt}} and {{diff}} placeholders; the output specification is always appended)")
return func() string {
if *path == "" {
return ""
}
data, err := os.ReadFile(*path)
if err != nil {
fs.Usage()
fmt.Fprintln(os.Stderr, "error:", err)
os.Exit(2)
}
return string(data)
}
}
