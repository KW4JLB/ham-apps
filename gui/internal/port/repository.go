package port

// AppRepository is the abstraction consumed by the UI layer for reading
// and writing app metadata and install state.
type AppRepository interface {
	// LoadApps returns all apps found in the hamappsDir. The second return
	// value collects per-app errors (e.g. missing metadata) without aborting
	// the entire load; callers should log errors and display remaining apps.
	LoadApps() ([]AppInfo, []error)

	// LoadCategories returns all categories defined in data/categories.
	LoadCategories() ([]Category, error)

	// IsInstalled reports whether the given slug is currently installed
	// (i.e. its state file exists under the status directory).
	IsInstalled(slug string) bool

	// MarkInstalled creates the install-state file for the given slug.
	MarkInstalled(slug string) error

	// MarkUninstalled removes the install-state file for the given slug.
	// It returns nil if the file is already absent.
	MarkUninstalled(slug string) error

	// ReadVersion returns the version string from the version file,
	// or "dev" if the file is absent or unreadable.
	ReadVersion() string

	// LoadIcon returns the raw bytes of the icon file for the given slug,
	// or nil if the icon is absent or the path escapes the apps/ directory.
	LoadIcon(slug string) []byte

	// ScriptsDir returns the absolute path to the scripts/ directory inside
	// the ham-apps root. Used by the UI layer to construct script paths for
	// the runner without needing direct knowledge of HAMAPPS_DIR.
	ScriptsDir() string
}
