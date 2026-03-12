package log

import (
	"flag"
	"log/slog"
	"os"
)

// FlagSet registers -v/--verbose flags on fs and returns a function that builds
// the appropriate logger after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() *slog.Logger {
	verbose := false
	fs.BoolVar(&verbose, "verbose", false, "enable debug logging")
	fs.BoolVar(&verbose, "v", false, "shorthand for --verbose")

	return func() *slog.Logger {
		if verbose {
			return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		}
		return slog.Default()
	}
}
