package log

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"testing"
)

func TestFlag(t *testing.T) {
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
			origCommandLine := flag.CommandLine
			t.Cleanup(func() { flag.CommandLine = origCommandLine })

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			resolve := Flag()
			flag.CommandLine.Parse(tt.args)
			logger := resolve()

			got := logger.Enabled(context.Background(), slog.LevelDebug)
			if got != tt.wantVerbose {
				t.Errorf("debug enabled = %v, want %v", got, tt.wantVerbose)
			}
		})
	}
}
