# Remote Install Script for ham-apps

## Overview

### Purpose
Provide a `curl | bash` one-liner install experience for ham-apps, allowing users to bootstrap the tool on a fresh Debian or Ubuntu system with a single command:

```
curl -fsSL https://raw.githubusercontent.com/KW4JLB/ham-apps/main/install.sh | bash
```

### Scope
- A single self-contained shell script `install.sh` at the repository root
- OS compatibility checking (Debian 11+ and Ubuntu 20.04+)
- Dependency installation (git, yad)
- Repository cloning to `~/ham-apps` or a user-specified directory via `HAMAPPS_DIR` env var
- PATH configuration in `~/.bashrc` and optionally `~/.zshrc`
- Friendly success message with next steps
- A shellcheck-passing, TDD-tested implementation

### Background
ham-apps is a YAD-based app manager for Amateur Radio software. New users need a simple bootstrap path since the project is not in standard package manager repositories. The install script must be fully self-contained because it runs before the repo is cloned; it cannot source `scripts/utils`.

---

## Requirements

### Functional Requirements

| ID   | Requirement                                                                                     | Priority |
|------|-------------------------------------------------------------------------------------------------|----------|
| FR-1 | Detect Debian version ≥ 11 (bullseye) or Ubuntu version ≥ 20.04 (focal); exit 1 with an error message on unsupported OS | Must     |
| FR-2 | Install `git` and `yad` via `apt-get` if not already present; use `sudo` only for apt commands  | Must     |
| FR-3 | Clone `https://github.com/KW4JLB/ham-apps.git` to the target directory                        | Must     |
| FR-4 | Default clone target is `$HOME/ham-apps`; honor `HAMAPPS_DIR` env var if set                  | Must     |
| FR-5 | If target directory already exists and is a valid git repo, print a warning and skip re-clone  | Must     |
| FR-6 | Append PATH export line to `~/.bashrc` if not already present                                  | Must     |
| FR-7 | If `~/.zshrc` exists, append the same PATH export line if not already present                  | Must     |
| FR-8 | Print a success banner with next steps (how to run `ham-apps`, how to open the GUI)            | Must     |
| FR-9 | Support a `--dry-run` flag that prints all actions without executing them                      | Should   |
| FR-10| Support a `--help` / `-h` flag that prints usage and exits 0                                   | Should   |

### Non-Functional Requirements

| ID    | Requirement                                                                       | Target              |
|-------|-----------------------------------------------------------------------------------|---------------------|
| NFR-1 | Script must pass `shellcheck -x` with zero warnings                              | Zero shellcheck warnings |
| NFR-2 | Script must use `set -euo pipefail` at the top                                   | Mandatory           |
| NFR-3 | Script must clean up any temp files via `trap ... EXIT`                          | Mandatory           |
| NFR-4 | Script must be fully self-contained (no `source scripts/utils`)                  | Mandatory           |
| NFR-5 | All output uses colour-coded prefixes matching the style in `scripts/utils`      | Consistent UX       |
| NFR-6 | Script must complete successfully on a clean Debian 11 and Ubuntu 20.04 install  | Compatibility       |
| NFR-7 | Script must not require running as root (uses sudo internally where needed)      | Security            |
| NFR-8 | Duplicate PATH entries must not be added on repeated runs                         | Idempotency         |

### Constraints

- Pure bash — no Python, no compiled helpers
- Target: Debian 11+ and Ubuntu 20.04+ only; non-Debian/Ubuntu systems are rejected
- Script lives at `install.sh` in repo root so it is served at the expected raw GitHub URL
- No hardcoded version strings that would need updating on each release (OS version floor is fixed in the spec)
- `sudo` is used only for `apt-get` calls; all other operations run as the invoking user

---

## Design

### Architecture

The script is a single bash file with the following logical sections:

```
install.sh
  1. Shebang + set -euo pipefail
  2. Colour helpers (inline — no source)
  3. Argument parsing (--help, --dry-run)
  4. OS detection + version check
  5. Dependency installation (git, yad)
  6. Clone / update repo
  7. PATH configuration
  8. Success message
```

No external tools are required beyond what is available on a minimal Debian/Ubuntu system (`bash`, `apt-get`, `git`, `grep`, `sed`, `lsb_release` or `/etc/os-release`).

### Inline Colour Helpers

Because `scripts/utils` cannot be sourced, the script must define its own logging functions:

```bash
info()    { echo -e "\e[34m[INFO]\e[0m  $*"; }
success() { echo -e "\e[32m[OK]\e[0m    $*"; }
warning() { echo -e "\e[33m[WARN]\e[0m  $*" >&2; }
error()   { echo -e "\e[31m[ERROR]\e[0m $*" >&2; }
die()     { error "$*"; exit 1; }
```

### OS Detection Logic

1. Source `/etc/os-release` to read `ID` and `VERSION_ID`
2. If `ID == "debian"`: require `VERSION_ID` integer ≥ 11
3. If `ID == "ubuntu"`: require `VERSION_ID` float ≥ 20.04 (compare major.minor)
4. Otherwise: call `die` with an unsupported OS message listing supported systems
5. The check uses only `/etc/os-release` (universally available on systemd-based Debian/Ubuntu); `lsb_release` is not required

### Dependency Installation

```bash
install_deps() {
    local missing=()
    command -v git &>/dev/null || missing+=(git)
    command -v yad &>/dev/null || missing+=(yad)
    if [[ ${#missing[@]} -gt 0 ]]; then
        info "Installing missing packages: ${missing[*]}"
        sudo apt-get update -qq
        sudo apt-get install -y "${missing[@]}"
    fi
}
```

### Install Directory Validation

Before using `HAMAPPS_DIR` in any file operation or appending it to shell RC files, it must be validated:

```bash
validate_install_dir() {
    local dir="$1"
    # Reject paths containing shell metacharacters: $ ( ) ` ; | & < > " ' \
    if [[ "$dir" =~ [\$\(\)\`\;\|\&\<\>\"\'\\] ]]; then
        die "HAMAPPS_DIR contains invalid characters. Use a plain filesystem path."
    fi
    # Must be an absolute path or start with ~ (home-relative)
    if [[ "$dir" != /* && "$dir" != "~"* ]]; then
        die "HAMAPPS_DIR must be an absolute path (e.g. /home/user/ham-apps)."
    fi
}
```

This function is called immediately after `HAMAPPS_DIR` is resolved (before any cloning or RC-file modification).

### Clone / Update Logic

```
if [[ -d "$HAMAPPS_DIR/.git" ]]; then
    warning "ham-apps already cloned at $HAMAPPS_DIR — skipping clone"
elif [[ -d "$HAMAPPS_DIR" ]]; then
    die "Directory $HAMAPPS_DIR exists but is not a git repo. Remove it or set HAMAPPS_DIR."
else
    git clone https://github.com/KW4JLB/ham-apps.git "$HAMAPPS_DIR"
fi
```

### PATH Configuration

The export line appended to shell RC files:

```bash
export PATH="$HAMAPPS_DIR:$PATH"
```

Written as a literal line using the resolved value of `HAMAPPS_DIR`. Before appending, grep the RC file for the export line; skip if found.

Both `~/.bashrc` and `~/.zshrc` (if it exists) are updated.

### Dry-Run Mode

When `--dry-run` is passed, all actions are printed as `[DRY-RUN] would: <action>` without executing. The script exits 0 after printing all planned actions.

### Success Banner

```
============================================================
  ham-apps installed successfully!
============================================================

  Installation directory : ~/ham-apps
  Added to PATH in       : ~/.bashrc

  To start using ham-apps:
    Reload your shell:    source ~/.bashrc
    Launch the GUI:       ham-apps gui
    List apps:            ham-apps list

  Happy DXing! 73 de KW4JLB
============================================================
```

---

## Test Specification

### Unit Tests (`tests/test-install-sh`)

All tests run without network access or sudo. They test static properties of the script and logic via source-with-stubs.

#### UT-1: Script has correct shebang
- **Given** `install.sh` exists at repo root
- **When** first line is read
- **Then** it equals `#!/usr/bin/env bash` or `#!/bin/bash`

#### UT-2: set -euo pipefail is present
- **Given** `install.sh` exists
- **When** its content is grepped for `set -euo pipefail`
- **Then** the pattern is found

#### UT-3: shellcheck passes
- **Given** `install.sh` exists and shellcheck is installed
- **When** `shellcheck -x install.sh` is run
- **Then** exit code is 0 (no warnings or errors)

#### UT-4: --help flag exits 0 and prints usage
- **Given** `install.sh` is executable
- **When** `bash install.sh --help` is run (without network/sudo)
- **Then** exit code is 0 and stdout contains "Usage:"

#### UT-5: --dry-run flag exits 0 without making changes
- **Given** `install.sh` is executable and no network/sudo available
- **When** `bash install.sh --dry-run` is run with a mock OS release
- **Then** exit code is 0 and no files are created

#### UT-6: OS detection rejects unsupported distro
- **Given** `/etc/os-release` is mocked with `ID=fedora VERSION_ID=38`
- **When** the OS check function is sourced and called
- **Then** it calls `die` (exits non-zero)

#### UT-7: OS detection accepts Debian 11
- **Given** `/etc/os-release` is mocked with `ID=debian VERSION_ID=11`
- **When** the OS check function is evaluated
- **Then** it does not exit

#### UT-8: OS detection accepts Ubuntu 20.04
- **Given** `/etc/os-release` is mocked with `ID=ubuntu VERSION_ID=20.04`
- **When** the OS check function is evaluated
- **Then** it does not exit

#### UT-9: OS detection rejects Debian 10
- **Given** `/etc/os-release` is mocked with `ID=debian VERSION_ID=10`
- **When** the OS check function is evaluated
- **Then** it exits non-zero

#### UT-10: OS detection rejects Ubuntu 18.04
- **Given** `/etc/os-release` is mocked with `ID=ubuntu VERSION_ID=18.04`
- **When** the OS check function is evaluated
- **Then** it exits non-zero

#### UT-11: PATH line not duplicated on second run
- **Given** `~/.bashrc` already contains the ham-apps PATH export
- **When** the PATH configuration function is called again
- **Then** the line count for the export in `~/.bashrc` does not increase

#### UT-12: HAMAPPS_DIR env var is honoured
- **Given** `HAMAPPS_DIR=/tmp/custom-ham-apps` is exported
- **When** the default directory logic is evaluated
- **Then** the clone target is `/tmp/custom-ham-apps`

#### UT-13: Script is executable
- **Given** `install.sh` in repo root
- **When** file permissions are checked
- **Then** the executable bit is set

#### UT-14: validate_install_dir rejects paths with shell metacharacters
- **Given** `HAMAPPS_DIR` is set to `/tmp/$(evil)` (contains a subshell expansion)
- **When** `validate_install_dir` is called
- **Then** the function exits non-zero with an error message about invalid characters

#### UT-15: validate_install_dir accepts a clean absolute path
- **Given** `HAMAPPS_DIR` is set to `/home/user/ham-apps`
- **When** `validate_install_dir` is called
- **Then** the function exits 0 without error

### Integration Tests

#### IT-1: End-to-end dry-run on current OS
- **Given** the current host is Debian 11+ or Ubuntu 20.04+
- **When** `bash install.sh --dry-run` is run
- **Then** exit code is 0, output includes planned actions, no files are mutated

### Acceptance Tests

#### AT-1: Full install smoke test (manual, CI optional)
- **Given** a clean Debian 11 or Ubuntu 20.04 container/VM without git/yad
- **When** `curl -fsSL .../install.sh | bash` is run
- **Then** `~/ham-apps` contains the cloned repo, `ham-apps` is in PATH after `source ~/.bashrc`, exit 0

---

## Security & Compliance

### Threat Model

| Threat | Mitigation |
|--------|------------|
| Script tampering via MITM | HTTPS enforced via `curl -fsSL`; no HTTP fallback |
| Privilege escalation | `sudo` used only for `apt-get`; all other ops as invoking user |
| Arbitrary code execution from malicious `HAMAPPS_DIR` | `validate_install_dir()` rejects paths containing `$ ( ) \` ; \| & < > " ' \` via regex; called before any file operation or RC-file write |
| Overwriting an existing non-git directory | Script aborts with `die` if `$HAMAPPS_DIR` exists but is not a git repo |
| RC file corruption | Uses `>>` append-only; checks for existing line before appending |

### Security Controls

- `set -euo pipefail` ensures unexpected failures abort the script
- `HAMAPPS_DIR` is used only as a path argument to `git clone` and `echo`; not evaluated via `eval` or passed to a shell
- No credentials, tokens, or secrets in the script
- Script does not disable or modify system package verification

### Compliance Requirements

- Shellcheck SC2086 (double-quote variables): all variables are quoted
- Shellcheck SC2164 (cd without error check): not applicable (no `cd` used)
- No `eval`, no `source` of remote content

---

## Implementation Plan

### Phase 1: Core Script (`install.sh`)
- Task 1: Write `install.sh` with all sections (colour helpers, arg parsing, OS check, deps, clone, PATH, banner)
- Acceptance: shellcheck passes, --help works without network

### Phase 2: Tests (`tests/test-install-sh`)
- Task 2: Write bash test script covering UT-1 through UT-13
- Acceptance: tests run, all defined PASS/FAIL lines printed correctly

### Task Overview

| ID  | Task                                  | Phase | Dependencies |
|-----|---------------------------------------|-------|--------------|
| T-1 | Write `install.sh`                    | 1     | none         |
| T-2 | Write `tests/test-install-sh`         | 2     | T-1          |

### Configuration

- `HAMAPPS_DIR`: env var (default `$HOME/ham-apps`)
- `HAMAPPS_REPO`: env var (default `https://github.com/KW4JLB/ham-apps.git`) — allows testing with a local path

---

## Deployment

### Deployment Steps

1. Merge `install.sh` to the `main` branch
2. The raw GitHub URL `https://raw.githubusercontent.com/KW4JLB/ham-apps/main/install.sh` becomes live automatically

### Rollback Plan

- Revert the commit adding `install.sh` to `main`; the raw URL will 404 until a new commit is pushed

### Monitoring

- No runtime monitoring required for a static script
- Users can report issues via GitHub Issues

---

## Acceptance Criteria

- [ ] `install.sh` exists at repo root and is executable (`chmod +x`)
- [ ] `shellcheck -x install.sh` exits 0 with no warnings
- [ ] `bash install.sh --help` exits 0 and prints usage information
- [ ] `bash install.sh --dry-run` exits 0 on a supported OS without making filesystem changes
- [ ] Script rejects Debian 10 and Ubuntu 18.04 with a clear error message
- [ ] Script accepts Debian 11, Debian 12, Ubuntu 20.04, Ubuntu 22.04, Ubuntu 24.04
- [ ] `HAMAPPS_DIR` env var overrides the default clone directory
- [ ] PATH line is not duplicated if `install.sh` is run twice
- [ ] All tests in `tests/test-install-sh` print only `PASS:` lines (no `FAIL:`)
- [ ] Script does not contain `eval` or remote `source` calls
- [ ] Script does not run as root (no `sudo` at the top level for the entire script)
- [ ] `validate_install_dir` rejects `HAMAPPS_DIR` values containing shell metacharacters (UT-14 passes)
- [ ] `validate_install_dir` accepts clean absolute paths (UT-15 passes)

---

## References

- `scripts/utils` — logging helper patterns replicated inline in install.sh
- `apps/trustedqsl/install` — example of `set -euo pipefail`, `trap`, and `sudo apt-get` usage
- `tests/test-utils-helpers` — pattern for PASS:/FAIL: bash test scripts
- Existing spec pattern: `specs/hamrs/spec.md`, `specs/trustedqsl/spec.md`
- GitHub raw URL format: `https://raw.githubusercontent.com/{owner}/{repo}/{branch}/{path}`
