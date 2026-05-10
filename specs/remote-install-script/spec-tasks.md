# Implementation Tasks: Remote Install Script

**Spec**: specs/remote-install-script/spec.md
**Status**: Complete
**Total Estimated Effort**: 5h
**Critical Path**: 1.1 → 1.2 → 2.1 → 2.2

## Summary

- Total Tasks: 6
- Phases: 2
- Key Dependencies: test tasks precede matching impl tasks (TDD order)

## Tasks

### Phase 1: Core Script

#### Task 1.1 — Write failing tests for install.sh
- **Type**: test
- **Estimate**: 1.5h
- **Priority**: Critical
- **Dependencies**: None
- **Status**: complete

**Description**: Create `tests/test-install-sh` covering all 15 unit tests specified in the spec (UT-1 through UT-15). Tests must be written first and must fail (or be skipped) before `install.sh` exists, confirming the red phase of TDD. The test file follows the project pattern: `PASS:/FAIL:` lines, no external frameworks.

**Acceptance Criteria**:
- [ ] File `tests/test-install-sh` exists and is executable
- [ ] Running the test file before `install.sh` is created causes all tests (except UT-3 shellcheck, which needs the file) to print `FAIL:` or skip gracefully
- [ ] All 15 test cases (UT-1 through UT-15) are represented
- [ ] Tests follow the `pass()/fail()` counter pattern from existing test files
- [ ] Final summary line prints `Tests: N passed, M failed`
- [ ] File passes `shellcheck -x` itself

**Implementation Notes**:
- Pattern from `tests/test-utils-helpers` and `tests/test-trustedqsl-scripts`
- UT-3 (shellcheck) should skip with a message if shellcheck is not installed
- UT-6 through UT-10, UT-14, UT-15 (OS check and validation functions) require the script to expose testable functions; use a `HAMAPPS_TEST_MODE=1` guard or source-with-stubs approach
- UT-4 (--help) and UT-5 (--dry-run) call `bash install.sh <flag>` directly; these will fail if the file does not exist — test should print `FAIL: install.sh not found` gracefully
- All path references must use `REPO_ROOT` derived from `readlink -f`

---

#### Task 1.2 — Implement install.sh
- **Type**: impl
- **Estimate**: 2h
- **Priority**: Critical
- **Dependencies**: Task 1.1
- **Status**: complete

**Description**: Write `install.sh` at the repository root. The script must be fully self-contained, pass shellcheck, and satisfy all acceptance criteria in the spec. Implement all eight sections: shebang + safety flags, inline colour helpers, argument parsing, OS detection, dependency installation, directory validation + clone/update, PATH configuration, and success banner.

**Acceptance Criteria**:
- [ ] `install.sh` exists at repo root (`/home/parallels/git/kw4jlb/ham-apps/install.sh`)
- [ ] `shellcheck -x install.sh` exits 0
- [ ] `bash install.sh --help` exits 0 and prints "Usage:"
- [ ] `bash install.sh --dry-run` exits 0 on supported OS without filesystem changes
- [ ] Script uses `set -euo pipefail` as first executable line
- [ ] Script uses `trap` for any temp file cleanup
- [ ] Inline colour helpers (info/success/warning/error/die) match style of `scripts/utils`
- [ ] OS check rejects Debian 10 / Ubuntu 18.04; accepts Debian 11+, Ubuntu 20.04+
- [ ] `HAMAPPS_DIR` defaults to `$HOME/ham-apps`; honours env override
- [ ] `validate_install_dir()` rejects metacharacter-containing paths (CWE-78 mitigation)
- [ ] Clone skipped if `$HAMAPPS_DIR/.git` exists; dies if dir exists but is not a git repo
- [ ] PATH not duplicated in `~/.bashrc` on repeated runs
- [ ] `~/.zshrc` updated if it exists
- [ ] Success banner printed with installation directory and next steps
- [ ] `install.sh` is executable (`chmod +x`)
- [ ] No `source scripts/utils` or any external source calls
- [ ] `HAMAPPS_REPO` env var overrides clone URL (for testing with local path)

**Implementation Notes**:
- File location: `/home/parallels/git/kw4jlb/ham-apps/install.sh`
- Expose testable functions (OS check, validation) via a `HAMAPPS_TEST_MODE` guard so tests can source-and-call without triggering the full install
- OS detection uses `/etc/os-release` (source it with `. /etc/os-release`); compare VERSION_ID numerically
- For Ubuntu version comparison, split on `.` and compare major then minor integers
- PATH append pattern: `grep -qF "export PATH=\"$HAMAPPS_DIR" ~/.bashrc || echo "..." >> ~/.bashrc`
- Dry-run: set a `DRY_RUN=1` variable; wrap side-effect commands in a `run_cmd()` helper that either executes or prints `[DRY-RUN]`
- `HAMAPPS_REPO` default: `https://github.com/KW4JLB/ham-apps.git`

---

### Phase 2: Validation & Quality

#### Task 2.1 — Validate install.sh passes all tests
- **Type**: integration
- **Estimate**: 0.5h
- **Priority**: High
- **Dependencies**: Task 1.1, Task 1.2
- **Status**: complete

**Description**: Run `tests/test-install-sh` against the completed `install.sh` and verify all 15 tests pass. Run shellcheck on both files. Confirm --help and --dry-run work correctly on the current host OS.

**Acceptance Criteria**:
- [ ] `bash tests/test-install-sh` prints only `PASS:` lines (zero `FAIL:` lines)
- [ ] `shellcheck -x install.sh` exits 0
- [ ] `shellcheck -x tests/test-install-sh` exits 0
- [ ] `bash install.sh --help` exit code 0
- [ ] `bash install.sh --dry-run` exit code 0 (on Debian/Ubuntu host)

**Implementation Notes**:
- Run from repo root: `cd /home/parallels/git/kw4jlb/ham-apps && bash tests/test-install-sh`
- If any FAIL: lines appear, loop back to Task 1.2 iteration

---

#### Task 2.2 — Validate idempotency and PATH deduplication
- **Type**: integration
- **Estimate**: 0.5h
- **Priority**: High
- **Dependencies**: Task 2.1
- **Status**: complete

**Description**: Verify that running `install.sh --dry-run` twice on the same machine does not produce duplicate PATH entries. Check that UT-11 (no duplicate PATH) passes. Verify UT-12 (HAMAPPS_DIR override) and UT-14/UT-15 (validation) pass.

**Acceptance Criteria**:
- [ ] UT-11 passes: after simulated double PATH-append, no duplicate line in test bashrc
- [ ] UT-12 passes: HAMAPPS_DIR env var is respected
- [ ] UT-14 passes: metacharacter path rejected
- [ ] UT-15 passes: clean absolute path accepted

**Implementation Notes**:
- Tests should use temp files for `~/.bashrc` simulation, not the real one
- UT-14/UT-15 are covered by the test file; running `bash tests/test-install-sh` covers both

---

#### Task 2.3 — Set file permissions and final shellcheck
- **Type**: setup
- **Estimate**: 0.25h
- **Priority**: Medium
- **Dependencies**: Task 2.1
- **Status**: complete

**Description**: Ensure `install.sh` has the executable bit set and that both `install.sh` and `tests/test-install-sh` have no shellcheck warnings. Confirm `install.sh` is tracked in git.

**Acceptance Criteria**:
- [ ] `chmod +x install.sh` applied
- [ ] `chmod +x tests/test-install-sh` applied
- [ ] `git add install.sh tests/test-install-sh` — both files staged
- [ ] `shellcheck -x install.sh` exits 0
- [ ] `shellcheck -x tests/test-install-sh` exits 0
- [ ] UT-1 (shebang), UT-2 (set -euo pipefail), UT-3 (shellcheck), UT-13 (executable) all pass

**Implementation Notes**:
- Check permissions with `stat -c '%a' install.sh` — must include execute bits
- Final confirmation: `bash tests/test-install-sh` shows 15/15 PASS

---

#### Task 2.4 — Documentation update
- **Type**: docs
- **Estimate**: 0.25h
- **Priority**: Medium
- **Dependencies**: Task 2.2
- **Status**: complete

**Description**: Ensure the README (if present) or any docs/ file that describes installation references the new one-liner. If no README exists, skip silently.

**Acceptance Criteria**:
- [ ] If `README.md` exists at repo root, it includes the `curl -fsSL` one-liner
- [ ] If no README.md exists, this task is marked complete with a note

**Implementation Notes**:
- Check with `ls /home/parallels/git/kw4jlb/ham-apps/README.md`
- If README.md exists, add a brief "Quick Install" section near the top
- Do not create a new README.md if it does not exist (not in scope)

---
