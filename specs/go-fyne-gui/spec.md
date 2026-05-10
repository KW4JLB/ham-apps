# Go + Fyne GUI Front-End for ham-apps

## Overview

### Purpose
Replace the two yad-based bash GUI scripts (`gui/app-list` and `gui/app-details`) with a single, self-contained Go binary built using the Fyne toolkit. The new binary (`gui/ham-apps-gui`) compiles to a single statically-linked executable, ships its own OpenGL renderer, and has no system GUI package requirements beyond `libGL` and `libX11`, which are present on every desktop Linux installation.

### Scope
- New Go module: `gui/` (source) → compiled binary `gui/ham-apps-gui`
- Replaces `gui/app-list` and `gui/app-details` (kept as legacy fallback during transition, removed after validation)
- Updates `ham-apps` entry point to invoke `gui/ham-apps-gui` instead of `gui/app-list`
- All bash backend scripts (`scripts/install-app`, `scripts/uninstall-app`, `scripts/list-apps`, `scripts/utils`) remain entirely unchanged
- Install-state tracking, app metadata filesystem layout, and `HAMAPPS_DIR` env var contract remain unchanged

### Background
ham-apps is a pure-bash app manager for Amateur Radio software on Debian/Ubuntu, inspired by pi-apps. The existing yad GUI requires a system package (`yad`) and is constrained by yad's GTK2/3 list widget. Go + Fyne was selected because:
- Compiles to a single binary; zero additional system packages required
- Ships its own OpenGL-based renderer (no system GTK, Qt, or WebView)
- Only system libraries needed: `libGL` (mesa) and `libX11`, universally present on any desktop Linux
- Tauri was explicitly ruled out: depends on system WebKitGTK
- Target: Debian 11+ and Ubuntu 20.04+, all desktop architectures (amd64, arm64)

---

## Requirements

### Functional Requirements

| ID     | Requirement | Priority |
|--------|-------------|----------|
| FR-01  | App list window: display all available apps in a scrollable list/card view with icon, name, category, install status badge, and short description | Critical |
| FR-02  | Search bar: real-time text search filtering across app name, category, and description fields | Critical |
| FR-03  | Category filter: dropdown or sidebar to filter by category; populated from `data/categories`; includes "All" option | Critical |
| FR-04  | Install action: selecting an app and clicking Install shows a confirmation dialog with full app metadata, then runs `scripts/install-app <slug>` | Critical |
| FR-05  | Uninstall action: selecting an app and clicking Uninstall shows a confirmation dialog, then runs `scripts/uninstall-app <slug>` | Critical |
| FR-06  | Progress indicator: during install/uninstall, display a modal progress dialog with pulsating indicator, app name, action label, and a Cancel button | Critical |
| FR-07  | Cancel install/uninstall: Cancel button kills the background process; install-state file is NOT written on cancel | Critical |
| FR-08  | Status badges: apps display "Installed" (green) or "Not installed" (grey) badges; state read from `~/.local/share/ham-apps/installed/<slug>` | High |
| FR-09  | Refresh: a Refresh button or menu item re-reads install state and redraws the list without closing the window | High |
| FR-10  | App detail view: confirmation dialog shows name, category, website (clickable link), full description, and icon (fallback to built-in placeholder if `icon.png` absent) | High |
| FR-11  | Sudo password prompt: if `sudo -n true` fails before running install/uninstall, prompt for password in a GUI dialog (not terminal); use `SUDO_ASKPASS` pattern | High |
| FR-12  | Success notification: after install/uninstall exits 0, show a success dialog with app name and action | High |
| FR-13  | Error notification: after install/uninstall exits non-zero, show an error dialog with app name, action, and exit code; preserve log | High |
| FR-14  | `HAMAPPS_DIR` env var: binary reads `HAMAPPS_DIR` to locate the app root; defaults to the directory containing the binary | High |
| FR-15  | Binary location: compiled binary placed at `gui/ham-apps-gui`; `ham-apps` entry point updated to call it | High |
| FR-16  | Window title: "ham-apps — Amateur Radio App Manager" | Medium |
| FR-17  | Window minimum size: 960×540; resizable; maximizable | Medium |
| FR-18  | Icon display: load `apps/<slug>/icon.png` (64×64 or 128×128); show placeholder if absent | Medium |
| FR-19  | Log output: capture stdout/stderr of install/uninstall script to a temp file; offer "Show Log" button in result dialog | Medium |
| FR-20  | CLI passthrough: binary exits immediately with usage if called with unrecognised args; GUI mode is the default (no args) | Low |

### Non-Functional Requirements

| ID      | Requirement | Target |
|---------|-------------|--------|
| NFR-01  | Binary size: self-contained, stripped | ≤ 25 MB |
| NFR-02  | Startup time: window visible | ≤ 2 s on target hardware |
| NFR-03  | App list render time for 50 apps | ≤ 500 ms |
| NFR-04  | No system GUI packages required beyond libGL + libX11 | 0 new apt dependencies |
| NFR-05  | Go version | ≥ 1.21 |
| NFR-06  | Fyne version | ≥ 2.4 |
| NFR-07  | CGO required for Fyne OpenGL renderer | CGO_ENABLED=1 |
| NFR-08  | Build reproducible via `go build` with pinned `go.sum` | Deterministic |
| NFR-09  | All Go code passes `go vet` and `golangci-lint` with zero errors | 0 errors |
| NFR-10  | Test coverage for non-GUI logic (metadata parsing, state reading) | ≥ 80% |
| NFR-11  | Binary must run on Debian 11, Debian 12, Ubuntu 20.04, Ubuntu 22.04, Ubuntu 24.04 | All 5 targets |
| NFR-12  | Existing bash tests must continue to pass after `ham-apps` entry point update | 100% pass |

### Constraints

- Go + Fyne only; no Python, no Electron, no Tauri, no web server
- CGO is required (Fyne's OpenGL renderer needs it); build environment must have `gcc` and GL headers
- The bash backend scripts are read-only — no modifications
- `HAMAPPS_DIR` is the single integration point for locating app data; must be validated as an absolute path with expected directory structure on startup
- Install-state directory: `~/.local/share/ham-apps/installed/<slug>` (read via filesystem, not via bash scripts)
- App metadata: `apps/<slug>/metadata` key=value format; `apps/<slug>/description` plain text; `apps/<slug>/icon.png` optional
- Categories: `data/categories` pipe-delimited (`id|display-name|description`)
- No hardcoded version strings — version read from `$HAMAPPS_DIR/version` file at runtime; embed via `go:embed` or runtime read
- All Go source lives under `gui/` as a standalone Go module with module path `github.com/kw4jlb/ham-apps/gui`; all `go` tool invocations run from the `gui/` directory
- Binary output path: `gui/ham-apps-gui`
- Tests in `tests/` follow existing bash naming convention for integration/smoke tests; Go unit tests live in `gui/*_test.go`
- Install scripts MUST NOT echo credentials, tokens, or passwords to stdout (convention enforced in CONTRIBUTING.md)
- `go.sum` MUST be committed; any PR that modifies `go.sum` requires review
- Process-group kill (Linux `Setpgid`) uses a `//go:build linux` build tag; non-Linux platforms use single-process kill stub

---

## Design

### Architecture

```
ham-apps (bash entry point)
  └─ gui/ham-apps-gui  (Go binary — replaces gui/app-list + gui/app-details)
        ├─ internal/app/        ← Fyne application bootstrap, window management
        ├─ internal/backend/    ← filesystem I/O: read metadata, parse categories, check install state
        ├─ internal/runner/     ← exec.Cmd wrapper: runs bash install/uninstall scripts, captures output
        ├─ internal/ui/         ← Fyne widget composition: list view, detail dialog, progress dialog
        └─ main.go              ← entrypoint: parse HAMAPPS_DIR, launch app
```

#### Component Responsibilities

**`internal/backend`** — pure Go, no Fyne dependency, fully testable without display:
- `LoadApps(hamappsDir string) ([]AppInfo, []error)` — scan `apps/*/` directories, parse metadata and description; returns partial results on per-app failures (invalid slugs or missing metadata are logged as warnings and skipped; remaining apps are returned)
- `LoadCategories(hamappsDir string) ([]Category, error)` — parse `data/categories`
- `IsInstalled(statusDir, slug string) bool` — check `~/.local/share/ham-apps/installed/<slug>`
- `MarkInstalled(statusDir, slug string) error` — create state file
- `MarkUninstalled(statusDir, slug string) error` — remove state file
- `ReadVersion(hamappsDir string) string` — read `version` file; returns `"dev"` if absent
- `LoadIcon(hamappsDir, slug string) []byte` — load `apps/<slug>/icon.png` or return nil
- `ValidateSlug(slug string) error` — enforces `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- `ValidateHamappsDir(dir string) error` — asserts absolute path; checks `apps/`, `scripts/`, `data/categories` exist

**`internal/port`** — interface definitions; no Fyne or filesystem dependency:
```go
// AppRepository is the abstraction consumed by the UI layer
type AppRepository interface {
    LoadApps() ([]AppInfo, []error)
    LoadCategories() ([]Category, error)
    IsInstalled(slug string) bool
    MarkInstalled(slug string) error
    MarkUninstalled(slug string) error
    ReadVersion() string
    LoadIcon(slug string) []byte
}

// RunnerService is the abstraction consumed by the UI layer
type RunnerService interface {
    Start(script, slug string) (cancel func(), done <-chan RunResult)
    CheckSudo() bool
    PromptSudo(appName string) error
}
```

**`internal/runner`** — implements `port.RunnerService`; exec.Cmd wrapper:
- `BashRunner` concrete implementation using `exec.Cmd`
- Process-group kill (Linux): `//go:build linux` file sets `Setpgid: true`; cancel sends SIGTERM, then SIGKILL after 3 s via `time.AfterFunc`
- Non-Linux stub: `//go:build !linux` file calls `cmd.Process.Kill()` immediately
- `SudoChecker` — wraps `sudo -n true` check; returns bool
- `AskPassRunner` — displays Fyne password dialog; writes askpass temp script with `os.CreateTemp` (mode 0700); registers cleanup that overwrites content before removing; installs signal handler (SIGTERM, SIGINT, SIGHUP) to ensure cleanup on crash

**`internal/ui`** — Fyne widget layer; depends on `internal/port` interfaces only (not `internal/backend` directly):
- `AppListWindow` — main window; contains search bar, category selector, app list, action buttons; displays empty-state widget when filtered list is empty
- `AppDetailDialog` — modal confirm dialog for install/uninstall
- `ProgressDialog` — modal progress with inline log viewer; Cancel button shows inline confirmation before killing process
- `ResultDialog` — success/error result with inline log viewer; error dialog specifies content explicitly (see below)

**ResultDialog content specification**:
- Success:
  - Title: "[App Name] Installed" (or "Uninstalled")
  - Body: "[App Name] was installed successfully."
  - Buttons: [Close]
- Error:
  - Title: "Installation Failed — [App Name]" (or "Uninstall Failed")
  - Body: "[App Name] could not be installed. Exit code: [N]. Check the log for details."
  - Buttons: [Show Log] [Copy Log] [Close]
  - Show Log expands an inline scrollable text area within the same dialog
  - Copy Log copies log contents to clipboard for support sharing
  - Log file deleted when dialog is closed (deferred cleanup)

**`internal/app`** — wires backend + runner + ui together:
- `Bootstrap(hamappsDir string) *fyne.App` — validates `HAMAPPS_DIR`; creates Fyne app; instantiates `backend.FilesystemRepository` and `runner.BashRunner`; passes them as `port.AppRepository` and `port.RunnerService` to the UI layer

#### Data Flow

```
startup
  └─ backend.LoadApps()        → []AppInfo
  └─ backend.LoadCategories()  → []Category
  └─ ui.AppListWindow.Render() → displays list

user clicks Install on app X
  └─ ui.AppDetailDialog.Show(appX)
      └─ user confirms
          └─ runner.SudoChecker.Check()
              ├─ if needs password → ui.PasswordDialog.Prompt() → sets SUDO_ASKPASS
              └─ runner.BashRunner.Run("install-app", "x") → backgrounded
                  └─ ui.ProgressDialog.Show(cancel func)
                      └─ on exit 0 → ui.ResultDialog.ShowSuccess()
                      └─ on exit != 0 → ui.ResultDialog.ShowError(logFile)
                      └─ on cancel → kill process, no state change
```

#### Install State Lifecycle

The Go binary reads install state directly from the filesystem (`~/.local/share/ham-apps/installed/<slug>`), the same as the bash scripts. It does NOT call `mark_installed` / `mark_uninstalled` bash functions directly — instead it replicates the contract:
- On successful install: `os.MkdirAll(statusDir, 0755); os.WriteFile(path.Join(statusDir, slug), []byte{}, 0644)`
- On successful uninstall: `os.Remove(path.Join(statusDir, slug))`
- On cancel or error: no state change

#### HAMAPPS_DIR Validation

On startup, after resolving `HAMAPPS_DIR`, the binary calls `backend.ValidateHamappsDir(dir)`:
```go
func ValidateHamappsDir(dir string) error {
    if !filepath.IsAbs(dir) {
        return fmt.Errorf("HAMAPPS_DIR must be an absolute path, got: %q", dir)
    }
    required := []string{"apps", "scripts", filepath.Join("data", "categories")}
    for _, rel := range required {
        if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
            return fmt.Errorf("HAMAPPS_DIR %q missing required path %q", dir, rel)
        }
    }
    return nil
}
```
If validation fails, the binary logs the error and exits 1 with a clear message before opening any window.

#### `ham-apps` Entry Point Update

The bash `ham-apps` script's `gui|""` case changes from:
```bash
"$HAMAPPS_DIR/gui/app-list"
```
to:
```bash
"$HAMAPPS_DIR/gui/ham-apps-gui"
```

No other changes to the entry point.

#### Build System

A `Makefile` in `gui/` handles:
```
make build   → CGO_ENABLED=1 go build -ldflags="-s -w" -o ham-apps-gui ./cmd/ham-apps-gui
make test    → go vet ./... && go test ./...
make lint    → golangci-lint run
make check   → make test && for t in ../tests/test-*; do bash "$t"; done
make clean   → rm -f ham-apps-gui
```

The Go module root is `gui/` with module path `github.com/kw4jlb/ham-apps/gui`. All `go` tool invocations (in Makefile, CI, developer workflow) are run from the `gui/` directory. `go.mod` and `go.sum` are checked into the repository. Any PR modifying `go.sum` requires explicit review.

#### CI/CD Pipeline

A new GitHub Actions workflow `.github/workflows/go-gui.yml` is added to the repository:

```yaml
name: Go GUI Build & Test
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Install build dependencies
        run: sudo apt-get install -y gcc libgl1-mesa-dev libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev
      - name: Go vet
        run: cd gui && go vet ./...
      - name: Go unit tests
        run: cd gui && go test ./...
      - name: Go vulnerability check
        run: cd gui && go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...
      - name: Build binary
        run: cd gui && CGO_ENABLED=1 go build -o ham-apps-gui ./cmd/ham-apps-gui
      - name: Verify --version
        run: HAMAPPS_DIR="$PWD" gui/ham-apps-gui --version
      - name: Bash tests
        run: for t in tests/test-*; do bash "$t"; done
```

#### Binary Distribution

Pre-built binaries for `linux/amd64` and `linux/arm64` are published as GitHub Release assets on each tagged release. The file naming convention is `ham-apps-gui-linux-amd64` and `ham-apps-gui-linux-arm64`.

The `install.sh` script is updated to:
1. Detect host architecture (`uname -m`)
2. Download the appropriate pre-built binary from the latest GitHub Release
3. Place it at `gui/ham-apps-gui` with mode `755`
4. Fall back to `make -C gui build` if Go toolchain is present and the pre-built binary is unavailable

This preserves the existing zero-build-tools-required user experience for standard amd64 and arm64 Debian/Ubuntu systems.

### UI Layout

#### Main Window (AppListWindow)

```
┌─────────────────────────────────────────────────────────────────┐
│ ham-apps — Amateur Radio App Manager              [_] [□] [X]   │
├─────────────────────────────────────────────────────────────────┤
│ [Search...                        ] [Category: All    ▼] [Refresh] │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ [icon] WSJT-X          digital-modes  [● Installed  ]       │ │
│ │        Weak signal digital modes ...                        │ │
│ ├─────────────────────────────────────────────────────────────┤ │
│ │ [icon] Fldigi          digital-modes  [○ Not Installed]     │ │
│ │        Multi-mode software modem ...                        │ │
│ ├─────────────────────────────────────────────────────────────┤ │
│ │ [icon] Direwolf        packet-aprs    [○ Not Installed]     │ │
│ │        Soundcard-based TNC ...                              │ │
│ └─────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                          [Install]  [Uninstall]                 │
└─────────────────────────────────────────────────────────────────┘
```

- Each row: icon (48×48), name (bold), category badge, status badge, description (truncated)
- Status badge: green background "Installed" or grey "Not Installed"
- Search bar filters in real time as user types (no button click needed)
- Category dropdown: "All" + each unique category id from `data/categories` (showing display-name)
- Refresh button re-reads filesystem state
- **Empty state**: when no apps match the current filter/search, display a centered message within the list area instead of a blank list:
  - Search mismatch: "No apps match '[query]'. Clear search to see all apps." with a [Clear Search] button
  - Category empty: "No apps in '[Category]'. Select a different category."
  - No apps at all: "No apps found. Check that HAMAPPS_DIR is set correctly."
- **Row interaction model**:
  - Single click: selects app; highlights row; enables Install and Uninstall buttons
  - Double-click or Enter (when list has focus): opens `AppDetailDialog` for the selected app
  - No app selected: Install and Uninstall buttons are disabled (greyed out); no error dialog shown
  - Install/Uninstall button clicks always route through `AppDetailDialog` for confirmation
- **Keyboard navigation** (Tab order and key bindings):
  - Tab order: Search bar → Category dropdown → App list → Install button → Uninstall button → Refresh button
  - Up/Down arrows in list: move selection
  - Enter on focused list item: open detail dialog
  - Escape: close any open dialog; clear search text if search bar has focus

#### App Detail Dialog (AppDetailDialog)

```
┌─────────────────────────────────────────────────────────┐
│ Install WSJT-X                                      [X] │
├─────────────────────────────────────────────────────────┤
│  [icon 96x96]   WSJT-X                                  │
│                 Category: Digital Modes                  │
│                 Website: https://wsjt.sourceforge.io    │
│                                                         │
│  Weak signal digital modes by K1JT — the standard      │
│  for FT8, FT4, and WSPR...                             │
│                                                         │
├─────────────────────────────────────────────────────────┤
│                          [Cancel]  [Install]            │
└─────────────────────────────────────────────────────────┘
```

- Website rendered as a hyperlink widget (Fyne `widget.Hyperlink`)
- Icon: `apps/<slug>/icon.png` if present; built-in placeholder otherwise
- Action button labeled "Install" or "Uninstall" depending on action

#### Progress Dialog

```
┌─────────────────────────────────────────────────────────┐
│ Installing WSJT-X                                   [X] │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ████████████████░░░░░░░░░░░░░░░  (pulsating)          │
│                                                         │
│  Installing WSJT-X, please wait...                     │
│                                                         │
├─────────────────────────────────────────────────────────┤
│                    [Show Log]  [Cancel]                 │
└─────────────────────────────────────────────────────────┘
```

- Pulsating progress bar (Fyne `widget.ProgressBarInfinite`)
- Show Log button opens a scrollable inline text area within the progress dialog (not a new window)
- Cancel button: shows an inline confirmation before killing the process:
  ```
  Cancel installation? Any files already downloaded may remain on disk.
  [Keep Installing]  [Confirm Cancel]
  ```
  After confirmed cancel: kills process group (two-phase SIGTERM → SIGKILL), then shows result dialog:
  "Installation was cancelled. You can try again later. Some temporary files may remain." with [Show Log] [Close]

### Data Models

```go
// AppInfo represents a single app's parsed metadata
type AppInfo struct {
    Slug        string
    Name        string
    Category    string
    Website     string
    Tags        []string
    MinOS       string
    Description string   // full text from description file
    Summary     string   // first line of description
    Installed   bool
    IconPath    string   // empty if no icon.png
}

// Category represents a category from data/categories
type Category struct {
    ID          string
    DisplayName string
    Description string
}
```

### Metadata Parsing

`apps/<slug>/metadata` is `key=value` format, one per line. Parser:
1. Open file; iterate lines
2. Split on first `=`; trim whitespace
3. Populate `AppInfo` fields by key name (`name`, `category`, `website`, `tags`, `min-os`)
4. `tags` field split by comma

`data/categories` is pipe-delimited (`id|display-name|description`). Lines starting with `#` are comments.

### Sudo / Privilege Escalation

The Go binary uses the same `SUDO_ASKPASS` pattern as the existing `gui/app-details` bash script:
1. Run `sudo -n true` (non-interactive check)
2. If exit 0: credentials cached, proceed directly
3. If exit non-zero: show Fyne password dialog; write temp script to `mktemp` location; set `SUDO_ASKPASS` env var on the `exec.Cmd`; run `sudo -A -v`; delete temp script
4. The install/uninstall script itself runs `sudo` as needed — the GUI does not run as root

### Process Management

```go
type RunResult struct {
    ExitCode int
    LogFile  string
    Err      error
}

// Start runs the script asynchronously; returns a cancel func and a channel
func (r *BashRunner) Start(script, slug string) (cancel func(), done <-chan RunResult)
```

- Linux only (`//go:build linux`): `exec.Cmd.SysProcAttr` sets `Setpgid: true` so `cancel()` can kill the entire process group
- Non-Linux stub (`//go:build !linux`): `cancel()` calls `cmd.Process.Kill()` directly
- `cancel()` two-phase kill (Linux):
  ```go
  func cancel() {
      syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
      timer := time.AfterFunc(3*time.Second, func() {
          syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
      })
      defer timer.Stop()
  }
  ```
- stdout and stderr redirected to `os.CreateTemp("", "hamapps-*.log")` with mode 0600
- Log file is deleted in a deferred cleanup after the result dialog is dismissed (not kept in `/tmp` indefinitely)
- A goroutine waits on `cmd.Wait()`, then sends `RunResult` on the `done` channel

---

## Test Specification

### Go Unit Tests (`gui/*_test.go`)

#### UT-01: metadata parser — all fields populated
- **Given** a temp `metadata` file with all standard keys
- **When** `backend.ParseMetadata(path)` is called
- **Then** returned `AppInfo` has correct Name, Category, Website, Tags, MinOS

#### UT-02: metadata parser — missing optional fields
- **Given** a `metadata` file with only `name=` and `category=`
- **When** `backend.ParseMetadata(path)` is called
- **Then** Website, Tags, MinOS are empty strings / nil slice; no error

#### UT-03: metadata parser — malformed line ignored
- **Given** a `metadata` file with one line missing `=`
- **When** `backend.ParseMetadata(path)` is called
- **Then** no error; other valid lines parsed correctly

#### UT-04: category parser — standard file
- **Given** the actual `data/categories` file
- **When** `backend.LoadCategories(hamappsDir)` is called
- **Then** returns 9 categories; first ID is `digital-modes`

#### UT-05: category parser — comment lines skipped
- **Given** a temp categories file with a `#`-prefixed comment line
- **When** `backend.LoadCategories(path)` is called
- **Then** comment line not included in results

#### UT-06: IsInstalled — file exists
- **Given** a temp status dir with `<slug>` file present
- **When** `backend.IsInstalled(statusDir, slug)` is called
- **Then** returns true

#### UT-07: IsInstalled — file absent
- **Given** a temp status dir without `<slug>` file
- **When** `backend.IsInstalled(statusDir, slug)` is called
- **Then** returns false

#### UT-08: LoadApps — discovers all apps
- **Given** temp `apps/` directory with 3 subdirectories each having `metadata` and `description`
- **When** `backend.LoadApps(hamappsDir)` is called
- **Then** returns slice of 3 AppInfo; slugs match directory names

#### UT-09: LoadApps — missing description file
- **Given** an app directory with `metadata` but no `description` file
- **When** `backend.LoadApps(hamappsDir)` is called
- **Then** app included with empty description; no error

#### UT-10: LoadApps — missing metadata file (partial-failure semantics)
- **Given** one app directory with no `metadata` file, and one valid app directory
- **When** `backend.LoadApps(hamappsDir)` is called
- **Then** returns one `AppInfo` (the valid app) and one error entry in the errors slice; no panic; no empty-list terminal failure

#### UT-11: app list filtering — by category
- **Given** 3 AppInfo items, two with category `digital-modes`, one with `packet-aprs`
- **When** `FilterApps(apps, "digital-modes", "")` is called
- **Then** returns 2 items

#### UT-12: app list filtering — by search text (name match)
- **Given** 3 AppInfo items
- **When** `FilterApps(apps, "", "wsjtx")` is called (case-insensitive)
- **Then** returns only WSJT-X

#### UT-13: app list filtering — by search text (description match)
- **Given** AppInfo with description containing "weak signal"
- **When** `FilterApps(apps, "", "weak signal")` is called
- **Then** returns that app

#### UT-14: app list filtering — combined category + search
- **Given** 4 apps: 2 in `digital-modes`, 2 in `packet-aprs`
- **When** `FilterApps(apps, "digital-modes", "fldigi")` is called
- **Then** returns 1 item (Fldigi only)

#### UT-15: version reading — file present
- **Given** temp `version` file containing `0.3.0\n`
- **When** `backend.ReadVersion(hamappsDir)` is called
- **Then** returns `"0.3.0"` (trimmed)

#### UT-16: version reading — file absent
- **Given** no `version` file
- **When** `backend.ReadVersion(hamappsDir)` is called
- **Then** returns `"dev"` fallback

#### UT-17: install state write — mark installed
- **Given** temp status dir
- **When** `backend.MarkInstalled(statusDir, "wsjtx")` is called
- **Then** file `<statusDir>/wsjtx` exists on disk

#### UT-18: install state write — mark uninstalled
- **Given** temp status dir with `wsjtx` file
- **When** `backend.MarkUninstalled(statusDir, "wsjtx")` is called
- **Then** file `<statusDir>/wsjtx` no longer exists; no error if already absent

#### UT-19: runner — process cancelled before completion (two-phase kill)
- **Given** a script that sleeps 30 seconds
- **When** `runner.Start()` is called then `cancel()` immediately
- **Then** `done` channel receives result with non-zero exit code within 2 seconds; temp log file exists

#### UT-19b: runner — SIGTERM-ignoring process killed by SIGKILL within 5 seconds
- **Given** a script that traps SIGTERM and ignores it but exits on SIGKILL
- **When** `runner.Start()` is called then `cancel()` immediately
- **Then** `done` channel receives result within 5 seconds (SIGKILL fires at 3 s)

#### UT-20: runner — process exits 0
- **Given** a script that `echo hello && exit 0`
- **When** `runner.Start()` is called and awaited
- **Then** `done` channel receives `RunResult{ExitCode: 0}`

#### UT-21: runner — process exits non-zero
- **Given** a script that `exit 42`
- **When** `runner.Start()` is called and awaited
- **Then** `done` channel receives `RunResult{ExitCode: 42}`

#### UT-22: HAMAPPS_DIR — defaults to binary directory
- **Given** `HAMAPPS_DIR` env var not set
- **When** `app.ResolveHamappsDir(binaryPath)` is called with `/usr/local/bin/ham-apps-gui`
- **Then** returns `/usr/local/bin`

#### UT-23: HAMAPPS_DIR — env var overrides binary directory
- **Given** `HAMAPPS_DIR=/opt/ham-apps` env var set
- **When** `app.ResolveHamappsDir(binaryPath)` is called
- **Then** returns `/opt/ham-apps`

#### UT-24: slug validation — valid slug passes
- **Given** slug `"wsjtx-2"` 
- **When** `backend.ValidateSlug("wsjtx-2")` is called
- **Then** returns nil (no error)

#### UT-25: slug validation — invalid slug rejected
- **Given** slug `"../../../etc/passwd"`
- **When** `backend.ValidateSlug("../../../etc/passwd")` is called
- **Then** returns non-nil error

#### UT-26: ValidateHamappsDir — valid directory passes
- **Given** a temp directory with `apps/`, `scripts/`, `data/categories` present
- **When** `backend.ValidateHamappsDir(dir)` is called
- **Then** returns nil

#### UT-27: ValidateHamappsDir — relative path rejected
- **Given** a relative path `"./ham-apps"`
- **When** `backend.ValidateHamappsDir("./ham-apps")` is called
- **Then** returns non-nil error mentioning "absolute path"

#### UT-28: ValidateHamappsDir — missing required subdirectory rejected
- **Given** a temp directory missing `data/categories`
- **When** `backend.ValidateHamappsDir(dir)` is called
- **Then** returns non-nil error mentioning the missing path

#### UT-29: AppRepository interface — FilesystemRepository implements it
- **Given** `backend.FilesystemRepository` struct
- **When** compile-time check `var _ port.AppRepository = (*backend.FilesystemRepository)(nil)` is present
- **Then** compiles without error

#### UT-30: RunnerService interface — BashRunner implements it
- **Given** `runner.BashRunner` struct
- **When** compile-time check `var _ port.RunnerService = (*runner.BashRunner)(nil)` is present
- **Then** compiles without error

### Bash Integration Tests (`tests/test-go-fyne-gui-*`)

#### IT-01: binary exists and is executable
- **Given** `gui/ham-apps-gui` has been built
- **When** `test -x gui/ham-apps-gui` is run
- **Then** exit code 0

#### IT-02: binary prints version and exits
- **Given** `gui/ham-apps-gui` binary present, `HAMAPPS_DIR` set to repo root
- **When** `HAMAPPS_DIR="$PWD" gui/ham-apps-gui --version` is run
- **Then** output matches `ham-apps [0-9]+\.[0-9]+\.[0-9]+`; exit code 0

#### IT-03: ham-apps entry point calls new binary
- **Given** `ham-apps` entry point updated; `gui/ham-apps-gui` built
- **When** `grep 'ham-apps-gui' ham-apps` is run
- **Then** match found

#### IT-04: go.mod exists in gui/
- **Given** repo checkout
- **When** `test -f gui/go.mod` is run
- **Then** exit code 0

#### IT-05: go.sum exists in gui/
- **Given** repo checkout
- **When** `test -f gui/go.sum` is run
- **Then** exit code 0

#### IT-06: go vet passes
- **Given** Go toolchain installed; `gui/` module
- **When** `cd gui && go vet ./...` is run
- **Then** exit code 0, zero output

#### IT-07: Go unit tests pass
- **Given** Go toolchain installed
- **When** `cd gui && go test ./...` is run
- **Then** exit code 0

#### IT-08: existing bash tests still pass
- **Given** `ham-apps` entry point updated
- **When** `for t in tests/test-*; do bash "$t"; done` is run
- **Then** all exit 0, all print only PASS: lines

#### IT-09: HAMAPPS_DIR smoke test — list command unaffected
- **Given** `HAMAPPS_DIR` set to repo root
- **When** `HAMAPPS_DIR="$PWD" bash ham-apps list` is run
- **Then** exit code 0; output contains app names

#### IT-10: binary not setuid and not world-writable
- **Given** `gui/ham-apps-gui` built
- **When** `stat -c '%a' gui/ham-apps-gui` is run
- **Then** permissions are `755` (not `4755` or `777`)

### Acceptance Tests (Manual)

#### AT-01: App list displays all apps on startup
- **Given** ham-apps installed with 5 apps
- **When** `ham-apps gui` is launched
- **Then** all 5 apps appear in the list within 2 seconds

#### AT-02: Search filters in real time
- **Given** main window open
- **When** user types "wsjtx" in search bar
- **Then** list narrows to WSJT-X within 200 ms

#### AT-03: Category dropdown filters list
- **Given** main window open
- **When** user selects "Digital Modes" from category dropdown
- **Then** only apps with category `digital-modes` are shown

#### AT-04: Install confirmation dialog
- **Given** WSJT-X selected, not installed
- **When** user clicks Install
- **Then** confirmation dialog shows name, category, website hyperlink, description, and icon

#### AT-05: Progress dialog during install
- **Given** user confirms install
- **When** install script starts
- **Then** modal progress dialog with pulsating bar and app name appears within 1 second

#### AT-06: Success dialog after install
- **Given** install script exits 0
- **When** progress dialog auto-closes
- **Then** success dialog shows "WSJT-X installed successfully."

#### AT-07: Status badge updates after install
- **Given** WSJT-X installed successfully
- **When** main window refreshed (or auto-refreshed)
- **Then** WSJT-X row shows green "Installed" badge

#### AT-08: Cancel install stops script
- **Given** install running (progress dialog visible)
- **When** user clicks Cancel
- **Then** process killed; "Installation cancelled" shown; install-state file NOT created

#### AT-09: Error dialog on failure
- **Given** install script exits non-zero
- **When** progress dialog closes
- **Then** error dialog shown with exit code; Show Log button reveals script output

#### AT-10: Sudo password prompt (GUI)
- **Given** sudo credentials not cached
- **When** user initiates install
- **Then** Fyne password dialog appears; entering correct password proceeds; wrong password shows error

#### AT-11: Uninstall flow mirrors install flow
- **Given** WSJT-X installed
- **When** user selects and clicks Uninstall → confirms
- **Then** progress, success/error, and state update follow same pattern as install

#### AT-12: Window is resizable
- **Given** main window open
- **When** user drags window edge
- **Then** window and list resize fluidly

#### AT-13: Refresh button updates state
- **Given** GUI open showing Direwolf as "Not Installed"
- **When** user manually creates `~/.local/share/ham-apps/installed/direwolf` then clicks Refresh
- **Then** Direwolf row updates to "Installed" without closing window

#### AT-14: Website link opens browser
- **Given** app detail dialog for WSJT-X
- **When** user clicks the website URL
- **Then** default browser opens `https://wsjt.sourceforge.io`

---

## Security & Compliance

### Threat Model

| Threat | Vector | Impact | Severity |
|--------|--------|--------|----------|
| Path traversal via app slug | Malicious `apps/` directory entry with `../` in dirname | Arbitrary file read/exec | Critical |
| Command injection via slug | Slug passed to `exec.Command("bash", script, slug)` | Code execution as user | Critical |
| Privilege escalation via SUDO_ASKPASS | Attacker replaces askpass temp script before `sudo -A -v` | Root execution of attacker code | High |
| Symlink attack on temp log file | Install script creates symlink at log path before write | Arbitrary file overwrite | High |
| Icon path traversal | `icon.png` path constructed from slug without validation | Arbitrary file read via image loader | Medium |
| Markup/rendering injection | App name with Fyne special characters corrupting layout | Visual glitch / crash | Low |

### Security Controls

#### SC-01: App slug validation
All app slugs read from the filesystem directory listing. Before use in any path construction or `exec.Command` argument:
```go
var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
func ValidateSlug(slug string) error {
    if !slugPattern.MatchString(slug) {
        return fmt.Errorf("invalid app slug: %q", slug)
    }
    return nil
}
```
Applied in `backend.LoadApps()` when scanning directory entries — invalid slugs are skipped with a warning log.

#### SC-02: No shell string interpolation
All subprocess invocations use `exec.Command` with argument arrays, never `exec.Command("bash", "-c", "... "+slug)`:
```go
// CORRECT
cmd := exec.Command("bash", scriptPath, slug)
// NEVER
cmd := exec.Command("bash", "-c", "bash "+scriptPath+" "+slug)
```

#### SC-03: Temp file security
- Log file: `os.CreateTemp("", "hamapps-*.log")` — created with mode 0600 by Go's os package
- Askpass script: written to `os.CreateTemp("", "hamapps-askpass-*.sh")` — set executable only for owner (`chmod 0700`)
- All temp files cleaned up via `defer os.Remove(path)` or `defer os.RemoveAll(dir)`
- Process group kill on cancel: `syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)` then `SIGKILL` after 3 s

#### SC-04: Icon path confinement
Icon path is constructed as `filepath.Join(hamappsDir, "apps", slug, "icon.png")`. The slug has already passed SC-01 validation. `filepath.Clean` is applied to the resulting path. The result is verified to be under `filepath.Join(hamappsDir, "apps")` using `strings.HasPrefix`:
```go
iconPath := filepath.Clean(filepath.Join(hamappsDir, "apps", slug, "icon.png"))
if !strings.HasPrefix(iconPath, filepath.Join(hamappsDir, "apps")) {
    return nil, fmt.Errorf("icon path escapes app directory")
}
```

#### SC-05: No setuid; no root execution
The `ham-apps-gui` binary runs as the current user. `sudo` is invoked only as a subprocess argument to the bash install/uninstall scripts. The binary itself never calls `setuid`, `setgid`, or opens privileged file descriptors.

#### SC-06: Askpass temp script secure cleanup
The askpass temp script is created with `os.CreateTemp` (atomic, mode 0700), the password written immediately, its path exported as `SUDO_ASKPASS`, `sudo -A -v` run, and then the file is securely deleted:
```go
defer func() {
    // Overwrite content before removing to clear password from filesystem blocks
    if f, err := os.OpenFile(askpassPath, os.O_WRONLY, 0); err == nil {
        f.Write(bytes.Repeat([]byte(" "), 256))
        f.Close()
    }
    os.Remove(askpassPath)
}()
```
A signal handler (`signal.Notify` for SIGTERM, SIGINT, SIGHUP) also runs this cleanup to handle process crashes.

#### SC-07: Log file cleanup
The temp log file (`os.CreateTemp("", "hamapps-*.log")`, mode 0600) is deleted via `defer os.Remove(logPath)` when the result dialog is closed. The log is never persisted beyond the dialog lifetime. The Show Log / Copy Log UI displays a notice: "Output may contain sensitive information." Install scripts MUST NOT echo credentials or passwords to stdout (convention documented in CONTRIBUTING.md).

### Compliance Requirements
- All Go code passes `go vet ./...` with zero issues
- All Go code passes `golangci-lint run` with zero issues (or project-agreed baseline)
- Binary permissions `755` (not setuid)
- No credentials, tokens, or secrets embedded in source or binary

---

## Implementation Plan

### Phases

**Phase 1 — Go Module Setup & Backend (no GUI)**
- Task 1.1 [setup]: Initialize Go module (`gui/go.mod` with module path `github.com/kw4jlb/ham-apps/gui`), add Fyne dependency, write `gui/Makefile` with `build`, `test`, `check`, `lint`, `clean` targets
- Task 1.2 [impl]: Define `gui/internal/port` package: `AppRepository` and `RunnerService` interfaces, `AppInfo`, `Category`, `RunResult` types
- Task 1.3 [test]: Write Go unit tests for backend package (metadata parser, category parser, IsInstalled, LoadApps partial-failure semantics, FilterApps, version reader, slug validator, ValidateHamappsDir, MarkInstalled, MarkUninstalled)
- Task 1.4 [impl]: Implement `gui/internal/backend` package implementing `port.AppRepository`: ParseMetadata, LoadCategories, LoadApps (returns `[]AppInfo, []error`), IsInstalled, MarkInstalled, MarkUninstalled, ReadVersion, ValidateSlug, ValidateHamappsDir, FilterApps, LoadIcon

**Phase 2 — Runner Package**
- Task 2.1 [test]: Write Go unit tests for runner package (Start, cancel two-phase kill, exit-0, exit-nonzero, log cleanup, askpass secure deletion)
- Task 2.2 [impl]: Implement `gui/internal/runner` implementing `port.RunnerService`:
  - `runner_linux.go` (`//go:build linux`): BashRunner with Setpgid, two-phase SIGTERM→SIGKILL cancel
  - `runner_other.go` (`//go:build !linux`): stub using `cmd.Process.Kill()`
  - `sudo.go`: SudoChecker, AskPassRunner with secure temp script creation and signal-handler cleanup

**Phase 3 — UI Components**
- Task 3.1 [test]: Write bash smoke tests (`tests/test-go-fyne-gui-smoke`): binary exists and executable, --version output format, go.mod/go.sum present, go vet passes, binary permissions 755
- Task 3.2 [impl]: Implement `gui/internal/ui/applist.go`: AppListWindow with real-time search, category filter dropdown, scrollable list with empty-state widget, Install/Uninstall buttons (disabled when no selection), Refresh button; keyboard navigation per spec
- Task 3.3 [impl]: Implement `gui/internal/ui/detail.go`: AppDetailDialog with icon (placeholder fallback), metadata display, website Hyperlink widget, confirm/cancel
- Task 3.4 [impl]: Implement `gui/internal/ui/progress.go`: ProgressDialog with pulsating bar, Show Log (inline), Cancel with inline confirmation flow
- Task 3.5 [impl]: Implement `gui/internal/ui/result.go`: ResultDialog for success (simple close) and error cases (title, body, inline log, Show Log, Copy Log, Close)

**Phase 4 — Application Bootstrap & Entry Point**
- Task 4.1 [impl]: Implement `gui/internal/app/bootstrap.go`: ResolveHamappsDir (env var or binary dir), ValidateHamappsDir call on startup, wires `backend.FilesystemRepository` and `runner.BashRunner` as `port.AppRepository`/`port.RunnerService`; passes to UI layer
- Task 4.2 [impl]: Implement `gui/cmd/ham-apps-gui/main.go`: arg parsing (--version, --help), calls bootstrap, launches Fyne app
- Task 4.3 [impl]: Update `ham-apps` bash entry point: replace `gui/app-list` call with `gui/ham-apps-gui`

**Phase 5 — Integration, CI & Distribution**
- Task 5.1 [test]: Write `tests/test-go-fyne-gui-integration`: binary permissions 755, HAMAPPS_DIR env var smoke test, ham-apps entry point contains `ham-apps-gui`, existing bash tests pass, go vet passes, --version output valid
- Task 5.2 [impl]: Add `.github/workflows/go-gui.yml` CI workflow: checkout, setup-go, apt deps, go vet, go test, govulncheck, build, --version check, bash tests
- Task 5.3 [impl]: Update `install.sh` to download pre-built binary from GitHub Releases for detected architecture (amd64/arm64); fallback to `make -C gui build` if Go toolchain present
- Task 5.4 [impl]: Build binary (`make -C gui build`); verify size ≤ 25 MB, permissions 755, `--version` output

### Task Overview

| Task | Files Created/Modified | Dependencies | Estimate |
|------|----------------------|--------------|----------|
| 1.1  | `gui/go.mod`, `gui/go.sum`, `gui/Makefile` | None | 1h |
| 1.2  | `gui/internal/port/*.go` | 1.1 | 1h |
| 1.3  | `gui/internal/backend/*_test.go` | 1.2 | 3h |
| 1.4  | `gui/internal/backend/*.go` | 1.3 | 4h |
| 2.1  | `gui/internal/runner/*_test.go` | 1.2 | 2h |
| 2.2  | `gui/internal/runner/*.go` | 2.1, 1.4 | 3h |
| 3.1  | `tests/test-go-fyne-gui-smoke` | 1.1 | 1h |
| 3.2  | `gui/internal/ui/applist.go` | 1.4, 2.2 | 4h |
| 3.3  | `gui/internal/ui/detail.go` | 1.2 | 3h |
| 3.4  | `gui/internal/ui/progress.go` | 2.2 | 2h |
| 3.5  | `gui/internal/ui/result.go` | 2.2 | 2h |
| 4.1  | `gui/internal/app/bootstrap.go` | 3.2–3.5, 1.4, 2.2 | 2h |
| 4.2  | `gui/cmd/ham-apps-gui/main.go` | 4.1 | 1h |
| 4.3  | `ham-apps` (entry point) | 4.2 | 0.5h |
| 5.1  | `tests/test-go-fyne-gui-integration` | 4.2, 4.3 | 2h |
| 5.2  | `.github/workflows/go-gui.yml` | 5.1 | 1h |
| 5.3  | `install.sh` (binary download logic) | 5.2 | 2h |
| 5.4  | `gui/ham-apps-gui` (binary artifact) | 5.3 | 0.5h |

**Total Estimated Effort**: ~35 hours

### Configuration

No new runtime configuration files. Build-time only:
- `gui/go.mod` — Go module descriptor
- `gui/go.sum` — dependency checksums
- `gui/Makefile` — build targets

Environment variables (runtime):
- `HAMAPPS_DIR` — existing variable; binary reads it on startup
- `DISPLAY` / `WAYLAND_DISPLAY` — standard X11/Wayland; Fyne handles automatically

---

## Deployment

### Build Prerequisites (Developer / CI)
```bash
# Install Go 1.21+
sudo apt install golang-go

# Install C toolchain (required for CGO/Fyne)
sudo apt install gcc libgl1-mesa-dev libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev

# Build the binary
make -C gui build
```

### End-User Deployment (via install.sh)
`install.sh` is updated to:
1. Detect host architecture: `uname -m` → `amd64` or `arm64`
2. Fetch the latest release tag from GitHub API
3. Download the matching pre-built binary (`ham-apps-gui-linux-amd64` or `ham-apps-gui-linux-arm64`) from GitHub Releases
4. Place at `gui/ham-apps-gui` with permissions `755`
5. If download fails and Go toolchain is present: fall back to `make -C gui build`
6. If neither works: print a clear error explaining the requirement and exit 1

### Developer Deployment Steps
1. Build: `make -C gui build` → produces `gui/ham-apps-gui`
2. Verify: `HAMAPPS_DIR="$PWD" gui/ham-apps-gui --version`
3. Run full check: `make -C gui check` (go vet + go test + bash tests)
4. Launch GUI: `HAMAPPS_DIR="$PWD" bash ham-apps gui`
5. Verify all acceptance criteria (AT-01 through AT-14) manually

### Rollback Plan
- Keep `gui/app-list` and `gui/app-details` in place during the transition period
- Rollback: `git checkout ham-apps` restores the entry point to call `gui/app-list`
- No data migration required; install-state directory is unchanged

### Monitoring
No server-side monitoring. User-visible feedback:
- Success/error dialogs after each action
- Log file accessible via Show Log button
- `--version` flag for build verification

---

## Acceptance Criteria

- [ ] FR-01: Main window displays all apps in a scrollable list with icon, name, category, status badge, and description
- [ ] FR-02: Search bar filters app list in real time across name, category, and description
- [ ] FR-03: Category dropdown populated from `data/categories`; "All" shows every app
- [ ] FR-04: Install confirmation dialog shows full metadata; Install button runs `scripts/install-app`
- [ ] FR-05: Uninstall confirmation dialog runs `scripts/uninstall-app`
- [ ] FR-06: Progress dialog with pulsating bar appears within 1 second of confirming action
- [ ] FR-07: Cancel kills the background bash process; install-state file not written
- [ ] FR-08: Status badges read from `~/.local/share/ham-apps/installed/<slug>`; green/grey visual distinction
- [ ] FR-09: Refresh button re-reads filesystem state without closing window
- [ ] FR-10: App detail shows icon (or placeholder), clickable website link, full description
- [ ] FR-11: Sudo GUI prompt appears when credentials not cached; correct password proceeds; wrong password shows error
- [ ] FR-12: Success dialog shown after exit-0 action
- [ ] FR-13: Error dialog shown after non-zero exit; title "Installation Failed — [App]"; Show Log inline; Copy Log button; log file deleted on dialog close
- [ ] FR-14: `HAMAPPS_DIR` env var respected; defaults to binary directory; `ValidateHamappsDir` called on startup
- [ ] FR-15: Binary at `gui/ham-apps-gui`; `ham-apps` entry point updated
- [ ] NFR-01: Binary ≤ 25 MB
- [ ] NFR-02: Window visible ≤ 2 s on target hardware
- [ ] NFR-04: No new system package requirements beyond libGL + libX11
- [ ] NFR-09: `go vet ./...` passes with zero issues
- [ ] NFR-10: Backend package unit test coverage ≥ 80%
- [ ] NFR-12: All existing bash tests pass after entry-point update
- [ ] SC-01: All app slugs validated against `^[a-zA-Z0-9][a-zA-Z0-9_-]*$` before use
- [ ] SC-02: All subprocess calls use `exec.Command` argument arrays; no `-c "string"` interpolation
- [ ] SC-03: Temp files created with `os.CreateTemp`; log file deleted after dialog closed; askpass script overwritten then removed; signal handler ensures cleanup on crash
- [ ] SC-04: Icon path verified to remain within `apps/` subdirectory
- [ ] SC-05: Binary not setuid; runs as current user only
- [ ] SC-07: Show Log / Copy Log UI displays "Output may contain sensitive information" notice
- [ ] UX-01: App list shows empty-state message (not blank) when search or category filter yields no results
- [ ] UX-02: Keyboard navigation: Tab order from search → category → list → buttons; arrow keys in list; Enter opens detail dialog
- [ ] UX-03: Install/Uninstall buttons disabled when no app selected
- [ ] UX-04: Cancel during install/uninstall shows inline confirmation before killing process
- [ ] CI-01: `.github/workflows/go-gui.yml` present; runs go vet, go test, govulncheck, build, bash tests
- [ ] CI-02: `install.sh` updated to download pre-built binary from GitHub Releases

---

## References

- `gui/app-list` — existing main browser (to be replaced)
- `gui/app-details` — existing detail/confirm dialog (to be replaced)
- `ham-apps` — bash entry point (to be updated)
- `scripts/utils` — shared bash helpers (unchanged; functions replicated in Go backend)
- `scripts/install-app`, `scripts/uninstall-app` — bash scripts invoked by runner
- `data/categories` — category definitions
- `apps/*/metadata`, `apps/*/description`, `apps/*/icon.png` — app data layout
- `version` — version string file
- Fyne documentation: https://developer.fyne.io
- Fyne API: `fyne.io/fyne/v2`
- Go `os/exec` package: https://pkg.go.dev/os/exec
- `specs/ui-ux-improvements/spec.md` — prior yad UX spec (reference for UX patterns)
