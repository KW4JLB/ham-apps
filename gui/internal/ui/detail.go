package ui

import (
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/kw4jlb/ham-apps/gui/internal/port"
)

// ShowAppDetailDialog presents a modal confirmation dialog for the given
// app and action ("install" or "uninstall"). onConfirm is called if the
// user confirms; it is not called if the user cancels.
func ShowAppDetailDialog(app port.AppInfo, action string, onConfirm func(), parent fyne.Window) {
	titleAction := strings.Title(action) //nolint:staticcheck // simpler than cases.Title for this context
	dialogTitle := titleAction + " " + app.Name

	// Icon (96×96)
	var iconObj fyne.CanvasObject
	if app.IconPath != "" {
		img := canvas.NewImageFromFile(app.IconPath)
		img.SetMinSize(fyne.NewSize(96, 96))
		img.FillMode = canvas.ImageFillContain
		iconObj = img
	} else {
		img := canvas.NewImageFromResource(theme.FyneLogo())
		img.SetMinSize(fyne.NewSize(96, 96))
		img.FillMode = canvas.ImageFillContain
		iconObj = img
	}

	// Metadata labels
	nameLabel := widget.NewLabel("Name: " + app.Name)
	catLabel := widget.NewLabel("Category: " + app.Category)

	// Website hyperlink
	var websiteObj fyne.CanvasObject
	if app.Website != "" {
		parsed, err := url.Parse(app.Website)
		if err == nil {
			websiteObj = widget.NewHyperlink("Website", parsed)
		} else {
			websiteObj = widget.NewLabel("Website: " + app.Website)
		}
	} else {
		websiteObj = widget.NewLabel("Website: (none)")
	}

	// Scrollable description
	descLabel := widget.NewLabel(app.Description)
	descLabel.Wrapping = fyne.TextWrapWord
	descScroll := container.NewScroll(descLabel)
	descScroll.SetMinSize(fyne.NewSize(500, 200))

	content := container.NewVBox(
		iconObj,
		nameLabel,
		catLabel,
		websiteObj,
		widget.NewSeparator(),
		descScroll,
	)

	// Buttons
	var d dialog.Dialog

	confirmBtn := widget.NewButton(titleAction, func() {
		d.Hide()
		onConfirm()
	})
	confirmBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})

	buttons := container.NewHBox(cancelBtn, confirmBtn)
	fullContent := container.NewBorder(nil, buttons, nil, nil, content)

	d = dialog.NewCustomWithoutButtons(dialogTitle, fullContent, parent)
	d.Show()
}
