package log

import (
	"context"
	"log/slog"
)

// IsLevelDebug reports whether the default logger has debug logging enabled.
func IsLevelDebug() bool {
	return slog.Default().Enabled(context.Background(), slog.LevelDebug)
}
