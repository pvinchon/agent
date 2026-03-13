package assistant

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers an --assistant flag on fs and returns a function that
// resolves the chosen Assistant after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() Assistant {
	return NamedFlagSet(fs, "assistant")
}

// NamedFlagSet registers a flag with the given flagName on fs and returns a
// function that resolves the chosen Assistant after fs.Parse() has been called.
func NamedFlagSet(fs *flag.FlagSet, flagName string) func() Assistant {
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
