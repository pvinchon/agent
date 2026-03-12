package io

import (
	"bytes"
	"io"
)

// PrefixWriter returns a writer that prepends prefix to each line written to w.
func PrefixWriter(w io.Writer, prefix string) io.Writer {
	return &prefixWriter{w: w, prefix: []byte(prefix)}
}

type prefixWriter struct {
	w      io.Writer
	prefix []byte
	// pending holds an incomplete line not yet written.
	pending []byte
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	n := len(b)
	for len(b) > 0 {
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			// No newline: buffer for later.
			p.pending = append(p.pending, b...)
			break
		}
		line := append(p.pending, b[:i+1]...)
		p.pending = p.pending[:0]
		if _, err := p.w.Write(p.prefix); err != nil {
			return 0, err
		}
		if _, err := p.w.Write(line); err != nil {
			return 0, err
		}
		b = b[i+1:]
	}
	return n, nil
}
