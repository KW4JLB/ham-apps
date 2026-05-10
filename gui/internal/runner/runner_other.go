//go:build !linux

package runner

import "os/exec"

// setPlatformAttrs is a no-op on non-Linux platforms.
func setPlatformAttrs(_ *exec.Cmd) {}

// makeCancelFunc returns a cancel function that calls Process.Kill directly.
// Non-Linux platforms do not support process-group kill.
func makeCancelFunc(cmd *exec.Cmd) func() {
	return func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
}
