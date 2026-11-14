package brew

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// runCaptured gives Homebrew time to handle an interrupt, restore links and
// forward the signal to its own subprocesses before forcing a stuck group down.
// Interactive commands stay in the terminal's group and receive its signals.
func runCaptured(cmd *exec.Cmd) error {
	const interruptGrace = 5 * time.Second
	finished := make(chan struct{})
	defer close(finished)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if err := signalGroup(cmd, syscall.SIGINT); err != nil {
			return err
		}
		go func() {
			timer := time.NewTimer(interruptGrace)
			defer timer.Stop()
			select {
			case <-finished:
				return
			case <-timer.C:
			}
			select {
			case <-finished:
				return
			default:
				signalGroup(cmd, syscall.SIGKILL)
			}
		}()
		return nil
	}
	// Bound waits for pipes retained by subprocesses, even after the leader exits.
	cmd.WaitDelay = interruptGrace + time.Second
	return cmd.Run()
}

func signalGroup(cmd *exec.Cmd, signal syscall.Signal) error {
	if err := syscall.Kill(-cmd.Process.Pid, signal); err != nil {
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}
