# Implementation Tasks: DXSpider App Entry

**Spec**: specs/dxspider/spec.md
**Status**: Complete
**Total Estimated Effort**: 3.5h
**Critical Path**: 1.1 → 1.2 → 2.1 → 2.2 → 2.3 → 3.1 → 3.2 → 4.1 → 4.2

## Summary

- Total Tasks: 9
- Phases: 4
- Key Dependencies: Test files (Phase 1) must exist before implementation (Phases 2-3); all implementation tasks must complete before validation (Phase 4)

---

## Tasks

### Phase 1: Test Files (TDD Red Phase)

#### Task 1.1 — Write test-dxspider-metadata
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Create `tests/test-dxspider-metadata` as a standalone bash test script following the exact pattern of `tests/test-hamrs-metadata`. All tests must be in the red phase (failing) because the app files don't exist yet. The script must print `PASS:` / `FAIL:` lines and exit non-zero if any fail.

**Acceptance Criteria**:
- [ ] File exists at `tests/test-dxspider-metadata`
- [ ] File is executable (`chmod +x`)
- [ ] Tests 1-12 match the spec test specification (metadata/description/install/uninstall existence, field values, category in data/categories, first-line length, DX cluster mention)
- [ ] Script exits non-zero when run against empty repo (red phase)
- [ ] Pattern matches `tests/test-hamrs-metadata` exactly in structure

**Implementation Notes**:
- Copy pattern from `tests/test-hamrs-metadata`
- Test 4 checks `dx-cluster|` in `data/categories`
- Test 10 checks for "DX cluster" or "DXSpider" in description
- Test 2 checks `name=DXSpider`
- Test 3 checks `category=dx-cluster`
- Test 5 checks `website=https://www.dxcluster.org/`

---

#### Task 1.2 — Write test-dxspider-scripts
- **Type**: test
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: Complete

**Description**: Create `tests/test-dxspider-scripts` as a standalone bash test script following the exact pattern of `tests/test-hamrs-scripts`. All tests must be in the red phase (failing). Covers both install and uninstall script structure and shellcheck compliance.

**Acceptance Criteria**:
- [ ] File exists at `tests/test-dxspider-scripts`
- [ ] File is executable (`chmod +x`)
- [ ] Install tests 1-11 match spec: set -euo pipefail, sources utils, trap, git clone URL, /home/sysop/spider, useradd, systemd service path, systemctl enable, mark_installed, does NOT contain `systemctl start`, shellcheck
- [ ] Uninstall tests 12-19 match spec: set -euo pipefail, sources utils, /home/sysop/spider removal, service file removal, systemctl stop+disable, warning about sysop user, mark_uninstalled, shellcheck
- [ ] Script exits non-zero when run against empty repo (red phase)
- [ ] Pattern matches `tests/test-hamrs-scripts` exactly in structure

**Implementation Notes**:
- Test 10 (no autostart): `grep -v 'systemctl start dxspider'` or negative grep — check that `systemctl start dxspider` does NOT appear in install
- Test for useradd: `grep -q 'useradd' "$INSTALL"`
- Test for warning about sysop user: `grep -q 'sysop' "$UNINSTALL"` combined with `grep -q 'warning' "$UNINSTALL"`

---

### Phase 2: Static Files

#### Task 2.1 — Add dx-cluster category to data/categories
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.1 (test for category exists)
- **Status**: Complete

**Description**: Append the `dx-cluster` category line to `data/categories`. Must follow the existing pipe-delimited format.

**Acceptance Criteria**:
- [ ] `data/categories` contains line: `dx-cluster|DX Cluster|DX cluster node and spotting network software`
- [ ] Line is appended after the last existing entry
- [ ] No existing lines are modified
- [ ] File ends with a newline

**Implementation Notes**:
- Exact line: `dx-cluster|DX Cluster|DX cluster node and spotting network software`
- File path: `/home/parallels/git/kw4jlb/ham-apps/data/categories`

---

#### Task 2.2 — Create apps/dxspider/metadata
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 2.1
- **Status**: Complete

**Description**: Create `apps/dxspider/metadata` with the correct key=value fields.

**Acceptance Criteria**:
- [ ] File exists at `apps/dxspider/metadata`
- [ ] Contains `name=DXSpider`
- [ ] Contains `category=dx-cluster`
- [ ] Contains `website=https://www.dxcluster.org/`
- [ ] Contains `tags=` with at least one tag
- [ ] Contains `min-os=Debian 11, Ubuntu 20.04`

**Implementation Notes**:
- Exact content per spec Design section:
  ```
  name=DXSpider
  category=dx-cluster
  website=https://www.dxcluster.org/
  tags=dx-cluster,dx-spots,packet,cluster,node,perl
  min-os=Debian 11, Ubuntu 20.04
  ```

---

#### Task 2.3 — Create apps/dxspider/description
- **Type**: impl
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 2.2
- **Status**: Complete

**Description**: Create `apps/dxspider/description` as plain text. First line is the short summary (≤120 chars). Must mention DX cluster.

**Acceptance Criteria**:
- [ ] File exists at `apps/dxspider/description`
- [ ] First line is non-empty and ≤120 characters
- [ ] File mentions "DX cluster" or "DXSpider"

**Implementation Notes**:
- First line: `Open-source DX cluster node software that connects to the worldwide DX spotting network.` (89 chars — within limit)
- Multi-line plain text body per spec Design section

---

### Phase 3: Scripts

#### Task 3.1 — Create apps/dxspider/install
- **Type**: impl
- **Estimate**: 1h
- **Priority**: Critical
- **Dependencies**: Task 1.2, Task 2.2, Task 2.3
- **Status**: Complete

**Description**: Create `apps/dxspider/install` as an executable bash script following the Install Script Design in the spec. Must pass shellcheck and all test-dxspider-scripts install tests.

**Acceptance Criteria**:
- [ ] File exists at `apps/dxspider/install`
- [ ] File is executable
- [ ] Contains `set -euo pipefail`
- [ ] Sources `$HAMAPPS_DIR/scripts/utils`
- [ ] Installs Perl deps via apt (`perl`, `libnet-telnet-perl`, `libdigest-sha-perl`)
- [ ] Creates `sysop` system user idempotently using `useradd -r -m -s /bin/bash sysop`
- [ ] Clones `https://github.com/f1evm/dxspider` to `/home/sysop/spider` (or pulls if already cloned)
- [ ] Contains `trap` for cleanup on failure
- [ ] Emits `warning` after clone about supply chain verification
- [ ] Writes systemd unit to `/etc/systemd/system/dxspider.service`
- [ ] Runs `systemctl daemon-reload` and `systemctl enable dxspider`
- [ ] Does NOT contain `systemctl start dxspider`
- [ ] Calls `mark_installed dxspider`
- [ ] Passes `shellcheck -x` with zero errors

**Implementation Notes**:
- HAMAPPS_DIR resolution: `HAMAPPS_DIR="$(dirname "$(dirname "$(dirname "$(readlink -f "$0")")")")"`
- Idempotent user check: `id -u sysop &>/dev/null || sudo useradd -r -m -s /bin/bash sysop`
- Idempotent clone/pull: check for `/home/sysop/spider/.git`
- Trap: `TMP_SPIDER_CLONE` not needed since clone goes directly to destination; use trap to remove partial clone dir on EXIT if clone fails
- Systemd unit content per spec Design section (cluster.pl entrypoint)
- warning log line: `warning "DXSpider cloned from https://github.com/f1evm/dxspider — verify the remote and review code before starting the service."`

---

#### Task 3.2 — Create apps/dxspider/uninstall
- **Type**: impl
- **Estimate**: 0.5h
- **Priority**: Critical
- **Dependencies**: Task 3.1
- **Status**: Complete

**Description**: Create `apps/dxspider/uninstall` as an executable bash script following the Uninstall Script Design in the spec. Must pass shellcheck and all test-dxspider-scripts uninstall tests.

**Acceptance Criteria**:
- [ ] File exists at `apps/dxspider/uninstall`
- [ ] File is executable
- [ ] Contains `set -euo pipefail`
- [ ] Sources `$HAMAPPS_DIR/scripts/utils`
- [ ] Stops and disables `dxspider` service if active/enabled (gracefully, no error if not running)
- [ ] Removes `/etc/systemd/system/dxspider.service`
- [ ] Runs `systemctl daemon-reload`
- [ ] Emits `warning` about data loss before removing `/home/sysop/spider`
- [ ] Removes `/home/sysop/spider`
- [ ] Emits `warning` advising operator to manually remove `sysop` user if desired
- [ ] Removes `libnet-telnet-perl libdigest-sha-perl` via apt (NOT `perl` itself)
- [ ] Calls `mark_uninstalled dxspider`
- [ ] Passes `shellcheck -x` with zero errors

**Implementation Notes**:
- Service stop pattern (graceful): use `systemctl is-active` / `systemctl is-enabled` with `|| true`
- Do NOT remove `perl` package — only DXSpider-specific modules
- HAMAPPS_DIR resolution same as install script

---

### Phase 4: Validation

#### Task 4.1 — Validate metadata tests pass
- **Type**: integration
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.1, Task 2.1, Task 2.2, Task 2.3
- **Status**: Complete

**Description**: Run `bash tests/test-dxspider-metadata` and confirm all tests pass (exit 0, all PASS lines).

**Acceptance Criteria**:
- [ ] `bash tests/test-dxspider-metadata` exits 0
- [ ] All lines print `PASS:`
- [ ] Zero `FAIL:` lines

**Implementation Notes**:
- Run from repo root: `HAMAPPS_DIR=$PWD bash tests/test-dxspider-metadata`
- If any tests fail, fix the underlying file (metadata, description, categories) and re-run

---

#### Task 4.2 — Validate scripts tests pass
- **Type**: integration
- **Estimate**: 0.25h
- **Priority**: Critical
- **Dependencies**: Task 1.2, Task 3.1, Task 3.2
- **Status**: Complete

**Description**: Run `bash tests/test-dxspider-scripts` and confirm all tests pass (exit 0, all PASS lines). Also run shellcheck directly on both scripts.

**Acceptance Criteria**:
- [ ] `bash tests/test-dxspider-scripts` exits 0
- [ ] All lines print `PASS:`
- [ ] Zero `FAIL:` lines
- [ ] `shellcheck -x apps/dxspider/install` produces zero output
- [ ] `shellcheck -x apps/dxspider/uninstall` produces zero output

**Implementation Notes**:
- Run from repo root
- If shellcheck fails, fix scripts and re-run
- shellcheck may require `# shellcheck source=scripts/utils` directive in scripts

---
