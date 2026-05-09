# ham-apps GUI UX Improvements

## Overview

### Purpose
Improve the yad-based GUI for ham-apps to feel polished, consistent, and intuitive for amateur radio operators. Address 10 identified UX pain points without changing the core architecture (pure bash + yad), and without introducing any new runtime dependencies.

### Scope
- `gui/app-list` — main browser window
- `gui/app-details` — install/uninstall confirmation dialog
- `scripts/utils` — shared utilities (minor additions only)

### Background
ham-apps provides a GUI and CLI for installing amateur radio software on Debian 11+ and Ubuntu 20.04+. The GUI uses yad (≥ 0.40) with GDK_BACKEND=x11 for XWayland compatibility. The 10 UX pain points were identified through a structured review of the existing two GUI scripts.

---

## Requirements

### Functional Requirements

| ID    | Requirement | Priority |
|-------|-------------|----------|
| FR-01 | Progress feedback: display a yad --progress pulse bar during install/uninstall, labeled with the app name and action | High |
| FR-02 | Category filter button: relabel as "Fil_ter: $CATEGORY" (Alt+T mnemonic) with tooltip "Click to filter by category". If the selected category yields zero apps, display an info dialog and reset to "All". | High |
| FR-03 | Row visual distinction: installed rows display a bold app name via yad markup; not-installed rows display normal weight | High |
| FR-04 | Resizable window: replace fixed --width=900 --height=500 with --width=960 --height=540 and add --maximizable and --resizable flags | Medium |
| FR-05 | Keyboard shortcuts: add GTK mnemonic labels — "_Install" (Alt+I), "_Uninstall" (Alt+U), "Fil_ter: $CATEGORY" (Alt+T), "Re_fresh" (Alt+E) — using underscore prefix in yad button labels | Medium |
| FR-06 | app-details icon: pass --image with the app's icon.png (fallback: dialog-information stock icon) | Medium |
| FR-07 | app-details website hyperlink: render the website as a clickable `<a href="...">` using yad --text with Pango markup | Medium |
| FR-08 | Refresh button: add a "Refresh" button (exit code 8) that re-reads install state and redraws the list without changing the category filter or search text | High |
| FR-09 | Error dialogs: replace bare yad --info with yad --error --image=dialog-error --title="Error" for all error cases in both GUI scripts | High |
| FR-10 | Preserve search on category change: store the current search text in a variable; after category picker, pass it back via --search-entry-text yad flag if available, else note limitation | Medium |
| FR-11 | Success notification: after install/uninstall terminal completes (exit 0), display yad --info with --image=dialog-information titled "Success" confirming the action | High |
| FR-12 | Progress: replace terminal-based install/uninstall with a background process + yad --progress pulsed dialog. Include Cancel (kills background process, shows "Cancelled" message) and Hide (closes dialog, install continues, completion notification shown on finish) buttons. | High |

### Non-Functional Requirements

| ID     | Requirement | Target |
|--------|-------------|--------|
| NFR-01 | Startup performance: app-list window must appear within 2 seconds for ≤ 50 apps | < 2 s |
| NFR-02 | All GUI scripts must pass shellcheck with zero errors | 0 errors |
| NFR-03 | All GUI scripts must use set -euo pipefail | 100% |
| NFR-04 | All GUI scripts must source scripts/utils | 100% |
| NFR-05 | No new runtime dependencies (bash + yad + x-terminal-emulator only) | 0 new deps |
| NFR-06 | GDK_BACKEND=x11 exported in all GUI scripts | 100% |
| NFR-07 | Icon fallback must work when apps/<slug>/icon.png is absent | 100% |

### Constraints

- Pure bash + yad (no Python, no compiled code, no web UI)
- yad version ≥ 0.40 on Debian 11+ and Ubuntu 20.04+
- `sudo` is used inside app install scripts only, never by gui scripts
- No hardcoded version strings without a `# TODO: dynamic version detection` comment
- All spec files reside in `specs/ui-ux-improvements/`
- No INDEX.md, SUMMARY.md, or versioned files

---

## Design

### Architecture

The GUI remains a two-script architecture:

```
ham-apps (entry point)
  └─ gui/app-list          ← main loop (modified)
        └─ gui/app-details ← detail/confirm dialog (modified)
              └─ scripts/install-app | uninstall-app  (unchanged)
```

#### FR-01 / FR-12: Progress Feedback Pattern

Instead of raw `x-terminal-emulator -e bash -c "..."`, `gui/app-details` will:

1. Create a secure temporary directory: `TMPDIR_WORK="$(mktemp -d /tmp/hamapps-XXXXXX)"`. Create a named FIFO inside it: `FIFO="${TMPDIR_WORK}/progress.fifo"; mkfifo "$FIFO"`. Create a log file: `LOG_FILE="${TMPDIR_WORK}/output.log"`. Create an exit-code file path: `EXIT_FILE="${TMPDIR_WORK}/exit.code"`.
2. Set up trap for full cleanup: `trap 'kill "$BG_PID" 2>/dev/null; rm -rf "$TMPDIR_WORK"' EXIT`.
3. Launch the install/uninstall script in the background: `bash "$HAMAPPS_DIR/scripts/$ACTION_SCRIPT" "$APP" >"$LOG_FILE" 2>&1; echo $? >"$EXIT_FILE"` & `BG_PID=$!` (store the background PID immediately).
4. Display a `yad --progress --pulsate --title="<Action> $name" --text="Running $ACTION for <b>$name</b>..." --button="Cancel:1" --button="Hide:0"` dialog, reading from the FIFO. If the user clicks Cancel (exit code 1): kill `$BG_PID`, display "Cancelled" info dialog, do NOT call `mark_installed`/`mark_uninstalled`, exit.
5. When the background process finishes, write EOF to the FIFO to auto-close the progress bar.
6. Read the exit code from `$EXIT_FILE` and display a success or error dialog accordingly.
7. Call `mark_installed` / `mark_uninstalled` only if exit code is 0 and user did not cancel.

**Cancellation contract**: Cancel button kills BG_PID and shows: "Installation cancelled. $name was not fully installed." Hiding the dialog (Hide button) closes the visual progress bar but continues the install in the background; a completion notification dialog appears when done.

The terminal window approach is preserved as a fallback when `yad` `--progress` is unavailable (yad < 0.40), but since we target ≥ 0.40 the primary path uses `--progress`.

Because amateur radio users often want to see the raw output (e.g., apt output), a "Show Details" `yad --text-info` tail-follow window is spawned alongside the progress bar via the log file. This is optional/non-blocking.

#### FR-08: Refresh Without State Loss

The main loop variable `CATEGORY` persists across iterations. A new variable `SEARCH_TEXT` is introduced. On exit code 8 (Refresh), the loop simply `continue`s without changing `CATEGORY` or `SEARCH_TEXT`. The next iteration of `build_app_data` re-reads `is_installed` state.

#### FR-10: Search Persistence

yad ≥ 0.40 does not expose `--search-entry-text` as a readable value from outside the dialog. The search bar content is not retrievable from the yad exit. Therefore:

- On category change (exit code 6), the search text is reset (accepted limitation, noted in UI as tooltip: "Search resets when changing category").
- This is documented as a known constraint of yad's GTK list widget.

#### FR-03: Row Visual Distinction via Markup

yad `--list` supports Pango markup in column values when `--enable-markup` is passed. Installed app names will be wrapped in `<b>...</b>`. Not-installed names will be plain text.

#### FR-05: Keyboard Shortcuts

yad button labels support GTK mnemonic notation with underscore: `"_Install"` creates Alt+I, `"_Uninstall"` creates Alt+U. The category filter button becomes `"Fil_ter: $CATEGORY"` for Alt+T (avoiding the Alt+F File-menu conflict). The refresh button uses `"Re_fresh"` for Alt+E.

#### FR-06 / FR-07: app-details Icon and Hyperlink

- `--image`: resolved path `$APPS_DIR/$APP/icon.png` if the file exists; otherwise `dialog-information` (GTK stock).
- Website as link: yad `--question` `--text` supports Pango markup; the website field is rendered as `<a href="$website">$website</a>`. GTK label links require yad to be built with `--enable-links` (standard in Debian/Ubuntu packages).

### Data Flow

```
build_app_data()
  ├─ list_all_apps()          → app slug list
  ├─ get_metadata(app, name)  → display name
  ├─ get_metadata(app, cat)   → category
  ├─ head -1 description      → short description
  ├─ is_installed(app)        → bool
  └─ emit 7 yad columns per app (icon, name[markup], category,
                                  status, description, id, search-target)
```

### Script Structure Changes

**gui/app-list** changes:
- Add `SEARCH_TEXT=""` variable
- Add `--enable-markup` to yad `--list` call
- Change `--button` labels to use GTK mnemonics: `"_Install:2"`, `"_Uninstall:4"`, `"Fil_ter\: $CATEGORY:6"`, `"Re_fresh:8"`, `"_Close:0"`
- Add `--resizable --maximizable` flags
- Change window size to `--width=960 --height=540`
- Handle exit code 8 (refresh: continue loop without changing CATEGORY)
- In `build_app_data`: call `escape_markup "$name"` before wrapping in `<b>...</b>` (installed) or emitting plain (not installed); also escape description
- Empty-state check: after `build_app_data "$CATEGORY"` produces output, if result is empty display `gui_info "No Apps Found" "No apps in category <b>$(escape_markup "$CATEGORY")</b>. Resetting to All."` and set `CATEGORY="All"`, then `continue`
- Use `gui_error` for all error conditions (no-selection, etc.) instead of `yad --info`

**gui/app-details** changes:
- Resolve icon path; set `--image` flag
- Render website as Pango `<a href>` link
- Replace `x-terminal-emulator` launch with background process + FIFO + progress bar
- Add success/failure dialog after action completes
- Use `yad --error` for error messages

**scripts/utils** changes:
- Move `escape_markup()` from `gui/app-details` into `scripts/utils` so both GUI scripts can use it
- Add `gui_error()` helper: `yad --error --title="Error" --text="$1" --image=dialog-error 2>/dev/null`
- Add `gui_info()` helper: `yad --info --title="$1" --text="$2" --image=dialog-information 2>/dev/null`

---

## Test Specification

### Unit Tests (shellcheck + bash unit)

#### UT-01: shellcheck passes on gui/app-list
- **Given** the modified `gui/app-list` script
- **When** `shellcheck gui/app-list` is run
- **Then** exit code 0, zero warnings or errors

#### UT-02: shellcheck passes on gui/app-details
- **Given** the modified `gui/app-details` script
- **When** `shellcheck gui/app-details` is run
- **Then** exit code 0, zero warnings or errors

#### UT-03: set -euo pipefail present in app-list
- **Given** `gui/app-list`
- **When** `grep -c 'set -euo pipefail' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-04: set -euo pipefail present in app-details
- **Given** `gui/app-details`
- **When** `grep -c 'set -euo pipefail' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-05: gui_error function exists in utils
- **Given** `scripts/utils`
- **When** `grep -c 'gui_error()' scripts/utils` is run
- **Then** output is ≥ 1

#### UT-06: gui_info function exists in utils
- **Given** `scripts/utils`
- **When** `grep -c 'gui_info()' scripts/utils` is run
- **Then** output is ≥ 1

#### UT-07: app-list sources utils
- **Given** `gui/app-list`
- **When** `grep -c 'source.*utils' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-08: app-details sources utils
- **Given** `gui/app-details`
- **When** `grep -c 'source.*utils' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-09: GDK_BACKEND=x11 set in app-list
- **Given** `gui/app-list`
- **When** `grep -c 'GDK_BACKEND=x11' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-10: Refresh button (exit code 8) present in app-list
- **Given** `gui/app-list`
- **When** `grep -c 'Re_fresh:8\|Refresh:8' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-11: enable-markup flag present in app-list yad call
- **Given** `gui/app-list`
- **When** `grep -c 'enable-markup' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-12: app-details uses --image flag
- **Given** `gui/app-details`
- **When** `grep -c '\-\-image' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-13: app-details renders website as anchor link
- **Given** `gui/app-details`
- **When** `grep -c 'href' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-14: app-details uses yad --progress or equivalent feedback
- **Given** `gui/app-details`
- **When** `grep -c '\-\-progress\|\-\-pulse' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-15: app-list uses --resizable flag
- **Given** `gui/app-list`
- **When** `grep -c '\-\-resizable' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-16: mnemonic shortcuts present in app-list buttons
- **Given** `gui/app-list`
- **When** `grep -c '_Install\|_Uninstall\|Fil_ter\|Re_fresh' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-21: escape_markup function exists in scripts/utils
- **Given** `scripts/utils`
- **When** `grep -c 'escape_markup()' scripts/utils` is run
- **Then** output is ≥ 1

#### UT-22: escape_markup NOT defined locally in gui/app-details (moved to utils)
- **Given** `gui/app-details`
- **When** `grep -c 'escape_markup()' gui/app-details` is run
- **Then** output is 0

#### UT-23: Progress tmpdir uses mktemp -d (not mktemp -u)
- **Given** `gui/app-details`
- **When** `grep -c 'mktemp -d' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-24: BG_PID variable is set in app-details
- **Given** `gui/app-details`
- **When** `grep -c 'BG_PID' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-25: Cancel button defined in progress dialog (app-details)
- **Given** `gui/app-details`
- **When** `grep -c 'Cancel\|cancel' gui/app-details` is run
- **Then** output is ≥ 1

#### UT-26: Empty-state check present in app-list main loop
- **Given** `gui/app-list`
- **When** `grep -c 'No Apps Found\|No apps\|-z.*app_data\|empty' gui/app-list` is run
- **Then** output is ≥ 1

#### UT-17: error dialogs use yad --error not yad --info for errors
- **Given** `gui/app-list`
- **When** `grep 'yad --info' gui/app-list` shows only non-error informational messages
- **Then** all error paths use `yad --error` or `gui_error`

#### UT-18: build_app_data wraps installed names in bold markup
- **Given** a stub environment where `is_installed` returns true for test-app
- **When** `build_app_data` is called
- **Then** the name column value contains `<b>` and `</b>` for test-app

#### UT-19: icon fallback to dialog-information when icon.png absent
- **Given** an app slug with no `icon.png` file
- **When** `gui/app-details` determines the image path
- **Then** the `--image` flag value is `dialog-information`

#### UT-20: success dialog shown after successful action in app-details
- **Given** `gui/app-details`
- **When** `grep -c 'Success\|success\|installed successfully\|removed successfully' gui/app-details`
- **Then** output is ≥ 1

### Integration Tests

#### IT-01: app-list displays without crashing on 3 test apps
- **Given** the ham-apps repo with direwolf, fldigi, wsjtx apps
- **When** `gui/app-list` is launched in dry-run (mocked yad returning exit 0 immediately)
- **Then** process exits 0 with no error output

#### IT-02: category filter reduces visible rows
- **Given** `build_app_data "digital-modes"` called with the 3 test apps
- **When** output lines are counted
- **Then** only apps with category=digital-modes appear (≥ 7 lines per app row × count)

#### IT-03: install state tracking round-trip
- **Given** direwolf is not installed (`mark_uninstalled direwolf`)
- **When** `build_app_data "All"` is called
- **Then** direwolf row status column = "Not installed", icon = gtk-no

#### IT-04: install state tracking after mark_installed
- **Given** `mark_installed direwolf` is called
- **When** `build_app_data "All"` is called
- **Then** direwolf row status column = "Installed", icon = gtk-yes, name contains `<b>`

#### IT-05: Refresh does not reset CATEGORY
- **Given** CATEGORY="digital-modes" in app-list loop
- **When** exit code 8 is received
- **Then** CATEGORY remains "digital-modes" in next iteration

#### IT-06: gui_error outputs valid yad command
- **Given** `source scripts/utils` and yad mocked to print its args
- **When** `gui_error "Test error message"` is called
- **Then** output contains `--error` and `Test error message`

### Acceptance Tests

#### AT-01: Progress bar appears during install (manual)
- **Given** a clean Ubuntu 20.04 VM with yad installed
- **When** the user selects WSJT-X and clicks Install
- **Then** a pulsating progress dialog titled "Installing WSJT-X" appears within 1 second

#### AT-02: Success dialog appears after install (manual)
- **Given** WSJT-X install script exits 0
- **When** install completes
- **Then** a dialog titled "Success" with message "WSJT-X installed successfully." appears

#### AT-03: Error dialog appears on failure (manual)
- **Given** an app whose install script exits 1
- **When** install fails
- **Then** a yad --error dialog titled "Error" with the failure message appears

#### AT-04: Category filter label shows current category (manual)
- **Given** CATEGORY="digital-modes" is selected
- **When** the main window is displayed
- **Then** the filter button label reads "Filter: digital-modes"

#### AT-05: Installed apps appear bold (manual)
- **Given** direwolf is marked installed
- **When** the app list is displayed
- **Then** the "Direwolf" name appears visually bold

#### AT-06: Window is resizable (manual)
- **Given** the main app-list window
- **When** the user drags the window corner
- **Then** the window resizes and columns reflow

#### AT-07: Alt+I triggers install (manual)
- **Given** an app is selected in the list
- **When** user presses Alt+I
- **Then** the install flow starts (app-details opens)

#### AT-08: Alt+U triggers uninstall (manual)
- **Given** an installed app is selected
- **When** user presses Alt+U
- **Then** the uninstall flow starts

#### AT-09: app-details shows app icon (manual)
- **Given** wsjtx has no icon.png
- **When** app-details opens for wsjtx
- **Then** the dialog shows the stock dialog-information icon

#### AT-10: Website is a clickable link (manual)
- **Given** app-details for wsjtx (website: https://wsjt.sourceforge.io)
- **When** the dialog is displayed
- **Then** the website appears as an underlined hyperlink

#### AT-11: Refresh button re-reads install state (manual)
- **Given** ham-apps GUI is open showing direwolf as "Not installed"
- **When** user manually creates `~/.local/share/ham-apps/installed/direwolf` and clicks Refresh
- **Then** direwolf row updates to "Installed" without closing the window

#### AT-12: Error dialog has error icon (manual)
- **Given** user clicks Install without selecting an app
- **When** the error appears
- **Then** a yad --error dialog with an error icon is shown (not plain --info)

---

## Security & Compliance

### Threat Model

| Threat | Vector | Impact |
|--------|--------|--------|
| Script injection via app slug | App slug passed to `x-terminal-emulator -e bash -c "... '$APP'"` | Code execution as user |
| Markup injection via app name/description | App name with `<b>` or `<script>` in name field | Visual corruption / crash |
| PATH manipulation | Attacker replaces `yad` binary | GUI hijack |
| Privilege escalation | GUI scripts calling `sudo` directly | Root escalation |

### Security Controls

#### SC-01: App slug sanitization
App slugs are filesystem directory names. Before passing to any bash command string, validate:
```bash
[[ "$APP" =~ ^[a-zA-Z0-9_-]+$ ]] || { gui_error "Invalid app identifier."; exit 1; }
```
This check must be added to `gui/app-details` before using `$APP` in any command substitution or string interpolation.

#### SC-02: Markup escaping
The `escape_markup()` function must be moved from `gui/app-details` to `scripts/utils` so it is available to both `gui/app-list` and `gui/app-details` via the shared `source "$HAMAPPS_DIR/scripts/utils"`. The function handles `&`, `<`, `>`. It must be called on ALL user-controlled fields (name, category, description) before inserting into Pango markup strings. The `build_app_data` function in `app-list` must call `escape_markup "$name"` before wrapping in `<b>...</b>` (FR-03). The `gui/app-details` script must remove its local definition of `escape_markup` since it will now come from `scripts/utils`.

#### SC-03: No sudo in GUI scripts
Neither `gui/app-list` nor `gui/app-details` should call `sudo`. The install/uninstall scripts handle privilege escalation internally.

#### SC-04: Temp file security
Progress feedback uses a secure temporary directory created with `mktemp -d /tmp/hamapps-XXXXXX` (atomically created, no TOCTOU race). The FIFO, log file, and exit-code file are all placed inside this directory. Full cleanup via: `trap 'kill "$BG_PID" 2>/dev/null; rm -rf "$TMPDIR_WORK"' EXIT`. The background PID is stored in `BG_PID` immediately after launching and is killed on both cancel and normal exit.

#### SC-05: yad path hardening
The GUI scripts use `command -v yad` to verify yad is present. No need to hardcode a path; yad is a system package.

#### SC-06: No shell injection in progress command
The background command for install/uninstall uses an array or heredoc form to avoid word-splitting on `$APP`:
```bash
bash "$HAMAPPS_DIR/scripts/$ACTION_SCRIPT" "$APP" >"$LOG_FILE" 2>&1
```
Not a string interpolation in `-c "..."`.

### Compliance Requirements
- All modified scripts must pass `shellcheck -S error` with zero findings
- Bash strict mode (`set -euo pipefail`) required in all GUI scripts

---

## Implementation Plan

### Phases

**Phase 1 — Foundation (scripts/utils)**
- Task T-01: Move `escape_markup` from `gui/app-details` to `scripts/utils`; add `gui_error` and `gui_info` helpers to `scripts/utils`

**Phase 2 — app-list improvements**
- Task T-02: Add `set -euo pipefail`, `--resizable`, `--maximizable`, window size update, `SEARCH_TEXT` variable
- Task T-03: Add `--enable-markup` and bold-name markup for installed apps in `build_app_data`
- Task T-04: Update button labels with mnemonics; add `Re_fresh:8` button; update category button label
- Task T-05: Handle exit code 8 (Refresh) in main loop; replace `yad --info` error dialogs with `gui_error`

**Phase 3 — app-details improvements**
- Task T-06: Add slug validation (SC-01); resolve icon path with fallback (FR-06)
- Task T-07: Render website as Pango anchor link (FR-07); pass `--image` to confirmation dialog
- Task T-08: Replace `x-terminal-emulator` launch with background process + yad --progress + success/failure dialog

**Phase 4 — Tests**
- Task T-09: Write shellcheck tests and grep-based unit tests for all acceptance criteria
- Task T-10: Write integration smoke tests for `build_app_data` category filtering and markup

### Task Overview

| Task | Files Modified | Dependencies |
|------|---------------|--------------|
| T-01 | scripts/utils | — |
| T-02 | gui/app-list | T-01 |
| T-03 | gui/app-list | T-02 |
| T-04 | gui/app-list | T-02 |
| T-05 | gui/app-list | T-03, T-04 |
| T-06 | gui/app-details | T-01 |
| T-07 | gui/app-details | T-06 |
| T-08 | gui/app-details | T-07 |
| T-09 | tests/test-ux-improvements | T-05, T-08 |
| T-10 | tests/test-ux-improvements | T-09 |

### Configuration
No new configuration files. No new environment variables beyond the existing `HAMAPPS_DIR` and `GDK_BACKEND`.

---

## Deployment

### Deployment Steps
1. Pull the branch
2. Verify yad ≥ 0.40: `yad --version`
3. Run `tests/test-ux-improvements` smoke test
4. Launch `gui/app-list` manually and verify all 10 UX improvements visually

### Rollback Plan
`git checkout gui/app-list gui/app-details scripts/utils` reverts all changes.

### Monitoring
No server-side monitoring. User-visible: error dialogs and success dialogs provide feedback.

---

## Acceptance Criteria

- [ ] FR-01/FR-12: Progress bar (pulsating) appears within 1 second of clicking Install/Uninstall
- [ ] FR-02: Category filter button label reads "Filter: All" or "Filter: <category>"
- [ ] FR-03: Installed app names appear bold in the list; not-installed names appear normal weight
- [ ] FR-04: Window is resizable by dragging; --maximizable flag present
- [ ] FR-05: Alt+I triggers Install, Alt+U triggers Uninstall keyboard shortcuts work
- [ ] FR-06: app-details dialog shows app's icon.png if present, else dialog-information
- [ ] FR-07: Website field in app-details is a clickable hyperlink (Pango anchor)
- [ ] FR-08: Refresh button reloads install state without resetting category filter
- [ ] FR-09: All error paths use yad --error with dialog-error image, not yad --info
- [ ] FR-10: Search reset on category change is documented as known yad limitation; empty-state dialog resets category to "All" automatically
- [ ] FR-12a: Cancel button on progress dialog kills background process; "Cancelled" message shown; mark_installed/mark_uninstalled NOT called on cancel
- [ ] FR-11: Success dialog shows after completed install/uninstall
- [ ] NFR-02: shellcheck passes with zero errors on both GUI scripts
- [ ] NFR-03: set -euo pipefail present in both GUI scripts
- [ ] SC-01: App slug validated against `^[a-zA-Z0-9_-]+$` before use in commands
- [ ] SC-02: escape_markup lives in scripts/utils; called in build_app_data before bold markup insertion and in app-details before all Pango markup strings; not defined locally in app-details
- [ ] SC-04: Temp files in mktemp -d directory; BG_PID tracked; trap EXIT kills BG_PID and removes tmpdir
- [ ] SC-06: install/uninstall launched with argument array, not string interpolation

---

## References

- `gui/app-list` — current implementation
- `gui/app-details` — current implementation
- `scripts/utils` — shared helpers
- `scripts/install-app` — install runner (unchanged)
- `scripts/uninstall-app` — uninstall runner (unchanged)
- `data/categories` — category definitions
- `apps/wsjtx/metadata`, `apps/fldigi/metadata`, `apps/direwolf/metadata` — sample metadata
- yad manual: `man yad` — `--progress`, `--enable-markup`, `--resizable`, `--maximizable`, `--error`
- GTK mnemonic labels: underscore prefix for Alt+key shortcuts
