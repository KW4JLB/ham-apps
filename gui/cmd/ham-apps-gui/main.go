// main is the entrypoint for ham-apps-gui.
//
// Usage:
//
//	ham-apps-gui            # launch GUI (default)
//	ham-apps-gui --version  # print version and exit 0
//	ham-apps-gui --help     # print usage and exit 0
package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"

	hamapp "github.com/kw4jlb/ham-apps/gui/internal/app"
	"github.com/kw4jlb/ham-apps/gui/internal/backend"
)

const appID = "io.github.kw4jlb.ham-apps"

const usageText = `Usage: ham-apps-gui [--version] [--help]

Launch the ham-apps graphical application manager.

Flags:
  --version   Print the ham-apps version and exit.
  --help      Print this help message and exit.

Environment:
  HAMAPPS_DIR  Override the ham-apps root directory.
               Defaults to the directory containing the binary.
`

func main() {
	// Simple flag parsing — we only support --version, --help, and no args.
	args := os.Args[1:]

	switch {
	case len(args) == 0:
		// GUI mode — default
		runGUI()

	case len(args) == 1 && (args[0] == "--version" || args[0] == "-v"):
		// Resolve HAMAPPS_DIR to read the version file, but do NOT validate
		// the full directory (avoids requiring a display for --version).
		dir := hamapp.ResolveHamappsDir(os.Args[0])
		repo := &backend.FilesystemRepository{HamappsDir: dir}
		fmt.Println("ham-apps " + repo.ReadVersion())
		os.Exit(0)

	case len(args) == 1 && (args[0] == "--help" || args[0] == "-h"):
		fmt.Print(usageText)
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "ham-apps-gui: unknown argument(s): %v\n\n", args)
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(1)
	}
}

func runGUI() {
	fyneApp := app.NewWithID(appID)

	w, err := hamapp.Bootstrap(os.Args[0], fyneApp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ham-apps-gui: %v\n", err)
		os.Exit(1)
	}

	w.ShowAndRun()
}
