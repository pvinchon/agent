package assistant

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

var modelDescription = func() string {
	parts := make([]string, 0, len(modelsByAssistant))
	for name, models := range modelsByAssistant {
		parts = append(parts, name+": "+strings.Join(models, ", "))
	}
	// keep output deterministic
	sort.Strings(parts)
	return strings.Join(parts, "; ")
}()

// FlagSet registers --assistant and --model flags on fs and returns a function
// that resolves the chosen Assistant after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() Assistant {
	name := fs.String("assistant", "claude", "AI assistant to use: "+assistantNames)
	model := fs.String("model", "", "model to use (optional, defaults to the assistant's default model; "+modelDescription+")")
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
