package brew

import (
	"brew-sentinel/internal/i18n"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var packageName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9@+_.-]*(/[a-zA-Z0-9][a-zA-Z0-9@+_.-]*){0,2}$`)

func ActionArgs(action, kind, name string) ([]string, error) {
	if action == "cleanup" || action == "autoremove" {
		return []string{action}, nil
	}
	if action != "install" && action != "upgrade" && action != "uninstall" {
		return nil, i18n.Errorf("invalid action")
	}
	if kind != "formula" && kind != "cask" {
		return nil, i18n.Errorf("invalid package type")
	}
	if !packageName.MatchString(name) || strings.Contains(name, "..") {
		return nil, i18n.Errorf("invalid package name")
	}
	return []string{action, "--" + kind, name}, nil
}

// Interactive attaches directly to the terminal; passwords never pass through logs.
func (c *Client) Interactive(args []string) error {
	cmd := exec.Command(c.Path, args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_INSTALL_CLEANUP=1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func PreviewArgs(action, kind, name string) ([]string, error) {
	if action == "uninstall" {
		return nil, i18n.Errorf("uninstall does not support dry-run")
	}
	args, e := ActionArgs(action, kind, name)
	if e != nil {
		return nil, e
	}
	return append(args, "--dry-run"), nil
}
