# Known Bugs — Go + Fyne GUI

Discovered during initial visual QA after Phase 4 implementation. Fix before proceeding to Phase 5.

---

## BUG-01 — App names and categories show placeholder text ✅ FIXED

**Severity**: High — all app rows display "App Name" and "Category" instead of real data

**File**: `gui/internal/ui/applist.go:249`

**Root cause**: `updateRowFromContainer` guards with `len(root.Objects) < 5` but Fyne's `container.NewBorder` does **not** store nil items in the Objects slice. The row was built as `container.NewBorder(nil, nil, icon, badge, rightCol)` which produces only 3 objects (icon, badge, rightCol), not 5. The guard always fires and returns early before updating any widget text.

**Fix**: Either remove the nil arguments so all 5 positions are non-nil objects, or navigate the container by type-asserting the known 3 objects directly (index 0 = rightCol VBox, index 1 = icon, index 2 = badge Stack) and adjust the guard to `< 3`.

---

## BUG-02 — Search entry is too small and does not expand ✅ FIXED

**Severity**: Medium — search bar is cramped and unusable

**File**: `gui/internal/ui/applist.go:221`

**Root cause**: The toolbar is built as:
```go
container.NewBorder(nil, nil,
    container.NewHBox(search, categorySelect),  // left side — fixed width
    container.NewHBox(installBtn, ...),          // right side
    nil,                                          // no center to absorb remaining space
)
```
An `HBox` placed on the left of a border layout takes only its natural (minimum) width. There is no center widget to absorb the remaining horizontal space, so the search entry never grows.

**Fix**: Give the search entry a dedicated expanding region. One approach: place `search` as the center of the border (it will stretch to fill available space), and put only `categorySelect` with the buttons on the right. Another approach: use `container.NewGridWithColumns` or `layout.NewFormLayout` for the toolbar instead.

---

## BUG-03 — No app icons (content gap, not a code bug)

**Severity**: Low — placeholder Fyne logo shown for all apps

**Root cause**: None of the apps in `apps/*/` currently have an `icon.png` file. The fallback to `theme.FyneLogo()` is working as designed.

**Fix**: Add `icon.png` (64×64 or 128×128 PNG) to each app directory. This is a content task, not a code change.

---

## BUG-04 — No ham-apps application icon ✅ FIXED

**Severity**: Low — default Fyne window icon shown in taskbar

**Root cause**: The main entrypoint does not call `app.SetIcon(...)`. Embedding an icon requires a bundled resource (generated via `fyne bundle`).

**Fix**: Create or source a ham-apps logo PNG, run `fyne bundle -o gui/internal/app/appicon.go icon.png`, then call `fyneApp.SetIcon(appicon.ResourceIconPng)` in `main.go` or `bootstrap.go`.
