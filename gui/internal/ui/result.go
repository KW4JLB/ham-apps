package ui

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowSuccessDialog presents a modal dialog reporting a successful install or
// uninstall. onClose is called when the user dismisses the dialog.
func ShowSuccessDialog(appName, action string, parent fyne.Window, onClose func()) {
	actionTitle := strings.Title(action) //nolint:staticcheck
	pastTense := actionTitle + "ed"

	titleText := appName + " " + pastTense
	bodyText := appName + " was " + strings.ToLower(pastTense) + " successfully."

	bodyLabel := widget.NewLabel(bodyText)
	bodyLabel.Wrapping = fyne.TextWrapWord

	var d dialog.Dialog
	closeBtn := widget.NewButton("Close", func() {
		d.Hide()
		onClose()
	})

	content := container.NewBorder(
		nil,
		container.NewCenter(closeBtn),
		nil, nil,
		bodyLabel,
	)

	d = dialog.NewCustomWithoutButtons(titleText, content, parent)
	d.SetOnClosed(onClose)
	d.Show()
}

// ShowErrorDialog presents a modal dialog reporting a failed install or
// uninstall. It includes inline log viewer, clipboard copy, and sensitive-data
// notice. cleanup deletes the log file; it is called when the dialog closes.
func ShowErrorDialog(
	appName, action string,
	exitCode int,
	logFile string,
	cleanup func(),
	parent fyne.Window,
	onClose func(),
) {
	actionTitle := strings.Title(action) //nolint:staticcheck
	titleText := fmt.Sprintf("%s Failed — %s", actionTitle, appName)
	bodyText := fmt.Sprintf(
		"%s could not be %sed. Exit code: %d. Check the log for details.",
		appName, strings.ToLower(action), exitCode,
	)
	noticeText := "Output may contain sensitive information."

	bodyLabel := widget.NewLabel(bodyText)
	bodyLabel.Wrapping = fyne.TextWrapWord

	noticeLabel := widget.NewLabel(noticeText)
	noticeLabel.Wrapping = fyne.TextWrapWord

	// Inline log viewer (hidden initially)
	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logScroll := container.NewScroll(logEntry)
	logScroll.SetMinSize(fyne.NewSize(560, 200))
	logScroll.Hide()

	logVisible := false

	cleanupOnce := makeOnce(cleanup)

	var d dialog.Dialog

	showLogBtn := widget.NewButton("Show Log", func() {
		if !logVisible {
			data, err := os.ReadFile(logFile)
			if err == nil {
				logEntry.SetText(string(data))
			} else {
				logEntry.SetText("(could not read log file)")
			}
			logScroll.Show()
			logVisible = true
		} else {
			logScroll.Hide()
			logVisible = false
		}
		d.Refresh()
	})

	copyLogBtn := widget.NewButton("Copy Log", func() {
		data, err := os.ReadFile(logFile)
		if err == nil {
			parent.Clipboard().SetContent(string(data))
		}
	})

	closeBtn := widget.NewButton("Close", func() {
		cleanupOnce()
		d.Hide()
		onClose()
	})

	buttons := container.NewHBox(showLogBtn, copyLogBtn, closeBtn)

	content := container.NewVBox(
		bodyLabel,
		noticeLabel,
		widget.NewSeparator(),
		logScroll,
		container.NewCenter(buttons),
	)

	d = dialog.NewCustomWithoutButtons(titleText, content, parent)
	d.SetOnClosed(func() {
		cleanupOnce()
		onClose()
	})
	d.Show()
}

// ShowCancelledDialog presents a modal dialog when an operation was cancelled
// by the user. cleanup deletes the log file on close; onClose is called after.
func ShowCancelledDialog(
	appName, action string,
	logFile string,
	cleanup func(),
	parent fyne.Window,
	onClose func(),
) {
	_ = appName // used in body text
	_ = action

	titleText := "Cancelled"
	bodyText := "Installation was cancelled. You can try again later. Some temporary files may remain."

	bodyLabel := widget.NewLabel(bodyText)
	bodyLabel.Wrapping = fyne.TextWrapWord

	// Inline log viewer (hidden initially)
	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapOff
	logScroll := container.NewScroll(logEntry)
	logScroll.SetMinSize(fyne.NewSize(560, 200))
	logScroll.Hide()

	logVisible := false

	cleanupOnce := makeOnce(cleanup)

	var d dialog.Dialog

	showLogBtn := widget.NewButton("Show Log", func() {
		if !logVisible {
			data, err := os.ReadFile(logFile)
			if err == nil {
				logEntry.SetText(string(data))
			} else {
				logEntry.SetText("(could not read log file)")
			}
			logScroll.Show()
			logVisible = true
		} else {
			logScroll.Hide()
			logVisible = false
		}
		d.Refresh()
	})

	closeBtn := widget.NewButton("Close", func() {
		cleanupOnce()
		d.Hide()
		onClose()
	})

	buttons := container.NewHBox(showLogBtn, closeBtn)

	content := container.NewVBox(
		bodyLabel,
		widget.NewSeparator(),
		logScroll,
		container.NewCenter(buttons),
	)

	d = dialog.NewCustomWithoutButtons(titleText, content, parent)
	d.SetOnClosed(func() {
		cleanupOnce()
		onClose()
	})
	d.Show()
}

// makeOnce returns a function that calls f exactly once, regardless of how
// many times the returned function is invoked.
func makeOnce(f func()) func() {
	var called bool
	return func() {
		if !called {
			called = true
			if f != nil {
				f()
			}
		}
	}
}
