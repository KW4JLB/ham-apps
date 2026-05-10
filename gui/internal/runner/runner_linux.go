//go:build linux

package runner

import (
	"os/exec"
	"syscall"
	"time"
)

// setPlatformAttrs sets Linux-specific process attributes: Setpgid=true causes
// the child and all its descendants to share a new process group, enabling
// kill-by-pgid for clean cancellation.
func setPlatformAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// makeCancelFunc returns a cancel function that performs a two-phase kill:
// 1. SIGTERM to the entire process group (-pgid).
// 2. SIGKILL after 3 seconds if the group hasn't exited.
//
// The timer is stopped if cancel is called only once (idiomatic usage).
func makeCancelFunc(cmd *exec.Cmd) func() {
	return func() {
		if cmd.Process == nil {
			return
		}
		pgid := cmd.Process.Pid
		// Phase 1: SIGTERM to the process group.
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
		// Phase 2: SIGKILL after 3 seconds.
		timer := time.AfterFunc(3*time.Second, func() {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		})
		_ = timer // Timer fires asynchronously; caller waits on done channel.
		// Note: We intentionally do NOT stop the timer here — if the process
		// ignores SIGTERM, SIGKILL must fire. The goroutine in runner.go will
		// receive the exit event and the caller should discard the timer.
	}
}
