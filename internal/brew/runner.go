package brew

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Result struct {
	Stdout, Stderr []byte
	ExitCode       int
}
type Client struct{ Path string }

func New() (*Client, error) { p, e := exec.LookPath("brew"); return &Client{Path: p}, e }
func (c *Client) Run(ctx context.Context, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, c.Path, args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ENV_HINTS=1", "HOMEBREW_NO_INSTALL_CLEANUP=1", "HOMEBREW_COLOR=0")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	e := cmd.Run()
	r := Result{Stdout: out.Bytes(), Stderr: stderr.Bytes()}
	var exit *exec.ExitError
	if errors.As(e, &exit) {
		r.ExitCode = exit.ExitCode()
		return r, nil
	}
	return r, e
}
func (c *Client) Read(ctx context.Context, args ...string) ([]byte, error) {
	r, e := c.Run(ctx, args...)
	if e != nil {
		return nil, e
	}
	if r.ExitCode != 0 {
		return nil, fmt.Errorf("brew %v: %s", args, r.Stderr)
	}
	return r.Stdout, nil
}

// Stream preserves a bounded diagnostic transcript while forwarding progress.
func (c *Client) Stream(ctx context.Context, args []string, emit func(string)) (Result, error) {
	cmd := exec.CommandContext(ctx, c.Path, args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_INSTALL_CLEANUP=1", "HOMEBREW_COLOR=0")
	writer := &streamWriter{emit: emit}
	cmd.Stdout = writer
	cmd.Stderr = writer
	e := cmd.Run()
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
	remaining := 1024*1024 - len(w.data)
	if remaining > 0 {
		part := p
		if len(part) > remaining {
			part = part[:remaining]
		}
		w.data = append(w.data, part...)
		w.emit(string(part))
	}
	return len(p), nil
}
