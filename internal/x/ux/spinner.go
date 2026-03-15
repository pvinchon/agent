package ux

import (
	"fmt"
	"os"
	"time"

	xlog "github.com/pvinchon/agent/internal/x/log"
)

// Spinner starts an animated terminal spinner on stderr and returns a stop
// function. When verbose logging is enabled, the spinner is suppressed since
// debug output already indicates progress.
func Spinner() func() {
	if xlog.IsLevelDebug() {
		return func() {}
	}
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		frames := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r \r")
				return
			case <-time.After(100 * time.Millisecond):
				fmt.Fprintf(os.Stderr, "\r%s", frames[i%len(frames)])
				i++
			}
		}
	}()
	return func() {
		close(done)
		<-stopped
	}
}
