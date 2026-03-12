package log

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"testing"
)

func TestFlagSet(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantVerbose bool
	}{
		{name: "--verbose", args: []string{"--verbose"}, wantVerbose: true},
		{name: "-v", args: []string{"-v"}, wantVerbose: true},
		{name: "default", args: []string{}, wantVerbose: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			resolve := FlagSet(fs)
			fs.Parse(tt.args)
			logger := resolve()

			got := logger.Enabled(context.Background(), slog.LevelDebug)
			if got != tt.wantVerbose {
				t.Errorf("debug enabled = %v, want %v", got, tt.wantVerbose)
			}
		})
	}
}
