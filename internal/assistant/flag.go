package assistant

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers an --assistant flag on fs and returns a function that
// resolves the chosen Assistant after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() Assistant {
	name := fs.String("assistant", "claude", "AI assistant to use: "+assistantNames)
	return func() Assistant {
		a, err := New(*name)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return a
	}
}
