// Package runner implements port.RunnerService using os/exec to run bash scripts.
// It provides BashRunner (the concrete implementation), platform-specific process-
// group management (Linux: runner_linux.go; other: runner_other.go), and sudo
// credential helpers (sudo.go).
package runner

import (
	"os"
	"os/exec"

	"github.com/kw4jlb/ham-apps/gui/internal/port"
)

// BashRunner implements port.RunnerService.
// It runs bash scripts as background processes, captures combined stdout+stderr
// to a temp log file, and provides two-phase cancel (Linux: SIGTERM → SIGKILL;
// other platforms: immediate Process.Kill).
type BashRunner struct{}

// Start executes script asynchronously: exec.Command("bash", script, slug).
// It returns a cancel function and a channel that receives one RunResult when
// the process completes or is killed.
//
// The log file is created with mode 0600. The caller is responsible for
// deleting it (its path is in RunResult.LogFile).
func (r *BashRunner) Start(script, slug string) (cancel func(), logFile string, done <-chan port.RunResult) {
	// SC-02: use argument array, never shell string interpolation.
	cmd := exec.Command("bash", script, slug)

	// Create the log file (mode 0600) for combined stdout+stderr. SC-03.
	logFd, err := os.CreateTemp("", "hamapps-*.log")
	if err != nil {
		ch := make(chan port.RunResult, 1)
		ch <- port.RunResult{ExitCode: -1, Err: err}
		return func() {}, "", ch
	}
	logPath := logFd.Name()
	// Set mode explicitly to 0600 (os.CreateTemp already does this on Linux,
	// but we set it explicitly for clarity and cross-platform safety).
	if err := logFd.Chmod(0600); err != nil {
		logFd.Close()
		os.Remove(logPath)
		ch := make(chan port.RunResult, 1)
		ch <- port.RunResult{ExitCode: -1, Err: err}
		return func() {}, "", ch
	}
	cmd.Stdout = logFd
	cmd.Stderr = logFd

	// Platform-specific: set process-group attributes (Linux: Setpgid=true).
	setPlatformAttrs(cmd)

	if err := cmd.Start(); err != nil {
		logFd.Close()
		os.Remove(logPath)
		ch := make(chan port.RunResult, 1)
		ch <- port.RunResult{ExitCode: -1, Err: err}
		return func() {}, "", ch
	}
	logFd.Close() // Close our handle; the subprocess holds its own fd.

	cancelFn := makeCancelFunc(cmd)

	ch := make(chan port.RunResult, 1)
	go func() {
		waitErr := cmd.Wait()
		exitCode := 0
		if waitErr != nil {
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}
		ch <- port.RunResult{
			ExitCode: exitCode,
			LogFile:  logPath,
			Err:      waitErr,
		}
	}()

	return cancelFn, logPath, ch
}

// CheckSudo returns true if sudo credentials are currently cached.
// It runs "sudo -n true" and checks for exit 0.
func (r *BashRunner) CheckSudo() bool {
	err := exec.Command("sudo", "-n", "true").Run()
	return err == nil
}

// PromptSudo is implemented in sudo.go.
func (r *BashRunner) PromptSudo(appName string) error {
	return promptSudo(appName)
}
