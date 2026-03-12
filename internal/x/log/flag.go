package log

import (
	"flag"
	"log/slog"
	"os"
)

// Flag registers -v/--verbose flags and returns a function that builds the
// appropriate logger after flag.Parse() has been called.
func Flag() func() *slog.Logger {
	verbose := false
	flag.BoolVar(&verbose, "verbose", false, "enable debug logging")
	flag.BoolVar(&verbose, "v", false, "shorthand for --verbose")

	return func() *slog.Logger {
		if verbose {
			return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		}
		return slog.Default()
	}
}
