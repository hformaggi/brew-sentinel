package brew

import (
	"context"
	"errors"
	"os/exec"
)

// Stream preserves a bounded diagnostic transcript while forwarding progress.
func (c *Client) Stream(ctx context.Context, args []string, emit func(string)) (Result, error) {
	cmd := c.command(ctx, args, false)
	writer := &streamWriter{emit: emit}
	cmd.Stdout = writer
	cmd.Stderr = writer
	e := runCaptured(cmd)
	r := Result{Stdout: writer.data}
	var exit *exec.ExitError
	if errors.As(e, &exit) {
		r.ExitCode = exit.ExitCode()
		return r, nil
	}
	return r, e
}

type streamWriter struct {
	data []byte
	emit func(string)
}

func (w *streamWriter) Write(p []byte) (int, error) {
	const transcriptLimit = 1024 * 1024
	if len(p) >= transcriptLimit {
		w.data = append(w.data[:0], p[len(p)-transcriptLimit:]...)
	} else {
		if excess := len(w.data) + len(p) - transcriptLimit; excess > 0 {
			copy(w.data, w.data[excess:])
			w.data = w.data[:len(w.data)-excess]
		}
		w.data = append(w.data, p...)
	}
	w.emit(string(p))
	return len(p), nil
}
