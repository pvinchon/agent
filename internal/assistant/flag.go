package assistant

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers an --assistant flag on fs and returns a function that
// resolves the chosen Assistant after fs.Parse() has been called.
// A non-empty suffix (e.g. "review" or "fix") produces a flag named
// --assistant-for-<suffix> instead of --assistant.
func FlagSet(fs *flag.FlagSet, suffix string) func() Assistant {
	flagName := "assistant"
	if suffix != "" {
		flagName = "assistant-for-" + suffix
	}
	name := fs.String(flagName, "", "AI assistant to use: "+assistantNames)
	return func() Assistant {
		if *name == "" {
			fs.Usage()
			fmt.Fprintf(os.Stderr, "error: --%s is required\n", flagName)
			os.Exit(2)
		}
		a, err := New(*name)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return a
	}
}
