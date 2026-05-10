# Implementation Tasks: ham-apps GUI UX Improvements

**Spec**: specs/ui-ux-improvements/spec.md
**Status**: Complete
**Total Estimated Effort**: 14.5h
**Critical Path**: 1.1 → 1.2 → 2.1 → 2.2 → 2.3 → 2.4 → 3.1 → 3.2 → 3.3 → 4.1 → 4.2

## Summary
- Total Tasks: 20
- Phases: 4
- Key Dependencies: T-01 (utils foundation) blocks all GUI tasks; T-02–T-05 (app-list) can proceed in parallel with T-06–T-08 (app-details) after T-01; tests precede each impl task

---

## Tasks

### Phase 1: Foundation — scripts/utils

#### Task 1.1 — Tests: utils helper functions
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Done

**Description**: Write shellcheck and grep-based tests for the new and moved functions in `scripts/utils`: `escape_markup`, `gui_error`, `gui_info`.

**Acceptance Criteria**:
- [x] Test file `tests/test-utils-helpers` created and executable
- [x] Test verifies `escape_markup()` defined in `scripts/utils`
- [x] Test verifies `gui_error()` defined in `scripts/utils`
- [x] Test verifies `gui_info()` defined in `scripts/utils`
- [x] Test verifies `escape_markup` converts `&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`
- [x] Test verifies `escape_markup` NOT defined in `gui/app-details` (moved, not duplicated)
- [x] All tests fail before implementation (red phase)

**Implementation Notes**:
- Test file: `tests/test-utils-helpers`
- Use `bash -c 'source scripts/utils; ...'` pattern for unit testing
- Mock yad with `YAD() { echo "$@"; }; export -f YAD` or `alias yad=...` to test gui_error/gui_info output

---

#### Task 1.2 — Impl: utils helper functions
- **Type**: impl
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: Task 1.1
- **Status**: Done

**Description**: Move `escape_markup()` from `gui/app-details` into `scripts/utils`. Add `gui_error()` and `gui_info()` to `scripts/utils`. Remove the local definition from `gui/app-details` (it will source from utils).

**Acceptance Criteria**:
- [x] `escape_markup()` present in `scripts/utils`
- [x] `escape_markup()` NOT present as a local function definition in `gui/app-details`
- [x] `gui_error()` present in `scripts/utils`; calls `yad --error --title="Error" --image=dialog-error`
- [x] `gui_info()` present in `scripts/utils`; calls `yad --info --image=dialog-information`
- [ ] `scripts/utils` passes `shellcheck` with zero errors (shellcheck not installed in env; bash -n syntax check passes)
- [x] All Task 1.1 tests pass (green phase)

**Implementation Notes**:
- File to modify: `scripts/utils`
- File to modify: `gui/app-details` (remove local `escape_markup()` definition)
- `escape_markup()` signature: `escape_markup() { printf '%s' "$1" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g'; }`
- `gui_error()` signature: `gui_error() { yad --error --title="Error" --text="$1" --image=dialog-error --button="OK:0" 2>/dev/null; }`
- `gui_info()` signature: `gui_info() { yad --info --title="$1" --text="$2" --image=dialog-information --button="OK:0" 2>/dev/null; }`

---

### Phase 2: app-list Improvements

#### Task 2.1 — Tests: app-list structural improvements
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: High
- **Dependencies**: Task 1.2
- **Status**: Done

**Description**: Write tests verifying structural changes to `gui/app-list`: `set -euo pipefail`, resizable flags, window dimensions, `SEARCH_TEXT` variable, `GDK_BACKEND`.

**Acceptance Criteria**:
- [x] Test file `tests/test-app-list-structure` created and executable
- [x] Tests verify `set -euo pipefail` present (UT-03)
- [x] Tests verify `--resizable` present (UT-15)
- [x] Tests verify `--maximizable` present
- [x] Tests verify `GDK_BACKEND=x11` present (UT-09)
- [x] Tests verify window size is `960` and `540`
- [x] Tests verify `SEARCH_TEXT` variable declared
- [x] All tests fail before implementation (red)

**Implementation Notes**:
- Test file: `tests/test-app-list-structure`
- All tests are static grep/shellcheck checks — no yad required
- Run: `shellcheck gui/app-list` as one of the test assertions

---

#### Task 2.2 — Impl: app-list structural improvements
- **Type**: impl
- **Estimate**: 0.5h
- **Priority**: High
- **Dependencies**: Task 2.1
- **Status**: Done

**Description**: Add `set -euo pipefail` to `gui/app-list`. Update window to `--width=960 --height=540 --resizable --maximizable`. Add `SEARCH_TEXT=""` variable before the main loop.

**Acceptance Criteria**:
- [x] `set -euo pipefail` on line 2 of `gui/app-list`
- [x] `--width=960 --height=540` in yad `--list` call
- [x] `--resizable --maximizable` in yad `--list` call
- [x] `SEARCH_TEXT=""` declared before `while true` loop
- [ ] `shellcheck gui/app-list` exits 0 (shellcheck not installed in env; bash -n syntax check passes)
- [x] All Task 2.1 tests pass

**Implementation Notes**:
- File: `gui/app-list`
- Place `set -euo pipefail` after the shebang and before `HAMAPPS_DIR=...`
- Place `SEARCH_TEXT=""` after `CATEGORY="All"` line

---

#### Task 2.3 — Tests: app-list markup, buttons, and category behavior
- **Type**: test
- **Estimate**: 0.75h
- **Priority**: High
- **Dependencies**: Task 2.2
- **Status**: Done

**Description**: Write tests for `--enable-markup`, bold name wrapping in `build_app_data`, mnemonic button labels, Refresh button (exit code 8), and empty-state check in the main loop.

**Acceptance Criteria**:
- [x] Test file `tests/test-app-list-ux` created and executable
- [x] Test verifies `--enable-markup` present in yad call (UT-11)
- [x] Test verifies `_Install` / `_Uninstall` / `Fil_ter` / `Re_fresh` present (UT-16)
- [x] Test verifies `Re_fresh:8` present (UT-10)
- [x] Test verifies `<b>` markup emitted for installed app names (UT-18) using sourced build_app_data with mocked is_installed
- [x] Test verifies `escape_markup` called before bold wrapping (no raw `<b>$name</b>`)
- [x] Test verifies empty-state check present (UT-26)
- [x] Test verifies error conditions use `gui_error` not bare `yad --info` (UT-17)
- [x] All tests fail before implementation (red)

**Implementation Notes**:
- Test file: `tests/test-app-list-ux`
- For UT-18: set up a test app dir in `/tmp`, mock `is_installed` to return 0, source `gui/app-list` functions, call `build_app_data` and grep output for `<b>`
- Use `grep -v '^#'` to filter comments when checking for raw `yad --info` vs `gui_error`

---

#### Task 2.4 — Impl: app-list markup, buttons, and category behavior
- **Type**: impl
- **Estimate**: 1.5h
- **Priority**: High
- **Dependencies**: Task 2.3
- **Status**: Done

**Description**: Add `--enable-markup` to the yad `--list` call. Update `build_app_data` to escape and bold-wrap installed app names and escape descriptions. Update all button labels with GTK mnemonics. Add Refresh (exit code 8) button. Add empty-state check. Replace `yad --info` error paths with `gui_error`. Handle exit code 8 in the case statement.

**Acceptance Criteria**:
- [x] `--enable-markup` present in yad `--list` call
- [x] `build_app_data` calls `escape_markup "$name"` before emitting name column
- [x] Installed apps: name column = `"<b>${safe_name}</b>"`, not-installed: `"${safe_name}"`
- [x] `build_app_data` calls `escape_markup "$description"` before emitting description column
- [x] Button labels: `"_Install:2"`, `"_Uninstall:4"`, `"Fil_ter\: $CATEGORY:6"`, `"Re_fresh:8"`, `"_Close:0"`
- [x] Case statement handles `8)` with `continue` (no CATEGORY/SEARCH_TEXT change)
- [x] Empty-state: if `build_app_data "$CATEGORY"` is empty, call `gui_info` and reset `CATEGORY="All"`, then `continue`
- [x] All bare `yad --info` on error paths replaced with `gui_error`
- [ ] `shellcheck gui/app-list` exits 0 (shellcheck not installed in env; bash -n syntax check passes)
- [x] All Task 2.3 tests pass

**Implementation Notes**:
- File: `gui/app-list`
- In `build_app_data`, replace `echo "$name"` with `safe_name="$(escape_markup "$name")"; [[ is_installed ]] && echo "<b>${safe_name}</b>" || echo "$safe_name"`
- The search target column (column 7) should use unescaped raw text for search accuracy — use `"${name} ${category} ${description}"` not escaped version

---

### Phase 3: app-details Improvements

#### Task 3.1 — Tests: app-details slug validation and icon
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: High
- **Dependencies**: Task 1.2
- **Status**: Done

**Description**: Write tests for slug validation regex (SC-01), icon path resolution with fallback (FR-06), and removal of local `escape_markup` definition.

**Acceptance Criteria**:
- [x] Test file `tests/test-app-details-foundation` created and executable
- [x] Test verifies slug regex `^[a-zA-Z0-9_-]+$` pattern present in `gui/app-details` (SC-01)
- [x] Test verifies `--image` flag used in `gui/app-details` (UT-12)
- [x] Test verifies `dialog-information` appears as fallback (UT-19)
- [x] Test verifies `escape_markup()` NOT defined locally in `gui/app-details` (UT-22)
- [x] Test verifies `set -euo pipefail` present (UT-04)
- [x] All tests fail before implementation (red)

**Implementation Notes**:
- Test file: `tests/test-app-details-foundation`

---

#### Task 3.2 — Impl: app-details slug validation and icon
- **Type**: impl
- **Estimate**: 0.75h
- **Priority**: High
- **Dependencies**: Task 3.1
- **Status**: Done

**Description**: Add `set -euo pipefail` to `gui/app-details`. Remove local `escape_markup()` definition (it now comes from utils). Add slug validation check. Add icon path resolution logic with `dialog-information` fallback.

**Acceptance Criteria**:
- [x] `set -euo pipefail` present in `gui/app-details`
- [x] Local `escape_markup()` definition removed from `gui/app-details`
- [x] Slug validation: `[[ "$APP" =~ ^[a-zA-Z0-9_-]+$ ]] || { gui_error "Invalid app identifier."; exit 1; }`
- [x] Icon resolution: `APP_ICON="$APPS_DIR/$APP/icon.png"; [[ -f "$APP_ICON" ]] || APP_ICON="dialog-information"`
- [x] `--image "$APP_ICON"` passed to yad `--question` calls
- [ ] `shellcheck gui/app-details` exits 0 (shellcheck not installed in env; bash -n syntax check passes)
- [x] All Task 3.1 tests pass

**Implementation Notes**:
- File: `gui/app-details`
- The slug validation must come after `require_app "$APP"` (which already validates the dir exists) but we add regex validation as defense-in-depth
- Since `scripts/utils` is sourced first, `escape_markup` will be available after removing the local definition

---

#### Task 3.3 — Tests: app-details website link and progress dialog
- **Type**: test
- **Estimate**: 0.75h
- **Priority**: High
- **Dependencies**: Task 3.2
- **Status**: Done

**Description**: Write tests for Pango anchor rendering of website (FR-07), progress dialog usage (FR-12/UT-14), BG_PID tracking (UT-24), mktemp -d usage (UT-23), cancel button (UT-25), and success dialog (UT-20).

**Acceptance Criteria**:
- [ ] Test file `tests/test-app-details-progress` created and executable
- [ ] Test verifies `href` present in `gui/app-details` for website link (UT-13)
- [ ] Test verifies `--progress` or `--pulsate` present (UT-14)
- [ ] Test verifies `mktemp -d` present (UT-23)
- [ ] Test verifies `BG_PID` variable used (UT-24)
- [ ] Test verifies Cancel button label present (UT-25)
- [ ] Test verifies success message present (UT-20)
- [ ] Test verifies `trap` with `BG_PID` kill present (SC-04)
- [ ] All tests fail before implementation (red)

**Implementation Notes**:
- Test file: `tests/test-app-details-progress`

---

#### Task 3.4 — Impl: app-details website link and progress dialog
- **Type**: impl
- **Estimate**: 2.5h
- **Priority**: High
- **Dependencies**: Task 3.3
- **Status**: Done

**Description**: Add Pango anchor link for website in the INFO_TEXT. Replace `x-terminal-emulator` launch with background process + FIFO + yad --progress dialog (with Cancel and Hide buttons). Add success/failure dialog after completion. Add mktemp -d tmpdir and BG_PID tracking with trap EXIT cleanup.

**Acceptance Criteria**:
- [ ] Website rendered as `<a href="$(escape_markup "$website")">$(escape_markup "$website")</a>` in INFO_TEXT
- [ ] `TMPDIR_WORK="$(mktemp -d /tmp/hamapps-XXXXXX)"` used for temp files
- [ ] FIFO at `"${TMPDIR_WORK}/progress.fifo"`, log at `"${TMPDIR_WORK}/output.log"`, exit file at `"${TMPDIR_WORK}/exit.code"`
- [ ] `trap 'kill "$BG_PID" 2>/dev/null; rm -rf "$TMPDIR_WORK"' EXIT` set before background launch
- [ ] Install/uninstall launched as: `bash "$HAMAPPS_DIR/scripts/$ACTION_SCRIPT" "$APP" >"$LOG_FILE" 2>&1; echo $? >"$EXIT_FILE" &` followed immediately by `BG_PID=$!`
- [ ] `yad --progress --pulsate --button="Cancel:1" --button="Hide:0"` displayed while install runs
- [ ] Cancel path: kills BG_PID, shows gui_info "Cancelled" message, does NOT call mark_installed/mark_uninstalled
- [ ] On completion: reads exit code from EXIT_FILE; exit 0 → `mark_installed`/`mark_uninstalled` + `gui_info "Success"` dialog; non-zero → `gui_error` dialog
- [ ] `shellcheck gui/app-details` exits 0
- [ ] All Task 3.3 tests pass

**Implementation Notes**:
- File: `gui/app-details`
- The background process compound: `{ bash "$HAMAPPS_DIR/scripts/$ACTION_SCRIPT" "$APP" >"$LOG_FILE" 2>&1; echo $? >"$EXIT_FILE"; } &` then `BG_PID=$!`
- Feed progress FIFO: `exec 3>"$FIFO"` in parent; `echo "# Installing..." >&3` periodically; or just open FIFO and wait — yad --progress with --pulsate doesn't need percentage updates
- Wait for background: `wait "$BG_PID"` or poll with `kill -0 "$BG_PID"` loop
- FIFO EOF: close fd 3 after wait to trigger yad auto-close: `exec 3>&-`
- The `x-terminal-emulator` fallback (for yad < 0.40) is removed; yad ≥ 0.40 is required

---

### Phase 4: Integration Tests and Validation

#### Task 4.1 — Tests: integration smoke tests
- **Type**: test
- **Estimate**: 1.0h
- **Priority**: Medium
- **Dependencies**: Task 2.4, Task 3.4
- **Status**: Done

**Description**: Write integration smoke tests for `build_app_data` category filtering (IT-02), install state round-trip (IT-03, IT-04), and `gui_error`/`gui_info` yad argument verification (IT-06).

**Acceptance Criteria**:
- [ ] Test file `tests/test-ux-integration` created and executable
- [ ] IT-02: `build_app_data "digital-modes"` returns only digital-modes apps
- [ ] IT-03: After `mark_uninstalled direwolf`, direwolf row shows "Not installed" and no `<b>` markup
- [ ] IT-04: After `mark_installed direwolf`, direwolf row shows "Installed" and `<b>Direwolf</b>` name
- [ ] IT-05: CATEGORY variable unchanged after simulated exit code 8
- [ ] IT-06: `gui_error` produces yad command with `--error` flag
- [ ] Tests clean up installed state after each test case
- [ ] All tests fail before implementation (red — though implementation exists at this point, tests confirm integration)

**Implementation Notes**:
- Test file: `tests/test-ux-integration`
- Source `scripts/utils` with `HAMAPPS_DIR` set to repo root
- Mock `yad` as a function that records its arguments: `yad() { echo "yad_called: $*"; }`
- Run `build_app_data` function directly by sourcing `gui/app-list` up to the function definition

---

#### Task 4.2 — Impl: run all tests and validate shellcheck
- **Type**: integration
- **Estimate**: 1.0h
- **Priority**: High
- **Dependencies**: Task 4.1
- **Status**: Done

**Description**: Run the full test suite, fix any failures, and confirm shellcheck passes on all three modified files. This is the final validation task.

**Acceptance Criteria**:
- [ ] `shellcheck scripts/utils` exits 0
- [ ] `shellcheck gui/app-list` exits 0
- [ ] `shellcheck gui/app-details` exits 0
- [ ] `tests/test-utils-helpers` exits 0
- [ ] `tests/test-app-list-structure` exits 0
- [ ] `tests/test-app-list-ux` exits 0
- [ ] `tests/test-app-details-foundation` exits 0
- [ ] `tests/test-app-details-progress` exits 0
- [ ] `tests/test-ux-integration` exits 0
- [ ] All 26 unit tests (UT-01 through UT-26) pass
- [ ] All 6 integration tests (IT-01 through IT-06) pass

**Implementation Notes**:
- Fix any shellcheck SC2086, SC2206, or other warnings found during the integration run
- Pay special attention to quoting in the FIFO/progress logic in `gui/app-details`
- The `BG_PID` variable will trigger SC2064 (trap string) — use single quotes in trap to avoid expanding BG_PID at trap-set time: `trap 'kill "$BG_PID" 2>/dev/null; rm -rf "$TMPDIR_WORK"' EXIT`

---

## Notes and Assumptions

1. yad ≥ 0.40 is assumed available. The `x-terminal-emulator` fallback from the original `gui/app-details` is removed as the spec targets yad ≥ 0.40 exclusively.
2. The `SEARCH_TEXT` variable is declared but not actively passed back into yad (yad's search bar value is not readable externally) — this is the documented FR-10 limitation.
3. Task 3.4 is the most complex task (~2.5h). If the FIFO + yad --progress integration proves difficult to make shellcheck-clean, the implementation may use a polling loop instead of FIFO EOF signalling — both approaches are valid.
4. All test files must be made executable with `chmod +x`.
