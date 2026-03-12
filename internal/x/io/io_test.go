package io_test

import (
	"bytes"
	"testing"

	xio "github.com/pvinchon/agent/internal/x/io"
)

func TestPrefixWriter(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		writes []string
		want   string
	}{
		{
			name:   "single line with newline",
			prefix: "pre> ",
			writes: []string{"hello\n"},
			want:   "pre> hello\n",
		},
		{
			name:   "multiple lines in one write",
			prefix: ">> ",
			writes: []string{"foo\nbar\nbaz\n"},
			want:   ">> foo\n>> bar\n>> baz\n",
		},
		{
			name:   "line split across writes",
			prefix: "> ",
			writes: []string{"hel", "lo\n"},
			want:   "> hello\n",
		},
		{
			name:   "multiple writes multiple lines",
			prefix: "> ",
			writes: []string{"foo\nbar\n", "baz\n"},
			want:   "> foo\n> bar\n> baz\n",
		},
		{
			name:   "partial last line is buffered",
			prefix: "> ",
			writes: []string{"foo\nbar"},
			want:   "> foo\n",
		},
		{
			name:   "empty write",
			prefix: "> ",
			writes: []string{""},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := xio.PrefixWriter(&buf, tt.prefix)
			for _, s := range tt.writes {
				if _, err := w.Write([]byte(s)); err != nil {
					t.Fatalf("Write() error: %v", err)
				}
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("PrefixWriter wrote %q, want %q", got, tt.want)
			}
		})
	}
}
