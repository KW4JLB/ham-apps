# Implementation Tasks: Trusted QSL App Addition

**Spec**: specs/trustedqsl/spec.md
**Status**: In Progress
**Total Estimated Effort**: 2h
**Critical Path**: 1.1 → 1.2 → 2.1 → 2.2 → 2.3 → 2.4

## Summary
- Total Tasks: 6
- Phases: 2
- Key Dependencies: Phase 2 tasks depend on Phase 1 tests existing; 2.1-2.4 are independent of each other but all depend on 1.1 and 1.2 being written first (TDD order)

---

## Tasks

### Phase 1: Test Writing (TDD)

#### Task 1.1 — Write metadata and file-existence tests
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Create `tests/test-trustedqsl-metadata` — a bash test script that checks all static file properties of the `apps/trustedqsl/` app entry. Tests must fail until Phase 2 creates the app files.

**Acceptance Criteria**:
- [ ] File `tests/test-trustedqsl-metadata` exists and is executable
- [ ] Test verifies `apps/trustedqsl/metadata` exists
- [ ] Test verifies metadata contains `name=Trusted QSL`
- [ ] Test verifies metadata contains `category=tools`
- [ ] Test verifies metadata contains a `website=https://` field
- [ ] Test verifies metadata contains a `tags=` field with at least one tag
- [ ] Test verifies metadata contains a `min-os=` field
- [ ] Test verifies `apps/trustedqsl/description` exists
- [ ] Test verifies description first line is non-empty
- [ ] Test verifies `apps/trustedqsl/install` exists and is executable
- [ ] Test verifies `apps/trustedqsl/uninstall` exists and is executable
- [ ] All tests FAIL when run before app files are created (red phase confirmed)
- [ ] Uses pass/fail counter pattern matching `tests/test-utils-helpers`

**Implementation Notes**:
- Follow the exact same pass/fail counter structure as `tests/test-utils-helpers`
- Use `REPO_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"` for path resolution
- Use `[[ -f path ]]` for file existence, `[[ -x path ]]` for executability
- Use `grep -q 'pattern' file` for content checks
- Do NOT use `apt-get` or `sudo` — static checks only

---

#### Task 1.2 — Write script structure and shellcheck tests
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Create `tests/test-trustedqsl-scripts` — a bash test script that verifies the install/uninstall scripts contain the required structural elements and pass shellcheck linting. Tests must fail until Phase 2 creates the scripts.

**Acceptance Criteria**:
- [ ] File `tests/test-trustedqsl-scripts` exists and is executable
- [ ] Test verifies install contains `set -euo pipefail`
- [ ] Test verifies install sources `scripts/utils`
- [ ] Test verifies install calls `sudo apt-get install -y trustedqsl`
- [ ] Test verifies install calls `mark_installed trustedqsl`
- [ ] Test verifies install passes `shellcheck` with zero errors
- [ ] Test verifies uninstall contains `set -euo pipefail`
- [ ] Test verifies uninstall sources `scripts/utils`
- [ ] Test verifies uninstall calls `sudo apt-get remove -y trustedqsl`
- [ ] Test verifies uninstall calls `mark_uninstalled trustedqsl`
- [ ] Test verifies uninstall passes `shellcheck` with zero errors
- [ ] All tests FAIL when run before app files are created (red phase confirmed)
- [ ] Uses pass/fail counter pattern matching `tests/test-utils-helpers`

**Implementation Notes**:
- Use `grep -qF 'pattern' file` for fixed-string content checks
- For shellcheck: `shellcheck "$INSTALL_SCRIPT"` — check exit code
- Skip shellcheck test gracefully if shellcheck is not installed (warn, don't fail)
- Do NOT run install or uninstall scripts (no apt-get, no sudo)

---

### Phase 2: App File Implementation

#### Task 2.1 — Create apps/trustedqsl/metadata
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.1, Task 1.2
- **Status**: Complete

**Description**: Create the `apps/trustedqsl/metadata` key=value file with all required fields.

**Acceptance Criteria**:
- [ ] File exists at `apps/trustedqsl/metadata`
- [ ] Contains `name=Trusted QSL`
- [ ] Contains `category=tools`
- [ ] Contains `website=https://sourceforge.net/projects/trustedqsl/`
- [ ] Contains `tags=lotw,qsl,digital-signature,logging`
- [ ] Contains `min-os=Debian 11, Ubuntu 20.04`
- [ ] Task 1.1 metadata-related tests pass after this file is created

**Implementation Notes**:
- Plain text key=value, no quotes around values
- Match the exact format of `apps/wsjtx/metadata`
- File path: `apps/trustedqsl/metadata`

---

#### Task 2.2 — Create apps/trustedqsl/description
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.1
- **Status**: Complete

**Description**: Create the `apps/trustedqsl/description` plain-text description file. First line must be the short one-line summary.

**Acceptance Criteria**:
- [ ] File exists at `apps/trustedqsl/description`
- [ ] First line reads: `Open-source tool for digital signatures supporting the LoTW QSL system.`
- [ ] Body explains tQSL's purpose (LoTW, certificate management, log signing)
- [ ] Task 1.1 description-related tests pass after this file is created

**Implementation Notes**:
- Plain text, no markdown
- First line = short summary (used by GUI and `list-apps`)
- Match the multi-line format of `apps/wsjtx/description`

---

#### Task 2.3 — Create apps/trustedqsl/install
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.1, Task 1.2
- **Status**: Complete

**Description**: Create the `apps/trustedqsl/install` bash script. Must be executable and follow all conventions.

**Acceptance Criteria**:
- [ ] File exists at `apps/trustedqsl/install` and is executable (`chmod +x`)
- [ ] First line is `#!/bin/bash`
- [ ] Uses `set -euo pipefail`
- [ ] Sets `HAMAPPS_DIR` via `readlink -f` three-level traversal
- [ ] Sources `$HAMAPPS_DIR/scripts/utils`
- [ ] Calls `info "Installing Trusted QSL..."`
- [ ] Calls `sudo apt-get install -y trustedqsl`
- [ ] Calls `mark_installed trustedqsl`
- [ ] Calls `success "Trusted QSL installed."`
- [ ] Passes `shellcheck` with zero errors
- [ ] Task 1.1 and 1.2 install-related tests pass after this file is created

**Implementation Notes**:
- Must be identical in structure to `apps/wsjtx/install` plus `mark_installed` call
- The `mark_installed` call goes AFTER the apt-get succeeds (before success message is fine)
- File path: `apps/trustedqsl/install`
- Set executable: `chmod +x apps/trustedqsl/install`

---

#### Task 2.4 — Create apps/trustedqsl/uninstall
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.1, Task 1.2
- **Status**: Complete

**Description**: Create the `apps/trustedqsl/uninstall` bash script. Must be executable and follow all conventions.

**Acceptance Criteria**:
- [ ] File exists at `apps/trustedqsl/uninstall` and is executable (`chmod +x`)
- [ ] First line is `#!/bin/bash`
- [ ] Uses `set -euo pipefail`
- [ ] Sets `HAMAPPS_DIR` via `readlink -f` three-level traversal
- [ ] Sources `$HAMAPPS_DIR/scripts/utils`
- [ ] Calls `info "Removing Trusted QSL..."`
- [ ] Calls `sudo apt-get remove -y trustedqsl`
- [ ] Calls `mark_uninstalled trustedqsl`
- [ ] Calls `success "Trusted QSL removed."`
- [ ] Passes `shellcheck` with zero errors
- [ ] Task 1.1 and 1.2 uninstall-related tests pass after this file is created

**Implementation Notes**:
- Must be identical in structure to `apps/wsjtx/uninstall` plus `mark_uninstalled` call
- The `mark_uninstalled` call goes AFTER the apt-get remove succeeds
- File path: `apps/trustedqsl/uninstall`
- Set executable: `chmod +x apps/trustedqsl/uninstall`
