package assistant

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers an --assistant flag on fs and returns a function that
// resolves the chosen Assistant after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() Assistant {
	name := fs.String("assistant", "", "AI assistant to use: "+assistantNames)
	return func() Assistant {
		n := *name
		if n == "" {
			n = "claude"
		}
		a, err := New(n)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return a
	}
}
