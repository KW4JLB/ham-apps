package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kw4jlb/ham-apps/gui/internal/app"
)

// UT-22: HAMAPPS_DIR env var not set → ResolveHamappsDir returns filepath.Dir(binaryPath)
func TestResolveHamappsDir_DefaultsToBinaryDir(t *testing.T) {
	// Ensure the env var is not set for this test.
	old, wasSet := os.LookupEnv("HAMAPPS_DIR")
	os.Unsetenv("HAMAPPS_DIR")
	if wasSet {
		defer os.Setenv("HAMAPPS_DIR", old)
	} else {
		defer os.Unsetenv("HAMAPPS_DIR")
	}

	binaryPath := "/usr/local/bin/ham-apps-gui"
	got := app.ResolveHamappsDir(binaryPath)
	want := "/usr/local/bin"
	if got != want {
		t.Errorf("ResolveHamappsDir(%q) = %q; want %q", binaryPath, got, want)
	}
}

// UT-23: HAMAPPS_DIR env var set → ResolveHamappsDir returns that value.
func TestResolveHamappsDir_EnvVarOverrides(t *testing.T) {
	want := "/opt/ham-apps"
	old, wasSet := os.LookupEnv("HAMAPPS_DIR")
	os.Setenv("HAMAPPS_DIR", want)
	if wasSet {
		defer os.Setenv("HAMAPPS_DIR", old)
	} else {
		defer os.Unsetenv("HAMAPPS_DIR")
	}

	got := app.ResolveHamappsDir("/usr/local/bin/ham-apps-gui")
	if got != want {
		t.Errorf("ResolveHamappsDir with HAMAPPS_DIR=%q = %q; want %q", want, got, want)
	}
}

// TestResolveHamappsDir_BinaryInSubDir verifies that the directory component
// is extracted correctly when the binary is in a sub-path.
func TestResolveHamappsDir_BinaryInSubDir(t *testing.T) {
	old, wasSet := os.LookupEnv("HAMAPPS_DIR")
	os.Unsetenv("HAMAPPS_DIR")
	if wasSet {
		defer os.Setenv("HAMAPPS_DIR", old)
	} else {
		defer os.Unsetenv("HAMAPPS_DIR")
	}

	binaryPath := "/home/user/ham-apps/gui/ham-apps-gui"
	got := app.ResolveHamappsDir(binaryPath)
	want := filepath.Dir(binaryPath)
	if got != want {
		t.Errorf("ResolveHamappsDir(%q) = %q; want %q", binaryPath, got, want)
	}
}
