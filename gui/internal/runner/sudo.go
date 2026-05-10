package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// promptSudo presents a terminal-based password prompt using SUDO_ASKPASS.
// Full GUI dialog integration (Fyne) happens in the UI layer (Task 3.x).
// This implementation creates a minimal askpass shell script, sets SUDO_ASKPASS,
// runs "sudo -A -v" to cache credentials, then securely deletes the temp script.
//
// Security controls (SC-03, SC-06):
//   - Temp script created with os.CreateTemp (mode 0600 by default).
//   - Chmod to 0700 immediately after creation.
//   - Content overwritten with spaces before removal.
//   - signal.Notify handler ensures cleanup on SIGTERM/SIGINT/SIGHUP.
func promptSudo(appName string) error {
	// This is a placeholder implementation. In the full application the UI
	// layer creates the askpass script with the actual Fyne dialog password.
	// Here we wire the security-critical plumbing; the password value is
	// provided by the caller (UI layer) in the real integration.
	return fmt.Errorf("PromptSudo: not yet wired to a GUI dialog (implemented in Task 4.x); appName=%q", appName)
}

// CreateAskpassScript creates a secure temporary askpass script that echoes the
// given password. The caller MUST call the returned cleanup function when done.
// cleanup overwrites the file content before removing it (SC-06).
//
// The returned path should be set as SUDO_ASKPASS before running sudo.
func CreateAskpassScript(password string) (path string, cleanup func(), err error) {
	f, err := os.CreateTemp("", "hamapps-askpass-*.sh")
	if err != nil {
		return "", nil, fmt.Errorf("CreateAskpassScript: CreateTemp: %w", err)
	}
	askpassPath := f.Name()

	// Set mode 0700 (owner execute only) before writing the password.
	if err := f.Chmod(0700); err != nil {
		f.Close()
		os.Remove(askpassPath)
		return "", nil, fmt.Errorf("CreateAskpassScript: chmod: %w", err)
	}

	// Write the askpass script content.
	scriptContent := fmt.Sprintf("#!/bin/bash\necho '%s'\n", sanitizeForShell(password))
	if _, err := f.WriteString(scriptContent); err != nil {
		f.Close()
		os.Remove(askpassPath)
		return "", nil, fmt.Errorf("CreateAskpassScript: write: %w", err)
	}
	f.Close()

	// Register signal handler to ensure cleanup on crash.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	cleanupFn := func() {
		// Overwrite content before removing to clear password from filesystem blocks (SC-06).
		if wf, err := os.OpenFile(askpassPath, os.O_WRONLY, 0); err == nil {
			wf.Write(bytes.Repeat([]byte(" "), 256))
			wf.Close()
		}
		os.Remove(askpassPath)
		signal.Stop(sigCh)
	}

	// Goroutine to handle signals.
	go func() {
		if _, ok := <-sigCh; ok {
			cleanupFn()
		}
	}()

	return askpassPath, cleanupFn, nil
}

// PromptSudoWithAskpass runs "sudo -A -v" with the given askpass script path.
// This is used by the UI layer after it has called CreateAskpassScript.
func PromptSudoWithAskpass(askpassPath string) error {
	cmd := exec.Command("sudo", "-A", "-v")
	cmd.Env = append(os.Environ(), "SUDO_ASKPASS="+askpassPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo -A -v failed: %w (output: %s)", err, string(out))
	}
	return nil
}

// sanitizeForShell escapes single quotes for embedding in a single-quoted shell string.
// This prevents password values containing ' from breaking the script syntax.
func sanitizeForShell(s string) string {
	// In bash, to include a ' in a single-quoted string, end the single-quoted
	// portion, add an escaped quote, then reopen: '\''
	result := ""
	for _, ch := range s {
		if ch == '\'' {
			result += `'\''`
		} else {
			result += string(ch)
		}
	}
	return result
}
