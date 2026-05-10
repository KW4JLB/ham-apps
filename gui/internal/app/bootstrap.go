// Package app is the dependency-injection root for ham-apps-gui.
// It resolves HAMAPPS_DIR, validates it, wires concrete backend and runner
// implementations into the UI layer, and returns the main Fyne window.
//
// Dependency direction (one-way only):
//
//	app → ui, app → backend, app → runner
//
// The UI layer has no compile-time dependency on backend or runner; it only
// imports the port interfaces.
package app

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"

	"github.com/kw4jlb/ham-apps/gui/internal/backend"
	"github.com/kw4jlb/ham-apps/gui/internal/runner"
	"github.com/kw4jlb/ham-apps/gui/internal/ui"
)

// ResolveHamappsDir returns the ham-apps root directory.
// It checks the HAMAPPS_DIR environment variable first; if that is unset or
// empty it falls back to the directory that contains the running binary
// (filepath.Dir(binaryPath)).
func ResolveHamappsDir(binaryPath string) string {
	if dir := os.Getenv("HAMAPPS_DIR"); dir != "" {
		return dir
	}
	return filepath.Dir(binaryPath)
}

// Bootstrap resolves HAMAPPS_DIR, validates it, creates the concrete
// implementations of port.AppRepository and port.RunnerService, wires them
// into the UI layer, and returns the main window ready to be shown.
//
// binaryPath is typically os.Args[0].
// fyneApp is the Fyne application instance (created by the caller so that
// fyne.NewApp is only called once in the process).
//
// Returns an error if HAMAPPS_DIR cannot be resolved to a valid ham-apps
// installation (e.g. missing apps/ or scripts/ subdirectories). In that case
// no window is created and the caller should print the error and exit 1.
func Bootstrap(binaryPath string, fyneApp fyne.App) (fyne.Window, error) {
	dir := ResolveHamappsDir(binaryPath)

	if err := backend.ValidateHamappsDir(dir); err != nil {
		return nil, fmt.Errorf("invalid HAMAPPS_DIR %q: %w", dir, err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine user home directory: %w", err)
	}
	statusDir := filepath.Join(homeDir, ".local", "share", "ham-apps", "installed")

	repo := &backend.FilesystemRepository{
		HamappsDir: dir,
		StatusDir:  statusDir,
	}

	bashRunner := &runner.BashRunner{}

	w := ui.NewAppListWindow(repo, bashRunner, fyneApp)
	return w, nil
}
