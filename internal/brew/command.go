package brew

import (
	"context"
	"os"
	"os/exec"
)

// command keeps previews and execution under the same Homebrew policy. Orphan
// removal is an explicit, separately previewed action, never an uninstall side effect.
func (c *Client) command(ctx context.Context, args []string, interactive bool) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Path, args...)
	cmd.Env = append(os.Environ(),
		"HOMEBREW_NO_AUTO_UPDATE=1",
		"HOMEBREW_NO_INSTALL_CLEANUP=1",
		"HOMEBREW_NO_AUTOREMOVE=1",
	)
	if !interactive {
		cmd.Env = append(cmd.Env, "HOMEBREW_NO_ENV_HINTS=1", "HOMEBREW_NO_COLOR=1", "NO_COLOR=1")
	}
	return cmd
}

// Interactive attaches directly to the terminal; passwords never pass through logs.
func (c *Client) Interactive(args []string) error {
	cmd := c.command(context.Background(), args, true)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
