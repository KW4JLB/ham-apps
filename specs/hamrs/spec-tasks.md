# Implementation Tasks: HAMRS App Entry

**Spec**: specs/hamrs/spec.md
**Status**: Complete
**Total Estimated Effort**: 3h
**Critical Path**: 1.1 → 1.2 → 2.1 → 2.2 → 2.3 → 2.4

## Summary

- Total Tasks: 6
- Phases: 2
- Key Dependencies: Task 2.1 (tests) must precede Task 2.2 (install impl); Task 2.3 (tests) must precede Task 2.4 (uninstall impl)

## Tasks

### Phase 1: Static Files

#### Task 1.1 — Write metadata file
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Create `apps/hamrs/metadata` with the key=value fields required by the ham-apps metadata format. Must match the exact format of `apps/trustedqsl/metadata`.

**Acceptance Criteria**:
- [ ] File exists at `apps/hamrs/metadata`
- [ ] Contains `name=HAMRS`
- [ ] Contains `category=logging`
- [ ] Contains `website=https://hamrs.app/`
- [ ] Contains `tags=` with at least one tag (pota, field-day, logging, qso, portable, appimage)
- [ ] Contains `min-os=Debian 11, Ubuntu 20.04`

**Implementation Notes**:
- File path: `apps/hamrs/metadata`
- Format: key=value, one per line, no quotes
- Category must match an ID in `data/categories` — use `logging`

---

#### Task 1.2 — Write description file
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Create `apps/hamrs/description` with a short one-line summary as the first line, followed by a longer description. Follows the same format as `apps/trustedqsl/description`.

**Acceptance Criteria**:
- [ ] File exists at `apps/hamrs/description`
- [ ] First line is non-empty and ≤120 characters
- [ ] Body mentions POTA, Field Day, and ADIF export
- [ ] Body mentions AppImage distribution

**Implementation Notes**:
- File path: `apps/hamrs/description`
- Plain text only — no markdown formatting
- First line used as short summary by `scripts/list-apps` and GUI

---

### Phase 2: Scripts (TDD Order)

#### Task 2.1 — Write tests for install and metadata (TDD)
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: Task 1.1, Task 1.2
- **Status**: Complete

**Description**: Create `tests/test-hamrs-metadata` and `tests/test-hamrs-scripts` following the exact patterns of `tests/test-trustedqsl-metadata` and `tests/test-trustedqsl-scripts`. Tests must fail (red phase) when `apps/hamrs/install` and `apps/hamrs/uninstall` do not yet exist, and pass (green phase) after implementation.

**Acceptance Criteria**:
- [ ] `tests/test-hamrs-metadata` exists and is executable
- [ ] `tests/test-hamrs-scripts` exists and is executable
- [ ] Both test files follow the `pass()`/`fail()` pattern from trustedqsl tests
- [ ] `test-hamrs-metadata` tests: metadata exists, name=HAMRS, category=logging, website, tags, min-os, description exists and first line non-empty, install and uninstall exist and are executable
- [ ] `test-hamrs-scripts` tests: install has `set -euo pipefail`, install sources utils, install has TODO comment, install calls mark_installed, install has trap, uninstall has `set -euo pipefail`, uninstall sources utils, uninstall calls mark_uninstalled, both pass shellcheck
- [ ] Running `tests/test-hamrs-metadata` before Task 1.1/1.2 implementation fails on missing files (but Tasks 1.1/1.2 precede this, so after they exist metadata tests pass, script tests that check install/uninstall files fail until 2.2/2.4)
- [ ] Running `tests/test-hamrs-scripts` before Task 2.2/2.4 fails on missing install/uninstall files

**Implementation Notes**:
- Test files: `tests/test-hamrs-metadata`, `tests/test-hamrs-scripts`
- Must be executable (`chmod +x`)
- Follow naming convention: no `.sh` extension (matches existing tests)
- Use `REPO_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"` to locate repo root
- shellcheck test must check both install and uninstall with `shellcheck -x`
- Add test: install script has `# TODO: dynamic version detection`
- Add test: install script has `trap` keyword
- Add test: install script downloads from `hamrs-dist.s3.amazonaws.com`
- Add test: uninstall script removes `/opt/hamrs/`
- Add test: desktop file content check — install script creates a `.desktop` file

---

#### Task 2.2 — Write install script
- **Type**: impl
- **Estimate**: 1h
- **Priority**: Critical
- **Dependencies**: Task 2.1
- **Status**: Complete

**Description**: Create `apps/hamrs/install` implementing the full install flow from the spec Design section. Must make all tests in `tests/test-hamrs-scripts` pass.

**Acceptance Criteria**:
- [ ] File exists at `apps/hamrs/install`
- [ ] File is executable
- [ ] Uses `set -euo pipefail`
- [ ] Sources `scripts/utils` via `$HAMAPPS_DIR`
- [ ] Has `# TODO: dynamic version detection` comment next to `HAMRS_VERSION`
- [ ] Version hardcoded to `2.49.0`
- [ ] Detects architecture with `uname -m` and maps to x86_64/arm64/armv7l
- [ ] Exits with error on unsupported architecture
- [ ] Constructs download URL from `hamrs-dist.s3.amazonaws.com` pattern
- [ ] Installs `libfuse2` or `fuse` dependency via apt-get
- [ ] Creates `/opt/hamrs/` with `sudo mkdir -p`
- [ ] Uses `mktemp -d` for temp dir and `trap` for cleanup
- [ ] Downloads AppImage with `curl -fsSL`
- [ ] Moves AppImage to `/opt/hamrs/hamrs.AppImage` with sudo
- [ ] Makes AppImage executable with `sudo chmod +x`
- [ ] Creates wrapper script at `/usr/local/bin/hamrs` with `sudo tee`
- [ ] Makes wrapper executable with `sudo chmod +x`
- [ ] Creates `.desktop` file at `/usr/local/share/applications/hamrs.desktop` with `sudo tee`
- [ ] Calls `sudo update-desktop-database` (with `|| true` for non-fatal)
- [ ] Calls `mark_installed hamrs`
- [ ] Calls `success` with version message
- [ ] Passes `shellcheck -x` with no errors

**Implementation Notes**:
- File path: `apps/hamrs/install`
- `$HAMAPPS_DIR` derived from `$(dirname "$(dirname "$(dirname "$(readlink -f "$0")")")")`
- Use `sudo tee` for writing files to system paths (not `echo > sudo` which doesn't work)
- Pattern for wrapper: `printf '#!/bin/bash\nexec /opt/hamrs/hamrs.AppImage "$@"\n' | sudo tee /usr/local/bin/hamrs > /dev/null`
- Pattern for `.desktop`: use heredoc piped to `sudo tee`
- FUSE dependency: check if `libfuse2` is available in apt cache; if not try `fuse`; same pattern as trustedqsl's wx package detection
- Download URL: `https://hamrs-dist.s3.amazonaws.com/hamrs-pro-${HAMRS_VERSION}-linux-${APPIMAGE_ARCH}.AppImage`

---

#### Task 2.3 — Write tests for uninstall (TDD)
- **Type**: test
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 2.1
- **Status**: Complete

**Description**: The uninstall tests are already included in `tests/test-hamrs-scripts` (created in Task 2.1). This task verifies that the uninstall-specific tests are complete and will fail before Task 2.4 creates the uninstall script. No new test file is needed if `test-hamrs-scripts` already covers uninstall thoroughly.

**Acceptance Criteria**:
- [ ] `tests/test-hamrs-scripts` contains tests verifying uninstall has `set -euo pipefail`
- [ ] `tests/test-hamrs-scripts` contains tests verifying uninstall sources `scripts/utils`
- [ ] `tests/test-hamrs-scripts` contains tests verifying uninstall calls `mark_uninstalled hamrs`
- [ ] `tests/test-hamrs-scripts` contains tests verifying uninstall removes `/opt/hamrs/`
- [ ] `tests/test-hamrs-scripts` contains test that uninstall passes shellcheck
- [ ] Running the tests before Task 2.4 produces failures on uninstall-related tests

**Implementation Notes**:
- If Task 2.1 fully covered uninstall tests, this task is a verification pass only
- Add any missing uninstall-specific tests to `tests/test-hamrs-scripts`

---

#### Task 2.4 — Write uninstall script
- **Type**: impl
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: Task 2.3, Task 2.2
- **Status**: Complete

**Description**: Create `apps/hamrs/uninstall` implementing the full uninstall flow from the spec Design section. Must make all uninstall-related tests in `tests/test-hamrs-scripts` pass.

**Acceptance Criteria**:
- [ ] File exists at `apps/hamrs/uninstall`
- [ ] File is executable
- [ ] Uses `set -euo pipefail`
- [ ] Sources `scripts/utils` via `$HAMAPPS_DIR`
- [ ] Calls `info` with removal message
- [ ] Removes `/opt/hamrs/` with `sudo rm -rf`
- [ ] Removes `/usr/local/bin/hamrs` with `sudo rm -f`
- [ ] Removes `/usr/local/share/applications/hamrs.desktop` with `sudo rm -f`
- [ ] Calls `sudo update-desktop-database` (with `|| true` for non-fatal)
- [ ] Calls `mark_uninstalled hamrs`
- [ ] Calls `success` with completion message
- [ ] Passes `shellcheck -x` with no errors

**Implementation Notes**:
- File path: `apps/hamrs/uninstall`
- `$HAMAPPS_DIR` derived from `$(dirname "$(dirname "$(dirname "$(readlink -f "$0")")")")`
- `sudo rm -rf /opt/hamrs/` is safe since it's a directory created exclusively by the install script
- `update-desktop-database` may not be present on all systems — use `|| true` to avoid script failure

---
