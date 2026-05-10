package port

// RunnerService is the abstraction consumed by the UI layer for executing
// install/uninstall scripts and managing sudo credentials.
type RunnerService interface {
	// Start executes the given bash script asynchronously with slug as an
	// argument. It returns a cancel function, the path of the live log file
	// (created before the process starts, so callers can tail it immediately),
	// and a channel that receives exactly one RunResult when the process exits
	// or is killed.
	Start(script, slug string) (cancel func(), logFile string, done <-chan RunResult)

	// CheckSudo returns true if sudo credentials are already cached
	// (i.e. "sudo -n true" exits 0).
	CheckSudo() bool

	// PromptSudo presents a GUI password dialog, configures SUDO_ASKPASS,
	// and runs "sudo -A -v" to cache credentials. Returns an error if the
	// user cancels or the password is incorrect.
	PromptSudo(appName string) error
}
