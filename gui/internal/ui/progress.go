package ui

import (
	"context"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/kw4jlb/ham-apps/gui/internal/port"
)

// ShowProgressDialog displays a modal progress dialog while a bash script runs.
// It shows a pulsating progress bar, optional inline log viewer, and a cancel
// confirmation flow.
//
// Parameters:
//   - appName:  display name of the app being operated on
//   - action:   "install" or "uninstall"
//   - logFile:  path to the log file written by the runner
//   - cancelFn: two-phase kill function returned by runner.Start
//   - done:     channel that receives one RunResult when the script exits
//   - parent:   parent window (dialog is modal)
//   - onResult: called with the RunResult when the operation completes or is cancelled
func ShowProgressDialog(
	appName, action string,
	logFile string,
	cancelFn func(),
	done <-chan port.RunResult,
	parent fyne.Window,
	onResult func(port.RunResult),
) {
	actionTitle := strings.Title(action) //nolint:staticcheck

	// Progress bar
	progress := widget.NewProgressBarInfinite()

	// Status label
	statusLabel := widget.NewLabel(actionTitle + "ing " + appName + ", please wait...")

	// Log viewer (hidden initially)
	logBinding := binding.NewString()
	_ = logBinding.Set("")
	logEntry := widget.NewEntryWithData(logBinding)
	logEntry.MultiLine = true
	logEntry.Wrapping = fyne.TextWrapOff
	logScroll := container.NewScroll(logEntry)
	logScroll.SetMinSize(fyne.NewSize(560, 200))
	logScroll.Hide()

	logVisible := false

	// Cancel confirmation widgets (hidden initially)
	confirmLabel := widget.NewLabel(
		"Cancel " + action + "? Any files already downloaded may remain on disk.",
	)
	confirmLabel.Wrapping = fyne.TextWrapWord
	confirmLabel.Hide()

	// Context to stop the log tailer goroutine when the dialog closes
	ctx, stopTailer := context.WithCancel(context.Background())

	var d dialog.Dialog
	var keepInstallingBtn, confirmCancelBtn *widget.Button

	keepInstallingBtn = widget.NewButton("Keep "+actionTitle+"ing", func() {
		confirmLabel.Hide()
		keepInstallingBtn.Hide()
		confirmCancelBtn.Hide()
		d.Refresh()
	})
	keepInstallingBtn.Hide()

	confirmCancelBtn = widget.NewButton("Confirm Cancel", func() {
		stopTailer()
		cancelFn()
		d.Hide()
	})
	confirmCancelBtn.Hide()

	// Show Log toggle button
	showLogBtn := widget.NewButton("Show Log", func() {
		logVisible = !logVisible
		if logVisible {
			logScroll.Show()
		} else {
			logScroll.Hide()
		}
		d.Refresh()
	})

	// Cancel button
	cancelBtn := widget.NewButton("Cancel", func() {
		confirmLabel.Show()
		keepInstallingBtn.Show()
		confirmCancelBtn.Show()
		d.Refresh()
	})

	buttonRow := container.NewHBox(showLogBtn, cancelBtn, keepInstallingBtn, confirmCancelBtn)

	content := container.NewVBox(
		statusLabel,
		progress,
		confirmLabel,
		logScroll,
		buttonRow,
	)

	d = dialog.NewCustomWithoutButtons("", content, parent)
	d.Show()

	// Log tailing goroutine — updates every 500 ms while dialog is open
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				data, err := os.ReadFile(logFile)
				if err == nil {
					_ = logBinding.Set(string(data))
				}
			case result := <-done:
				// Script finished
				stopTailer()
				d.Hide()
				onResult(result)
				return
			}
		}
	}()

	// Ensure goroutine stops when dialog is dismissed via X button
	d.SetOnClosed(func() {
		stopTailer()
	})
}
