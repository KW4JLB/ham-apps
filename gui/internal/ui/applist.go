// Package ui provides the Fyne-based GUI components for ham-apps.
// Each component accepts port interfaces — no direct imports of
// internal/backend or internal/runner (except the pure FilterApps func).
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/kw4jlb/ham-apps/gui/internal/backend"
	"github.com/kw4jlb/ham-apps/gui/internal/port"
)

// appListState holds mutable state for the AppListWindow.
type appListState struct {
	allApps      []port.AppInfo
	filtered     []port.AppInfo
	selected     int // -1 = nothing selected
	searchText   string
	categoryText string
}

// appRow is the concrete type for each list row, held as a struct so the
// UpdateItem closure can locate sub-widgets by name, not positional index.
type appRow struct {
	icon      *canvas.Image
	nameRich  *widget.RichText
	catLabel  *widget.Label
	badgeRect *canvas.Rectangle
	badgeLbl  *widget.Label
	root      *fyne.Container
}

func newAppRow() *appRow {
	r := &appRow{}

	r.icon = canvas.NewImageFromResource(theme.FyneLogo())
	r.icon.SetMinSize(fyne.NewSize(48, 48))
	r.icon.FillMode = canvas.ImageFillContain

	r.nameRich = widget.NewRichTextFromMarkdown("**App Name**")
	r.catLabel = widget.NewLabel("Category")

	r.badgeRect = canvas.NewRectangle(color.RGBA{R: 108, G: 117, B: 125, A: 200})
	r.badgeRect.SetMinSize(fyne.NewSize(90, 22))
	r.badgeLbl = widget.NewLabel("Not Installed")
	badge := container.NewStack(r.badgeRect, container.NewCenter(r.badgeLbl))

	right := container.NewVBox(r.nameRich, r.catLabel)
	r.root = container.NewBorder(nil, nil, r.icon, badge, right)
	return r
}

// NewAppListWindow constructs and returns the main ham-apps window.
// It accepts port interfaces so the UI has no compile-time dependency on
// concrete backend or runner types.
func NewAppListWindow(repo port.AppRepository, runner port.RunnerService, app fyne.App) fyne.Window {
	w := app.NewWindow("ham-apps — Amateur Radio App Manager")
	w.Resize(fyne.NewSize(960, 540))
	w.SetMaster()

	state := &appListState{selected: -1}

	// --- Load initial data ---
	apps, _ := repo.LoadApps()
	state.allApps = apps
	state.filtered = apps

	cats, _ := repo.LoadCategories()
	catNames := make([]string, 0, len(cats)+1)
	catNames = append(catNames, "All")
	for _, c := range cats {
		catNames = append(catNames, c.DisplayName)
	}

	// --- Install / Uninstall buttons (disabled until selection) ---
	installBtn := widget.NewButton("Install", nil)
	installBtn.Disable()
	uninstallBtn := widget.NewButton("Uninstall", nil)
	uninstallBtn.Disable()

	updateButtons := func() {
		if state.selected < 0 || state.selected >= len(state.filtered) {
			installBtn.Disable()
			uninstallBtn.Disable()
			return
		}
		installBtn.Enable()
		uninstallBtn.Enable()
	}

	showDetail := func(action string) {
		if state.selected < 0 || state.selected >= len(state.filtered) {
			return
		}
		selectedApp := state.filtered[state.selected]
		ShowAppDetailDialog(selectedApp, action, func() {
			// onConfirm: launch install/uninstall flow (wired in task 4.1)
			_ = runner
		}, w)
	}

	// --- Empty-state widgets ---
	emptyLabel := widget.NewLabel("No apps match your search.")
	clearSearchBtn := widget.NewButton("Clear Search", nil)
	emptyState := container.NewVBox(emptyLabel, clearSearchBtn)

	// --- App list widget ---
	var list *widget.List

	list = widget.NewList(
		func() int { return len(state.filtered) },
		func() fyne.CanvasObject {
			return newAppRow().root
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(state.filtered) {
				return
			}
			appInfo := state.filtered[id]

			// Retrieve the appRow embedded in the root container via type assertion.
			// Because Fyne recycles template objects we store a sentinel in the root
			// via its layout — we use a wrapper container to carry the appRow pointer.
			rowWrapper := obj.(*fyne.Container)
			// The root border container is the first (and only) element.
			// We locate sub-widgets by their position in the container tree:
			//   border: Objects[0]=center, Objects[1]=top, Objects[2]=bottom,
			//            Objects[3]=left, Objects[4]=right
			// Actually Fyne's border layout stores: center first, then edges.
			// Use a safer approach: keep a registry keyed by the object pointer.
			updateRowFromContainer(rowWrapper, appInfo, repo)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		state.selected = id
		updateButtons()
	}
	list.OnUnselected = func(id widget.ListItemID) {
		state.selected = -1
		updateButtons()
	}

	listArea := container.NewStack(list, emptyState)

	refreshListArea := func() {
		if len(state.filtered) == 0 {
			list.Hide()
			if state.searchText != "" {
				emptyLabel.SetText("No apps match your search.")
				clearSearchBtn.Show()
			} else {
				emptyLabel.SetText("No apps available.")
				clearSearchBtn.Hide()
			}
			emptyState.Show()
		} else {
			emptyState.Hide()
			list.Show()
			list.Refresh()
		}
		listArea.Refresh()
	}

	applyFilters := func() {
		state.filtered = backend.FilterApps(state.allApps, state.categoryText, state.searchText)
		state.selected = -1
		updateButtons()
		refreshListArea()
	}

	// --- Search bar ---
	search := widget.NewEntry()
	search.SetPlaceHolder("Search apps...")
	search.OnChanged = func(s string) {
		state.searchText = s
		applyFilters()
	}

	// --- Category dropdown ---
	categorySelect := widget.NewSelect(catNames, func(s string) {
		state.categoryText = s
		applyFilters()
	})
	categorySelect.SetSelected("All")

	// --- Clear Search wires ---
	clearSearchBtn.OnTapped = func() {
		search.SetText("")
		state.searchText = ""
		applyFilters()
	}

	// --- Action button handlers ---
	installBtn.OnTapped = func() { showDetail("install") }
	uninstallBtn.OnTapped = func() { showDetail("uninstall") }

	// --- Refresh button ---
	refreshBtn := widget.NewButton("Refresh", func() {
		fresh, _ := repo.LoadApps()
		state.allApps = fresh
		state.selected = -1
		updateButtons()
		applyFilters()
	})

	// --- Enter key → show detail ---
	w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyReturn || ev.Name == fyne.KeyEnter {
			showDetail("install")
		}
	})

	// --- Layout ---
	toolbar := container.NewBorder(
		nil, nil,
		categorySelect,
		container.NewHBox(installBtn, uninstallBtn, refreshBtn),
		search,
	)

	content := container.NewBorder(toolbar, nil, nil, nil, listArea)
	w.SetContent(content)

	// Initial render
	refreshListArea()

	return w
}

// updateRowFromContainer updates the widgets inside a recycled list-row container.
// Fyne's border layout places items in Objects in this order:
//
//	[0] = center content
//	[1] = top (nil placeholder)
//	[2] = bottom (nil placeholder)
//	[3] = left
//	[4] = right
//
// Our row was built as: NewBorder(nil, nil, icon, badge, right-col)
// so: Objects[3]=icon(*canvas.Image), Objects[4]=badge(Stack), Objects[0]=right-col(VBox)
func updateRowFromContainer(root *fyne.Container, appInfo port.AppInfo, repo port.AppRepository) {
	// NewBorder(nil, nil, icon, badge, rightCol) stores objects as:
	//   [0] = rightCol (center/variadic, stored first)
	//   [1] = icon    (left)
	//   [2] = badge   (right)
	// Nil top/bottom are omitted from the slice entirely.
	if len(root.Objects) < 3 {
		return
	}

	// Center = right column VBox
	rightCol, ok := root.Objects[0].(*fyne.Container)
	if !ok || len(rightCol.Objects) < 2 {
		return
	}
	nameRich, ok := rightCol.Objects[0].(*widget.RichText)
	if ok {
		nameRich.ParseMarkdown("**" + appInfo.Name + "**")
	}
	catLbl, ok := rightCol.Objects[1].(*widget.Label)
	if ok {
		catLbl.SetText(appInfo.Category)
	}

	// Left = icon
	iconImg, ok := root.Objects[1].(*canvas.Image)
	if ok {
		iconBytes := repo.LoadIcon(appInfo.Slug)
		if iconBytes != nil {
			res := fyne.NewStaticResource(appInfo.Slug+".png", iconBytes)
			iconImg.Resource = res
		} else {
			iconImg.Resource = theme.FyneLogo()
		}
		iconImg.Refresh()
	}

	// Right = badge Stack
	badgeStack, ok := root.Objects[2].(*fyne.Container)
	if !ok || len(badgeStack.Objects) < 2 {
		return
	}
	badgeRect, ok := badgeStack.Objects[0].(*canvas.Rectangle)
	if ok {
		if appInfo.Installed {
			badgeRect.FillColor = color.RGBA{R: 40, G: 167, B: 69, A: 220}
		} else {
			badgeRect.FillColor = color.RGBA{R: 108, G: 117, B: 125, A: 200}
		}
		badgeRect.Refresh()
	}
	centerBox, ok := badgeStack.Objects[1].(*fyne.Container)
	if ok && len(centerBox.Objects) > 0 {
		badgeLbl, ok := centerBox.Objects[0].(*widget.Label)
		if ok {
			if appInfo.Installed {
				badgeLbl.SetText("Installed")
			} else {
				badgeLbl.SetText("Not Installed")
			}
		}
	}
}
