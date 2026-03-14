package assistant

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers --assistant and --model flags on fs and returns a function
// that resolves the chosen Assistant after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() Assistant {
	name := fs.String("assistant", "claude", "AI assistant to use: "+assistantNames)
	model := fs.String("model", "", "model to use (optional, defaults to the assistant's own default model)")
	return func() Assistant {
		a, err := New(*name, *model)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return a
	}
}
