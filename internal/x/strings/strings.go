package strings

import "strings"

// StripMarkdownFence removes a leading ```lang fence and trailing ``` from s.
func StripMarkdownFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = s[3:]
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimSuffix(strings.TrimRight(s, "\n"), "```")
	return strings.TrimSpace(s)
}
