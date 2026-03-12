package strings_test

import (
	"testing"

	xstrings "github.com/pvinchon/agent/internal/x/strings"
)

func TestStripMarkdownFence(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fence",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "plain fence",
			input: "```\nhello\n```",
			want:  "hello",
		},
		{
			name:  "language fence",
			input: "```json\n{\"key\": \"value\"}\n```",
			want:  "{\"key\": \"value\"}",
		},
		{
			name:  "multiline",
			input: "```go\nfunc foo() {}\nfunc bar() {}\n```",
			want:  "func foo() {}\nfunc bar() {}",
		},
		{
			name:  "no closing fence",
			input: "```\nhello",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xstrings.StripMarkdownFence(tt.input)
			if got != tt.want {
				t.Errorf("StripMarkdownFence(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
