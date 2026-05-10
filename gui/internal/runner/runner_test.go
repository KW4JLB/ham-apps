// Package runner_test contains unit tests for the runner package.
// Tests are written TDD-first (red phase) before the implementation exists.
//
// Linux-only tests (signal tests) use build constraints or t.Skip.
package runner_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/kw4jlb/ham-apps/gui/internal/port"
	"github.com/kw4jlb/ham-apps/gui/internal/runner"
)

// Compile-time check: BashRunner must implement port.RunnerService (UT-30).
var _ port.RunnerService = (*runner.BashRunner)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newRunner creates a BashRunner for testing with a temp hamappsDir.
func newRunner(t *testing.T) *runner.BashRunner {
	t.Helper()
	return &runner.BashRunner{}
}

// writeTempScript writes a bash script to a temp file and returns the path.
func writeTempScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(path, []byte("#!/bin/bash\n"+content+"\n"), 0755); err != nil {
		t.Fatalf("writeTempScript: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// UT-20: runner — process exits 0
// ---------------------------------------------------------------------------

func TestBashRunner_ExitZero(t *testing.T) {
	r := newRunner(t)
	script := writeTempScript(t, "echo hello && exit 0")

	cancel, _, done := r.Start(script, "test-app")
	_ = cancel // not used in this test

	select {
	case result := <-done:
		if result.ExitCode != 0 {
			t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
		}
		if result.Err != nil {
			t.Errorf("Err: got %v, want nil", result.Err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for process to exit")
	}
}

// ---------------------------------------------------------------------------
// UT-21: runner — process exits non-zero
// ---------------------------------------------------------------------------

func TestBashRunner_ExitNonZero(t *testing.T) {
	r := newRunner(t)
	script := writeTempScript(t, "exit 42")

	_, _, done := r.Start(script, "test-app")

	select {
	case result := <-done:
		if result.ExitCode != 42 {
			t.Errorf("ExitCode: got %d, want 42", result.ExitCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for process to exit")
	}
}

// ---------------------------------------------------------------------------
// Log file creation and mode test
// ---------------------------------------------------------------------------

func TestBashRunner_LogFileMode(t *testing.T) {
	r := newRunner(t)
	script := writeTempScript(t, "echo logtest && exit 0")

	_, _, done := r.Start(script, "test-app")

	select {
	case result := <-done:
		if result.LogFile == "" {
			t.Fatal("LogFile: expected non-empty path")
		}
		info, err := os.Stat(result.LogFile)
		if err != nil {
			t.Fatalf("LogFile stat: %v", err)
		}
		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("LogFile mode: got %04o, want 0600", mode)
		}
		// Cleanup.
		os.Remove(result.LogFile)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for process to exit")
	}
}

// ---------------------------------------------------------------------------
// UT-19: runner — process cancelled before completion (two-phase kill)
// ---------------------------------------------------------------------------

func TestBashRunner_Cancel(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("signal tests are Linux-only")
	}

	r := newRunner(t)
	script := writeTempScript(t, "sleep 30")

	cancel, _, done := r.Start(script, "test-app")

	// Cancel immediately.
	cancel()

	select {
	case result := <-done:
		if result.ExitCode == 0 {
			t.Errorf("ExitCode after cancel: got 0, want non-zero (process was killed)")
		}
		if result.LogFile != "" {
			if _, err := os.Stat(result.LogFile); err != nil {
				t.Errorf("LogFile after cancel should still exist: %v", err)
			}
			os.Remove(result.LogFile)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: cancel should have stopped the process within 2 seconds")
	}
}

// ---------------------------------------------------------------------------
// UT-19b: runner — SIGTERM-ignoring process killed by SIGKILL within 5 seconds
// ---------------------------------------------------------------------------

func TestBashRunner_SIGTERMIgnoreKilledBySIGKILL(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("signal tests are Linux-only")
	}

	r := newRunner(t)
	// Script traps and ignores SIGTERM but will exit when SIGKILL is sent.
	script := writeTempScript(t, "trap '' TERM; sleep 60")

	cancel, _, done := r.Start(script, "test-app")

	// Cancel immediately — should SIGTERM first, then SIGKILL after 3s.
	cancel()

	select {
	case result := <-done:
		if result.ExitCode == 0 {
			t.Errorf("ExitCode: got 0, want non-zero (process was SIGKILL'd)")
		}
		if result.LogFile != "" {
			os.Remove(result.LogFile)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: SIGKILL should have fired within 3s; total wait 5s")
	}
}

// ---------------------------------------------------------------------------
// CheckSudo test
// ---------------------------------------------------------------------------

// TestBashRunner_CheckSudo verifies that CheckSudo returns a bool without panic.
// The actual return value depends on the test environment (may be true or false).
func TestBashRunner_CheckSudo(t *testing.T) {
	r := newRunner(t)
	// Just verify it doesn't panic or block indefinitely.
	done := make(chan bool, 1)
	go func() {
		done <- r.CheckSudo()
	}()

	select {
	case <-done:
		// Pass — we don't assert true/false since test env may not have sudo.
	case <-time.After(10 * time.Second):
		t.Fatal("CheckSudo: timed out after 10 seconds")
	}
}
