package ux

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

func TestSpinner_verbose(t *testing.T) {
	orig := slog.Default()
	t.Cleanup(func() { slog.SetDefault(orig) })
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))

	stop := Spinner()
	stop() // no-op, must not block
}

func TestSpinner_normal(t *testing.T) {
	// Suppress spinner output during test
	oldStderr := os.Stderr
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = devNull
	t.Cleanup(func() {
		os.Stderr = oldStderr
		devNull.Close()
	})

	stop := Spinner()
	stop() // must not block or panic
}
