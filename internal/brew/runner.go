package brew

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type Result struct {
	Stdout, Stderr []byte
	ExitCode       int
}
type Client struct{ Path string }

func New() (*Client, error) { p, e := exec.LookPath("brew"); return &Client{Path: p}, e }
func (c *Client) Run(ctx context.Context, args ...string) (Result, error) {
	cmd := c.command(ctx, args, false)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	e := runCaptured(cmd)
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
