package assistant

import (
	"flag"
	"fmt"
	"os"
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
	model := fs.String(modelFlagName, "", "model to use for the assistant (leave empty for default; run `<assistant> models` to list available models)")
	return func() Assistant {
		if *name == "" {
			fs.Usage()
			fmt.Fprintf(os.Stderr, "error: --%s is required\n", assistantFlagName)
			os.Exit(2)
		}
		a, err := New(*name, *model)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return a
	}
}
