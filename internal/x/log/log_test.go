package log

import (
	"io"
	"log/slog"
	"testing"
)

func TestIsLevelDebug(t *testing.T) {
	orig := slog.Default()
	t.Cleanup(func() { slog.SetDefault(orig) })

	if IsLevelDebug() {
		t.Error("expected false with default logger")
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))
	if !IsLevelDebug() {
		t.Error("expected true with debug logger")
	}
}
