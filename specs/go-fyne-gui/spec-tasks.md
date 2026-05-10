# Implementation Tasks: Go + Fyne GUI Front-End for ham-apps

**Spec**: specs/go-fyne-gui/spec.md
**Status**: In Progress
**Total Estimated Effort**: 35h
**Critical Path**: 1.1 → 1.2 → 1.3 → 1.4 → 2.1 → 2.2 → 3.2 → 3.3 → 3.4 → 3.5 → 4.1 → 4.2 → 4.3 → 5.1 → 5.2 → 5.3 → 5.4

## Summary
- Total Tasks: 18
- Phases: 5
- Key Dependencies:
  - All backend and runner tasks require the port interfaces (1.2)
  - UI components require both backend (1.4) and runner (2.2) implementations
  - Bootstrap (4.1) requires all UI components (3.2–3.5)
  - Entry point update (4.3) requires binary available (4.2)
  - CI/distribution (5.2, 5.3) require integration tests to pass (5.1)

---

## Tasks

### Phase 1: Go Module Setup, Interfaces & Backend

#### Task 1.1 — Go Module Initialization
- **Type**: setup
- **Estimate**: 1h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Initialize the Go module at `gui/` with the canonical module path `github.com/kw4jlb/ham-apps/gui`. Add the Fyne v2.4+ dependency. Write `gui/Makefile` with `build`, `test`, `check`, `lint`, and `clean` targets. The `make test` target must run `go vet ./... && go test ./...`. The `make check` target must additionally run the bash tests in `../tests/test-*`.

**Acceptance Criteria**:
- [ ] `gui/go.mod` exists with `module github.com/kw4jlb/ham-apps/gui` and `go 1.21`
- [ ] `gui/go.sum` exists with pinned Fyne dependency checksums
- [ ] `make -C gui build` compiles (even if empty main.go placeholder) without error
- [ ] `make -C gui test` runs `go vet ./... && go test ./...`
- [ ] `make -C gui check` additionally runs `for t in ../tests/test-*; do bash "$t"; done`
- [ ] `make -C gui clean` removes `gui/ham-apps-gui`

**Implementation Notes**:
- Module path: `github.com/kw4jlb/ham-apps/gui`
- Fyne import path: `fyne.io/fyne/v2`
- CGO_ENABLED=1 required for build; document in Makefile comment
- Use `-ldflags="-s -w"` in build target to strip debug symbols
- Files: `gui/go.mod`, `gui/go.sum`, `gui/Makefile`

---

#### Task 1.2 — Port Interface Package
- **Type**: impl
- **Estimate**: 1h
- **Priority**: Critical
- **Dependencies**: Task 1.1
- **Status**: Complete

**Description**: Define the `internal/port` package containing the `AppRepository` and `RunnerService` interfaces, plus the shared data types `AppInfo`, `Category`, and `RunResult`. This package has zero dependencies on Fyne or filesystem packages — it is the pure abstraction layer that decouples UI from infrastructure.

**Acceptance Criteria**:
- [ ] `gui/internal/port/repository.go` defines `AppRepository` interface with all methods: `LoadApps() ([]AppInfo, []error)`, `LoadCategories() ([]Category, error)`, `IsInstalled(slug string) bool`, `MarkInstalled(slug string) error`, `MarkUninstalled(slug string) error`, `ReadVersion() string`, `LoadIcon(slug string) []byte`
- [ ] `gui/internal/port/runner.go` defines `RunnerService` interface: `Start(script, slug string) (cancel func(), done <-chan RunResult)`, `CheckSudo() bool`, `PromptSudo(appName string) error`
- [ ] `gui/internal/port/types.go` defines `AppInfo`, `Category`, `RunResult` structs matching spec data models
- [ ] Package compiles with `go build ./internal/port/...`
- [ ] Package has no imports of `fyne.io`, `os`, `os/exec` (pure types and interfaces only)

**Implementation Notes**:
- `AppInfo` fields: Slug, Name, Category, Website, Tags ([]string), MinOS, Description, Summary, Installed (bool), IconPath (string)
- `Category` fields: ID, DisplayName, Description
- `RunResult` fields: ExitCode (int), LogFile (string), Err (error)
- File layout: `gui/internal/port/repository.go`, `gui/internal/port/runner.go`, `gui/internal/port/types.go`

---

#### Task 1.3 — Backend Unit Tests (TDD Red Phase)
- **Type**: test
- **Estimate**: 3h
- **Priority**: Critical
- **Dependencies**: Task 1.2
- **Status**: Complete

**Description**: Write all Go unit tests for the backend package BEFORE implementing the backend. Tests must fail (red phase) at this point. Tests cover: metadata parser, category parser, IsInstalled, LoadApps (including partial-failure semantics), FilterApps, version reader, slug validator, ValidateHamappsDir, MarkInstalled, MarkUninstalled, and interface compliance checks.

**Acceptance Criteria**:
- [ ] `gui/internal/backend/backend_test.go` exists
- [ ] Tests for UT-01 through UT-30 are written (see spec Test Specification)
- [ ] Tests compile but fail (red) since backend package not yet implemented
- [ ] Each test uses `t.TempDir()` for isolation; no global state
- [ ] UT-29 compile-time interface check: `var _ port.AppRepository = (*FilesystemRepository)(nil)`
- [ ] UT-30 compile-time interface check: `var _ port.RunnerService = (*runner.BashRunner)(nil)` (in runner test file)
- [ ] `go test ./internal/backend/... 2>&1 | grep -c FAIL` returns nonzero

**Implementation Notes**:
- Test file: `gui/internal/backend/backend_test.go`
- Use `os.TempDir()` / `t.TempDir()` for fixture directories
- Helper `makeMetadataFile(t, dir, slug string, fields map[string]string)` to reduce boilerplate
- UT-04 uses the actual `data/categories` file — compute path from test binary's working directory using `os.Getenv("HAMAPPS_DIR")` or a relative path helper
- For UT-19b (SIGTERM-ignoring): write a temp bash script that does `trap '' TERM; sleep 60`

---

#### Task 1.4 — Backend Package Implementation
- **Type**: impl
- **Estimate**: 4h
- **Priority**: Critical
- **Dependencies**: Task 1.3
- **Status**: Complete

**Description**: Implement the `gui/internal/backend` package providing `FilesystemRepository` which implements `port.AppRepository`. All functions are pure Go with no Fyne dependency. The implementation must make all UT-01 through UT-28 tests pass.

**Acceptance Criteria**:
- [ ] `FilesystemRepository` implements `port.AppRepository` (compile-time check from UT-29 passes)
- [ ] `LoadApps` returns `([]AppInfo, []error)` — valid apps in first return, per-app errors in second; never panics
- [ ] `ValidateSlug` rejects slugs not matching `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- [ ] `ValidateHamappsDir` requires absolute path and presence of `apps/`, `scripts/`, `data/categories`
- [ ] `FilterApps(apps []AppInfo, category, search string) []AppInfo` — case-insensitive search across Name, Category, Description; category filter exact-match; "All" shows everything
- [ ] `LoadIcon` returns nil (not error) if `icon.png` absent; validates path stays within `apps/` dir
- [ ] `MarkInstalled` creates `~/.local/share/ham-apps/installed/<slug>` with `os.MkdirAll` then empty file
- [ ] `MarkUninstalled` removes state file; no error if already absent (`os.Remove` ignores `ErrNotExist`)
- [ ] All UT-01 through UT-28 tests pass: `cd gui && go test ./internal/backend/...`

**Implementation Notes**:
- Files: `gui/internal/backend/backend.go`, `gui/internal/backend/filter.go`
- `FilesystemRepository` struct holds `hamappsDir` and `statusDir` strings
- Metadata parser: `bufio.Scanner` over file lines; `strings.SplitN(line, "=", 2)`
- `LoadIcon` path confinement: `filepath.Clean` + `strings.HasPrefix` check
- `ReadVersion`: `os.ReadFile(path.Join(hamappsDir, "version"))`; `strings.TrimSpace`; fallback `"dev"`

---

### Phase 2: Runner Package

#### Task 2.1 — Runner Unit Tests (TDD Red Phase)
- **Type**: test
- **Estimate**: 2h
- **Priority**: Critical
- **Dependencies**: Task 1.2
- **Status**: Complete

**Description**: Write all Go unit tests for the runner package BEFORE implementing the runner. Tests cover: Start and await, cancel (two-phase kill), SIGTERM-ignoring process killed by SIGKILL, exit-0, exit-nonzero, log file creation and mode, askpass secure deletion.

**Acceptance Criteria**:
- [ ] `gui/internal/runner/runner_test.go` exists with tests for UT-19, UT-19b, UT-20, UT-21
- [ ] Test for log file mode: after `Start()`, log file at `RunResult.LogFile` has mode 0600
- [ ] Test for log cleanup: after result dialog closes (simulated), log file deleted
- [ ] Tests compile but fail (red) since runner not yet implemented
- [ ] UT-30 compile-time check present: `var _ port.RunnerService = (*BashRunner)(nil)`

**Implementation Notes**:
- Test file: `gui/internal/runner/runner_test.go`
- UT-19b: write a temp script `trap '' TERM; sleep 60` to test SIGKILL escalation
- Tests that use signals are Linux-only: add `//go:build linux` to those test cases or use `t.Skip` on non-Linux
- Askpass cleanup test: create a fake askpass file; call cleanup func; verify file absent and content zeroed

---

#### Task 2.2 — Runner Package Implementation
- **Type**: impl
- **Estimate**: 3h
- **Priority**: Critical
- **Dependencies**: Task 2.1, Task 1.4
- **Status**: Complete

**Description**: Implement `gui/internal/runner` providing `BashRunner` which implements `port.RunnerService`. Include platform-specific process-group kill (Linux) and a cross-platform stub (non-Linux). Implement SudoChecker and AskPassRunner with secure temp script cleanup.

**Acceptance Criteria**:
- [ ] `BashRunner` implements `port.RunnerService` (compile-time check from UT-30 passes)
- [ ] `Start(script, slug string)` returns `(cancel func(), done <-chan RunResult)`
- [ ] On Linux: `SysProcAttr.Setpgid = true`; cancel sends SIGTERM then SIGKILL after 3 s via `time.AfterFunc`
- [ ] On non-Linux: cancel calls `cmd.Process.Kill()`
- [ ] Log file created with `os.CreateTemp("", "hamapps-*.log")` mode 0600
- [ ] Log file deleted via deferred cleanup when `RunResult` consumer calls cleanup (pass cleanup func in RunResult or via separate callback)
- [ ] `CheckSudo()` runs `exec.Command("sudo", "-n", "true")`; returns true if exit 0
- [ ] `AskPassRunner` creates temp script with mode 0700; overwrites content before `os.Remove`; registers `signal.Notify` handler for SIGTERM, SIGINT, SIGHUP
- [ ] All runner tests pass: `cd gui && go test ./internal/runner/...`

**Implementation Notes**:
- Files: `gui/internal/runner/runner.go`, `gui/internal/runner/runner_linux.go`, `gui/internal/runner/runner_other.go`, `gui/internal/runner/sudo.go`
- `runner_linux.go` build tag: `//go:build linux`
- `runner_other.go` build tag: `//go:build !linux`
- Use `os/signal` package for cleanup signal handler; cancel channel to stop handler goroutine
- `exec.Command("bash", scriptPath, slug)` — argument array, not string interpolation

---

### Phase 3: UI Components

#### Task 3.1 — Bash Smoke Tests
- **Type**: test
- **Estimate**: 1h
- **Priority**: High
- **Dependencies**: Task 1.1
- **Status**: Complete

**Description**: Write bash smoke tests at `tests/test-go-fyne-gui-smoke` following the existing test convention. These tests verify the presence of module files and, once the binary is built, verify its basic properties. Tests that require the binary to exist are gated on a `test -x` check.

**Acceptance Criteria**:
- [ ] `tests/test-go-fyne-gui-smoke` exists and is executable
- [ ] Test file follows `PASS:`/`FAIL:` convention; exits non-zero on any failure
- [ ] Tests check: `gui/go.mod` exists, module path contains `ham-apps/gui`
- [ ] Tests check: `gui/go.sum` exists
- [ ] Tests check: `gui/Makefile` exists and contains `build:` target
- [ ] If binary exists: `test -x gui/ham-apps-gui` passes
- [ ] If binary exists: `stat -c '%a' gui/ham-apps-gui` returns `755`
- [ ] If binary exists and `HAMAPPS_DIR` set: `--version` output matches `^ham-apps [0-9]+\.[0-9]+\.[0-9]+`
- [ ] `bash tests/test-go-fyne-gui-smoke` exits 0 on a fresh checkout (before binary is built, binary checks are skipped)

**Implementation Notes**:
- File: `tests/test-go-fyne-gui-smoke`
- Pattern: `if test -x gui/ham-apps-gui; then ... fi` for binary-dependent checks
- Follow `tests/test-utils-helpers` pattern for PASS:/FAIL: output format

---

#### Task 3.2 — AppListWindow UI Component
- **Type**: impl
- **Estimate**: 4h
- **Priority**: Critical
- **Dependencies**: Task 1.4, Task 2.2
- **Status**: Complete

**Description**: Implement `gui/internal/ui/applist.go` containing the `AppListWindow` Fyne widget. The window is the primary user-facing component: it displays the app list, search bar, category filter, status badges, and action buttons. It accepts `port.AppRepository` and `port.RunnerService` interfaces (not concrete types).

**Acceptance Criteria**:
- [ ] `AppListWindow` struct accepts `port.AppRepository` and `port.RunnerService`; no direct imports of `internal/backend` or `internal/runner`
- [ ] Search bar: `widget.Entry` bound to a string; filter applied on each `OnChanged` call
- [ ] Category dropdown: `widget.Select` populated from `repo.LoadCategories()`; "All" is the first and default option
- [ ] App list: `widget.List` (or `container.NewScroll` with custom rows); each row shows icon (48×48 canvas.Image), bold app name (`widget.RichText`), category label, status badge (green/grey `canvas.Rectangle` + `widget.Label`)
- [ ] Empty-state: when `FilterApps` returns 0 results, list area shows a `widget.Label` with appropriate message and (for search) a [Clear Search] button; when results present, list shows normally
- [ ] Install/Uninstall buttons: `widget.Button`; disabled (`SetEnabled(false)`) when no list item selected; enabled on single-click selection
- [ ] Refresh button: re-calls `repo.LoadApps()` and redraws list; does not close window
- [ ] Single click: selects row and enables action buttons
- [ ] Double-click or Enter on focused row: creates and shows `AppDetailDialog`
- [ ] Tab order follows spec: search → category → list → Install → Uninstall → Refresh
- [ ] Minimum window size: 960×540; `window.SetMasterWindow()` and `window.Resize()` called
- [ ] Window title: "ham-apps — Amateur Radio App Manager"

**Implementation Notes**:
- File: `gui/internal/ui/applist.go`
- Use `fyne.io/fyne/v2/widget`, `fyne.io/fyne/v2/canvas`, `fyne.io/fyne/v2/container`
- Status badge: Fyne `canvas.NewRectangle` with `color.RGBA` for green/grey; overlaid with `widget.Label`
- Icon loading: call `repo.LoadIcon(slug)` → `[]byte`; decode with `fyne.NewStaticResource` → `canvas.NewImageFromResource`; fallback to `theme.FyneLogo()` or a built-in placeholder resource
- Filtering: `backend.FilterApps` is a pure function; import it or pass a filter func to decouple

---

#### Task 3.3 — AppDetailDialog UI Component
- **Type**: impl
- **Estimate**: 3h
- **Priority**: High
- **Dependencies**: Task 1.2
- **Status**: Complete

**Description**: Implement `gui/internal/ui/detail.go` containing the `AppDetailDialog`. This modal dialog shows full app metadata before install or uninstall. It accepts the action (`"install"` or `"uninstall"`) and calls the provided callback on confirmation.

**Acceptance Criteria**:
- [ ] `AppDetailDialog` function or struct accepts `app port.AppInfo`, `action string`, `onConfirm func()` callback
- [ ] Dialog title: "Install [App Name]" or "Uninstall [App Name]"
- [ ] Icon displayed at 96×96; placeholder shown when `app.IconPath` is empty or `LoadIcon` returns nil
- [ ] Name, Category, Website displayed; Website rendered as `widget.Hyperlink` opening default browser
- [ ] Full description displayed in a scrollable `widget.RichText` or `widget.Label` with wrapping
- [ ] Action button labeled "Install" or "Uninstall"; Cancel button labeled "Cancel"
- [ ] Cancel dismisses dialog without calling `onConfirm`
- [ ] Confirm calls `onConfirm()` and dismisses dialog

**Implementation Notes**:
- File: `gui/internal/ui/detail.go`
- Use `dialog.NewCustom` or `dialog.NewCustomConfirm` from `fyne.io/fyne/v2/dialog`
- Website hyperlink: `widget.NewHyperlink(text, url)` where `url` is parsed with `url.Parse`
- For long descriptions, wrap in `container.NewScroll(widget.NewLabel(desc))`

---

#### Task 3.4 — ProgressDialog UI Component
- **Type**: impl
- **Estimate**: 2h
- **Priority**: Critical
- **Dependencies**: Task 2.2
- **Status**: Complete

**Description**: Implement `gui/internal/ui/progress.go` containing the `ProgressDialog`. This modal is shown while a bash install/uninstall script is running in the background. It presents a pulsating progress bar, inline log viewer (Show Log), and a Cancel flow with inline confirmation.

**Acceptance Criteria**:
- [ ] `ProgressDialog` accepts app name, action string, log file path, and cancel func from runner
- [ ] Displays `widget.ProgressBarInfinite` (pulsating)
- [ ] Shows label: "Installing [App Name], please wait..."
- [ ] Show Log button: toggles visibility of an inline `widget.Entry` (read-only, scrollable) tailing the log file; updates every 500 ms via `time.Ticker`
- [ ] Cancel button: shows inline confirmation widget ("Cancel installation? Any files already downloaded may remain on disk.") with [Keep Installing] and [Confirm Cancel] buttons
- [ ] [Confirm Cancel] calls the cancel func (two-phase kill happens in runner), then closes dialog and triggers result flow
- [ ] [Keep Installing] dismisses confirmation widget and returns to normal progress view
- [ ] Dialog closes automatically when `done` channel receives a `RunResult`
- [ ] Dialog is modal (blocks parent window interaction)

**Implementation Notes**:
- File: `gui/internal/ui/progress.go`
- Use `dialog.NewCustom` for the modal
- Log tailing: read log file with `os.ReadFile` in a goroutine; update `widget.Entry.SetText` on each tick via `fyne.CurrentApp().SendNotification` or direct Fyne thread-safe call
- Inline confirmation: use `widget.NewLabel` + two buttons inside a `container.NewHBox`; toggle visibility with `widget.Show/Hide`

---

#### Task 3.5 — ResultDialog UI Component
- **Type**: impl
- **Estimate**: 2h
- **Priority**: Critical
- **Dependencies**: Task 2.2
- **Status**: Complete

**Description**: Implement `gui/internal/ui/result.go` containing the `ResultDialog`. Shows success or failure outcome after a runner `RunResult` is received. The error variant includes inline log viewer, Copy Log, and a sensitive-data notice. Log file is deleted when dialog is closed.

**Acceptance Criteria**:
- [ ] `ShowSuccess(appName, action string)` displays: title "[App Name] Installed/Uninstalled", body "[App Name] was installed/uninstalled successfully.", [Close] button
- [ ] `ShowError(appName, action string, exitCode int, logFile string, cleanup func())` displays:
  - Title: "Installation Failed — [App Name]" or "Uninstall Failed — [App Name]"
  - Body: "[App Name] could not be installed/uninstalled. Exit code: [N]. Check the log for details."
  - Notice: "Output may contain sensitive information."
  - Buttons: [Show Log] [Copy Log] [Close]
- [ ] [Show Log] expands inline scrollable `widget.Entry` (read-only) with log contents
- [ ] [Copy Log] copies log file contents to system clipboard via `fyne.CurrentApp().Clipboard().SetContent(...)`
- [ ] [Close] calls `cleanup()` (which deletes log file) then dismisses dialog
- [ ] Log file also deleted on dialog window close (X button) via `dialog.SetOnClosed`

**Implementation Notes**:
- File: `gui/internal/ui/result.go`
- Use `fyne.io/fyne/v2/dialog` and `fyne.io/fyne/v2/widget`
- `cleanup func()` passed from the runner layer; calls `os.Remove(logFile)`
- Cancelled result (not error, not success): title "Cancelled", body "Installation was cancelled. You can try again later. Some temporary files may remain." with [Show Log] [Close]

---

### Phase 4: Application Bootstrap & Entry Point

#### Task 4.1 — Application Bootstrap
- **Type**: impl
- **Estimate**: 2h
- **Priority**: Critical
- **Dependencies**: Task 3.2, Task 3.3, Task 3.4, Task 3.5, Task 1.4, Task 2.2
- **Status**: Complete

**Description**: Implement `gui/internal/app/bootstrap.go` which is the dependency-injection root. It resolves `HAMAPPS_DIR`, validates it, creates concrete implementations of `port.AppRepository` and `port.RunnerService`, and wires them into the UI layer.

**Acceptance Criteria**:
- [ ] `ResolveHamappsDir(binaryPath string) string` checks `HAMAPPS_DIR` env var first; falls back to `filepath.Dir(binaryPath)`
- [ ] `Bootstrap(binaryPath string) error` calls `ResolveHamappsDir`, then `backend.ValidateHamappsDir`; returns error if validation fails
- [ ] On success, creates `backend.FilesystemRepository` and `runner.BashRunner` as concrete types
- [ ] Passes them as `port.AppRepository` and `port.RunnerService` to `ui.NewAppListWindow`
- [ ] No direct import of `internal/ui` from `internal/backend` (one-way dependency)
- [ ] UT-22 and UT-23 pass (ResolveHamappsDir behavior)

**Implementation Notes**:
- File: `gui/internal/app/bootstrap.go`
- The `statusDir` is hardcoded as `filepath.Join(os.UserHomeDir(), ".local/share/ham-apps/installed")` within `FilesystemRepository`
- Pass `hamappsDir` to both `FilesystemRepository{hamappsDir: dir}` and `BashRunner{hamappsDir: dir}`

---

#### Task 4.2 — Main Entrypoint
- **Type**: impl
- **Estimate**: 1h
- **Priority**: Critical
- **Dependencies**: Task 4.1
- **Status**: Complete

**Description**: Implement `gui/cmd/ham-apps-gui/main.go`. Parse `--version` and `--help` flags (exit immediately). For GUI mode, call `app.Bootstrap`, create the Fyne application, and run the main window event loop.

**Acceptance Criteria**:
- [ ] `--version` flag: prints `ham-apps <version>` (reading version from `ReadVersion`) and exits 0
- [ ] `--help` flag: prints usage string and exits 0
- [ ] Unknown args: prints usage and exits 1
- [ ] No args: calls `app.Bootstrap` → creates `fyne.NewApp()` → shows main window → calls `fyne.CurrentApp().Run()`
- [ ] If `app.Bootstrap` returns error: prints error to stderr and exits 1 (no window opened)
- [ ] Binary compiles: `cd gui && go build -o ham-apps-gui ./cmd/ham-apps-gui`

**Implementation Notes**:
- File: `gui/cmd/ham-apps-gui/main.go`
- Use `os.Args` for simple flag parsing or `flag` package
- `fyne.NewApp()` sets app ID: `"io.github.kw4jlb.ham-apps"`
- Window must be created before `Run()` is called; `Bootstrap` returns the window or the `main` func creates it using the wired components

---

#### Task 4.3 — Update ham-apps Entry Point
- **Type**: impl
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: Task 4.2
- **Status**: Complete

**Description**: Update the `ham-apps` bash entry point to call `gui/ham-apps-gui` instead of `gui/app-list`. This is the only change to the bash entry point.

**Acceptance Criteria**:
- [ ] `ham-apps` file: `gui|""` case now calls `"$HAMAPPS_DIR/gui/ham-apps-gui"`
- [ ] `grep 'ham-apps-gui' ham-apps` returns a match
- [ ] `grep 'gui/app-list' ham-apps` returns no match (old call removed)
- [ ] `shellcheck -x ham-apps` passes with zero errors
- [ ] `HAMAPPS_DIR="$PWD" bash ham-apps --version` still works
- [ ] `HAMAPPS_DIR="$PWD" bash ham-apps list` still works

**Implementation Notes**:
- File: `ham-apps`
- Change only the `gui|""` case; leave all other cases unchanged
- `gui/app-list` and `gui/app-details` files are kept in place during transition (not deleted in this task)

---

### Phase 5: Integration, CI & Distribution

#### Task 5.1 — Integration Test Suite
- **Type**: test
- **Estimate**: 2h
- **Priority**: High
- **Dependencies**: Task 4.2, Task 4.3
- **Status**: Complete

**Description**: Write `tests/test-go-fyne-gui-integration` covering the integration between the binary, the entry point, the bash test suite, and the HAMAPPS_DIR contract. These tests verify the seam between the Go binary and the bash world.

**Acceptance Criteria**:
- [ ] `tests/test-go-fyne-gui-integration` exists and is executable
- [ ] Follows `PASS:`/`FAIL:` convention; exits non-zero on any failure
- [ ] Checks: `ham-apps` file contains `ham-apps-gui` string (entry point updated)
- [ ] Checks: `ham-apps` file does NOT contain call to `gui/app-list` (old call removed)
- [ ] Checks: `gui/go.mod` contains `module github.com/kw4jlb/ham-apps/gui`
- [ ] Checks: `gui/Makefile` contains `go vet` in test target
- [ ] If binary exists and Go toolchain present: `cd gui && go vet ./...` exits 0
- [ ] If binary exists: binary permissions are `755` (not `4755`)
- [ ] If binary exists: `HAMAPPS_DIR="$PWD" gui/ham-apps-gui --version` exits 0 and output matches version pattern
- [ ] Existing bash tests: `for t in tests/test-{utils-helpers,hamrs-metadata,hamrs-scripts,trustedqsl-metadata,trustedqsl-scripts,install-sh}; do bash "$t"; done` all pass

**Implementation Notes**:
- File: `tests/test-go-fyne-gui-integration`
- Gate binary-dependent checks: `if command -v go &>/dev/null && test -x gui/ham-apps-gui; then ... fi`
- Reuse HAMAPPS_DIR detection: `HAMAPPS_DIR="${HAMAPPS_DIR:-$(dirname "$(readlink -f "$0")/..")}"`

---

#### Task 5.2 — GitHub Actions CI Workflow
- **Type**: impl
- **Estimate**: 1h
- **Priority**: High
- **Dependencies**: Task 5.1
- **Status**: Complete

**Description**: Add `.github/workflows/go-gui.yml` implementing the CI pipeline specified in the spec. The workflow runs on push to main and on pull requests targeting main.

**Acceptance Criteria**:
- [ ] `.github/workflows/go-gui.yml` exists
- [ ] Workflow triggers on `push` to `main` and `pull_request` targeting `main`
- [ ] Steps in order: checkout, setup-go (1.21), apt install build deps, go vet, go test, govulncheck, build binary, verify --version, run bash tests
- [ ] `govulncheck` installed via `go install golang.org/x/vuln/cmd/govulncheck@latest` and run as `govulncheck ./...`
- [ ] All `go` commands run from `gui/` directory (`cd gui && ...` or `working-directory: gui`)
- [ ] Build step: `CGO_ENABLED=1 go build -o ham-apps-gui ./cmd/ham-apps-gui`
- [ ] Verify step: `HAMAPPS_DIR="$GITHUB_WORKSPACE" gui/ham-apps-gui --version`
- [ ] Bash tests step: `for t in tests/test-*; do bash "$t"; done`
- [ ] Workflow YAML is syntactically valid

**Implementation Notes**:
- File: `.github/workflows/go-gui.yml`
- Ubuntu runner has mesa/GL headers available: `libgl1-mesa-dev`
- X11 headers: `libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev`
- `actions/checkout@v4`, `actions/setup-go@v5` are current stable versions

---

#### Task 5.3 — install.sh Binary Distribution Support
- **Type**: impl
- **Estimate**: 2h
- **Priority**: High
- **Dependencies**: Task 5.2
- **Status**: Complete

**Description**: Update `install.sh` to handle downloading the pre-built `ham-apps-gui` binary from GitHub Releases. The update preserves the existing installation flow; the GUI binary download is added as an additional step.

**Acceptance Criteria**:
- [ ] `install.sh` updated to detect architecture: `uname -m` → `x86_64` → `amd64`, `aarch64` → `arm64`
- [ ] Fetches latest release tag from GitHub API: `https://api.github.com/repos/KW4JLB/ham-apps/releases/latest`
- [ ] Downloads binary: `https://github.com/KW4JLB/ham-apps/releases/download/<tag>/ham-apps-gui-linux-<arch>`
- [ ] Places binary at `$HAMAPPS_DIR/gui/ham-apps-gui` with mode `755`
- [ ] If download fails (curl error, 404): attempts `make -C "$HAMAPPS_DIR/gui" build` if `go` is in PATH
- [ ] If both fail: prints clear error "ham-apps GUI binary could not be installed. See docs/getting-started/installation.md" and exits 1
- [ ] `shellcheck -x install.sh` passes with zero errors
- [ ] Existing test `tests/test-install-sh` continues to pass

**Implementation Notes**:
- File: `install.sh`
- Use `curl -fsSL` for downloads with `--output` flag
- Architecture mapping: `case "$(uname -m)" in x86_64) arch=amd64;; aarch64|arm64) arch=arm64;; *) arch="";; esac`
- Conditional build fallback: `if command -v go &>/dev/null; then make -C "$HAMAPPS_DIR/gui" build; fi`
- Keep existing install.sh logic intact; only add the GUI binary download block

---

#### Task 5.4 — Build and Verify Binary
- **Type**: impl
- **Estimate**: 0.5h
- **Priority**: High
- **Dependencies**: Task 5.3
- **Status**: Complete

**Description**: Build the final binary, verify its properties (size, permissions, version output), and confirm all integration tests pass.

**Acceptance Criteria**:
- [ ] `make -C gui build` exits 0
- [ ] `gui/ham-apps-gui` exists and is executable (`test -x gui/ham-apps-gui`)
- [ ] Binary size: `du -b gui/ham-apps-gui | awk '{print $1}'` ≤ 26214400 (25 MB)
- [ ] Permissions: `stat -c '%a' gui/ham-apps-gui` = `755`
- [ ] Version output: `HAMAPPS_DIR="$PWD" gui/ham-apps-gui --version` prints `ham-apps 0.3.0` (or current version)
- [ ] Integration tests pass: `bash tests/test-go-fyne-gui-integration`
- [ ] Smoke tests pass: `bash tests/test-go-fyne-gui-smoke`
- [ ] All existing bash tests pass: `for t in tests/test-*; do bash "$t"; done`

**Implementation Notes**:
- This task is the final verification gate
- If binary exceeds 25 MB, apply additional `-ldflags="-s -w"` and consider `upx --brute` compression (note in findings if used)
- The binary is NOT committed to git; it is built from source or downloaded via install.sh
