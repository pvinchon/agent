package assistant

import (
	"flag"
)

// Flag registers an --assistant flag and returns a function that resolves the
// chosen Assistant after flag.Parse() has been called.
func Flag() func() (Assistant, error) {
	name := flag.String("assistant", "claude", "AI assistant to use: "+assistantNames)
	return func() (Assistant, error) {
		return New(*name)
	}
}
