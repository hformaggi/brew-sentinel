package brew

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestStreamCancellationStopsChildProcesses(t *testing.T) {
	c := Client{Path: "/bin/sh"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	started := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Stream(ctx, []string{"-c", `trap 'kill "$child"; exit 130' INT; sleep 30 & child=$!; printf '%s' "$child" > "$1"; printf ready; wait`, "sh", pidFile}, func(string) {
			select {
			case started <- struct{}{}:
			default:
			}
		})
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("process did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// Clean up the deliberately spawned child if cancellation regresses.
		raw, err := os.ReadFile(pidFile)
		if err == nil {
			if pid, err := strconv.Atoi(string(raw)); err == nil {
				syscall.Kill(pid, syscall.SIGKILL)
			}
		}
		<-done
		t.Fatal("child process kept the command running after cancellation")
	}
}

func TestStreamCancellationAllowsInterruptCleanup(t *testing.T) {
	c := Client{Path: "/bin/sh"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanupFile := filepath.Join(t.TempDir(), "cleanup")
	started := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Stream(ctx, []string{"-c", `trap 'printf cleaned > "$1"; exit 130' INT; printf ready; while :; do sleep 0.05; done`, "sh", cleanupFile}, func(string) {
			select {
			case started <- struct{}{}:
			default:
			}
		})
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("process did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("cancellation did not finish")
	}
	if data, err := os.ReadFile(cleanupFile); err != nil || string(data) != "cleaned" {
		t.Fatalf("interrupt cleanup was skipped: %q %v", data, err)
	}
}

func TestStreamCancellationForcesUnresponsiveCommand(t *testing.T) {
	c := Client{Path: "/bin/sh"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.Stream(ctx, []string{"-c", `trap '' INT; printf ready; exec sleep 30`}, func(string) {
			select {
			case started <- struct{}{}:
			default:
			}
		})
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("process did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("unresponsive command was not terminated")
	}
}
