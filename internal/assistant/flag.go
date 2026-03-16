package assistant

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// FlagSet registers --assistant and --model flags on fs and returns a function
// that resolves the chosen Assistant after fs.Parse() has been called.
// A non-empty suffix (e.g. "review" or "fix") produces flags named
// --assistant-for-<suffix> and --model-for-<suffix> instead of --assistant/--model.
func FlagSet(fs *flag.FlagSet, suffix string) func() Assistant {
	assistantFlagName := "assistant"
	modelFlagName := "model"
	if suffix != "" {
		assistantFlagName = "assistant-for-" + suffix
		modelFlagName = "model-for-" + suffix
	}
	name := fs.String(assistantFlagName, "", "AI assistant to use: "+assistantNames)
	model := fs.String(modelFlagName, "", "model to use for the assistant (leave empty for default; available models depend on the chosen assistant)")
	return func() Assistant {
		if *name == "" {
			fs.Usage()
			fmt.Fprintf(os.Stderr, "error: --%s is required\n", assistantFlagName)
			os.Exit(2)
		}
		a, err := New(*name, *model)
		if err != nil {
			fs.Usage()
			if *model != "" {
				// provide available models in the error message
				if models, ok := modelsByAssistant[*name]; ok {
					fmt.Fprintf(os.Stderr, "error: %s\navailable models for %s: %s\n", err, *name, strings.Join(models, ", "))
					os.Exit(2)
				}
			}
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return a
	}
}
