// Package port defines the pure interface and type layer for ham-apps-gui.
// It has no dependencies on Fyne, os, or os/exec — it is the abstraction
// boundary between the UI layer and the infrastructure layer.
package port

// AppInfo represents a single app's parsed metadata.
type AppInfo struct {
	Slug        string
	Name        string
	Category    string
	Website     string
	Tags        []string
	MinOS       string
	Description string // full text from description file
	Summary     string // first line of description
	Installed   bool
	IconPath    string // empty if no icon.png
}

// Category represents a category from data/categories.
type Category struct {
	ID          string
	DisplayName string
	Description string
}

// RunResult holds the outcome of a runner.Start invocation.
type RunResult struct {
	ExitCode int
	LogFile  string
	Err      error
}
