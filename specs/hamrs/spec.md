# HAMRS App Entry for ham-apps

## Overview

Add HAMRS to the ham-apps app manager as an installable entry under `apps/hamrs/`. HAMRS is a simple portable logbook tailored for POTA (Parks on the Air), Field Day, and general ham radio QSO logging. It is distributed as an AppImage from the HAMRS S3 bucket (`hamrs-dist.s3.amazonaws.com`), not from GitHub.

### Background

The GitHub repository listed in third-party references (KD0YQ/hamrs) returns 404. The authoritative distribution is from the HAMRS website at https://hamrs.app/, which serves AppImages directly from AWS S3 at the pattern:

```
https://hamrs-dist.s3.amazonaws.com/hamrs-pro-<version>-linux-<arch>.AppImage
```

Supported Linux architectures: `x86_64`, `arm64`, `armv7l`.

Current version as of spec date: **2.49.0** (branding: `hamrs-pro`).

### Scope

Four files under `apps/hamrs/`:
- `metadata` — key=value app metadata
- `description` — plain text description
- `install` — bash install script
- `uninstall` — bash uninstall script

No GUI changes. No changes to `data/categories` (logging category already exists). No changes to core scripts.

---

## Requirements

### Functional Requirements

| ID   | Requirement | Priority |
|------|-------------|----------|
| FR-1 | `apps/hamrs/metadata` must contain name, category, website, tags, min-os fields | Critical |
| FR-2 | `apps/hamrs/description` must have first line as short summary, followed by longer description | Critical |
| FR-3 | `install` must detect the host architecture (x86_64, arm64, armv7l) and download the matching AppImage | Critical |
| FR-4 | `install` must download the AppImage from the S3 URL pattern | Critical |
| FR-5 | `install` must place the AppImage at `/opt/hamrs/hamrs.AppImage` | Critical |
| FR-6 | `install` must make the AppImage executable | Critical |
| FR-7 | `install` must create a wrapper script at `/usr/local/bin/hamrs` | Critical |
| FR-8 | `install` must create a `.desktop` file at `/usr/local/share/applications/hamrs.desktop` | Critical |
| FR-9 | `install` must call `mark_installed hamrs` on success | Critical |
| FR-10 | `uninstall` must remove `/opt/hamrs/`, the wrapper at `/usr/local/bin/hamrs`, and the `.desktop` file | Critical |
| FR-11 | `uninstall` must call `mark_uninstalled hamrs` on success | Critical |
| FR-12 | Both scripts must source `scripts/utils` for logging and helper functions | Critical |
| FR-13 | `install` must use a temp dir with `trap` for cleanup on exit | High |
| FR-14 | `install` must install `fuse` or `libfuse2` if not present (AppImage dependency) | High |
| FR-15 | `install` must call `update-desktop-database` after installing the `.desktop` file | Medium |
| FR-16 | `uninstall` must call `update-desktop-database` after removing the `.desktop` file | Medium |
| FR-17 | Hardcoded version string must be accompanied by `# TODO: dynamic version detection` comment | Critical |

### Non-Functional Requirements

| ID    | Requirement | Target |
|-------|-------------|--------|
| NFR-1 | Scripts must pass `shellcheck` with no errors | All scripts |
| NFR-2 | Scripts must use `set -euo pipefail` | Both install and uninstall |
| NFR-3 | `sudo` used inside app scripts only — not by the core runner | Convention |
| NFR-4 | No hardcoded credentials or tokens | Security |
| NFR-5 | Install must be idempotent: re-running install over an existing install must not error | High |

### Constraints

- Target OS: Debian 11+, Ubuntu 20.04+
- Shell: bash only — no Python, no compiled code
- Paths must be resolved relative to `$HAMAPPS_DIR` for core paths; `/opt/hamrs/` and `/usr/local/` are system paths installed via `sudo`
- `$HAMAPPS_DIR` is derived via `dirname` chain from `$0` — same pattern as `apps/trustedqsl/install`
- Category must be `logging` (already defined in `data/categories`)

---

## Design

### Architecture

The HAMRS app entry is a self-contained set of four text files. No new modules or shared code are introduced. The install and uninstall scripts follow the identical structural pattern as `apps/trustedqsl/install` and `apps/trustedqsl/uninstall`.

```
apps/hamrs/
  metadata     — static key=value file
  description  — static text file
  install      — bash script (create dirs, download, chmod, write wrapper, write .desktop)
  uninstall    — bash script (rm -rf dirs, rm files, update-desktop-database)
```

### Install Flow

```
install
  ├── source scripts/utils
  ├── set HAMRS_VERSION (hardcoded, with TODO comment)
  ├── detect ARCH (uname -m → map to AppImage arch suffix)
  ├── set HAMRS_URL from version + arch
  ├── check/install fuse dependency (apt-get install -y libfuse2 or fuse)
  ├── create /opt/hamrs/ (sudo mkdir -p)
  ├── mktemp tmpdir + trap cleanup
  ├── curl download AppImage to tmpdir
  ├── sudo mv AppImage to /opt/hamrs/hamrs.AppImage
  ├── sudo chmod +x /opt/hamrs/hamrs.AppImage
  ├── sudo write /usr/local/bin/hamrs wrapper
  ├── sudo chmod +x /usr/local/bin/hamrs
  ├── sudo write /usr/local/share/applications/hamrs.desktop
  ├── sudo update-desktop-database
  ├── mark_installed hamrs
  └── success message
```

### Uninstall Flow

```
uninstall
  ├── source scripts/utils
  ├── info message
  ├── sudo rm -rf /opt/hamrs/
  ├── sudo rm -f /usr/local/bin/hamrs
  ├── sudo rm -f /usr/local/share/applications/hamrs.desktop
  ├── sudo update-desktop-database
  ├── mark_uninstalled hamrs
  └── success message
```

### Architecture Detection

```bash
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)          APPIMAGE_ARCH="x86_64" ;;
  aarch64|arm64)   APPIMAGE_ARCH="arm64" ;;
  armv7l)          APPIMAGE_ARCH="armv7l" ;;
  *)
    error "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac
```

### Download URL Pattern

```
https://hamrs-dist.s3.amazonaws.com/hamrs-pro-${HAMRS_VERSION}-linux-${APPIMAGE_ARCH}.AppImage
```

### Wrapper Script (`/usr/local/bin/hamrs`)

```bash
#!/bin/bash
exec /opt/hamrs/hamrs.AppImage "$@"
```

### Desktop Entry (`/usr/local/share/applications/hamrs.desktop`)

```ini
[Desktop Entry]
Name=HAMRS
Comment=Portable ham radio logbook for POTA, Field Day, and more
Exec=/usr/local/bin/hamrs
Icon=hamrs
Type=Application
Categories=HamRadio;Utility;
Terminal=false
StartupNotify=true
```

### Metadata File

```
name=HAMRS
category=logging
website=https://hamrs.app/
tags=pota,field-day,logging,qso,portable,appimage
min-os=Debian 11, Ubuntu 20.04
```

### Description File

```
Simple portable logbook tailored for POTA, Field Day, and more.
HAMRS is a cross-platform ham radio logging application optimised for portable
operating. It supports Parks on the Air (POTA), Field Day, SOTA, and general
QSO logging with ADIF export. The Linux version is distributed as an AppImage
and requires no system-wide installation of dependencies beyond FUSE.
```

---

## Test Specification

This spec produces static text files and bash scripts. The test suite uses `shellcheck` for static analysis and a smoke test harness that mocks `sudo`, `curl`, `apt-get`, and `uname` to exercise the install/uninstall scripts without root or network access.

### Unit Tests

**T-01 — shellcheck passes on install script**
- Given: `apps/hamrs/install` exists
- When: `shellcheck apps/hamrs/install` is run
- Then: exit code 0, no errors or warnings

**T-02 — shellcheck passes on uninstall script**
- Given: `apps/hamrs/uninstall` exists
- When: `shellcheck apps/hamrs/uninstall` is run
- Then: exit code 0, no errors or warnings

**T-03 — metadata file has required fields**
- Given: `apps/hamrs/metadata` exists
- When: each required field is grepped (name, category, website, tags, min-os)
- Then: all five fields are present with non-empty values

**T-04 — metadata category matches known category**
- Given: `data/categories` and `apps/hamrs/metadata`
- When: the category field value is compared to category IDs in `data/categories`
- Then: `logging` appears as a category ID in `data/categories`

**T-05 — description first line is short summary**
- Given: `apps/hamrs/description`
- When: the first line is read
- Then: it is non-empty and ≤120 characters

**T-06 — install script sources scripts/utils**
- Given: `apps/hamrs/install`
- When: the file is grepped for `source.*scripts/utils`
- Then: the pattern is found

**T-07 — install script has set -euo pipefail**
- Given: `apps/hamrs/install`
- When: the file is grepped for `set -euo pipefail`
- Then: the pattern is found

**T-08 — install script has trap for temp dir cleanup**
- Given: `apps/hamrs/install`
- When: the file is grepped for `trap`
- Then: the pattern is found

**T-09 — install script has TODO comment for hardcoded version**
- Given: `apps/hamrs/install`
- When: the file is grepped for `# TODO: dynamic version detection`
- Then: the pattern is found

**T-10 — install script handles x86_64 architecture**
- Given: a mock environment where `uname -m` returns `x86_64`
- When: the ARCH detection case statement is evaluated
- Then: `APPIMAGE_ARCH` is set to `x86_64`

**T-11 — install script handles arm64/aarch64 architecture**
- Given: a mock environment where `uname -m` returns `aarch64`
- When: the ARCH detection case statement is evaluated
- Then: `APPIMAGE_ARCH` is set to `arm64`

**T-12 — install script exits on unsupported architecture**
- Given: a mock environment where `uname -m` returns `riscv64`
- When: the ARCH detection case statement is evaluated
- Then: script exits with non-zero status and prints an error

**T-13 — uninstall script sources scripts/utils**
- Given: `apps/hamrs/uninstall`
- When: the file is grepped for `source.*scripts/utils`
- Then: the pattern is found

**T-14 — uninstall script has set -euo pipefail**
- Given: `apps/hamrs/uninstall`
- When: the file is grepped for `set -euo pipefail`
- Then: the pattern is found

**T-15 — install script calls mark_installed**
- Given: `apps/hamrs/install`
- When: the file is grepped for `mark_installed hamrs`
- Then: the pattern is found

**T-16 — uninstall script calls mark_uninstalled**
- Given: `apps/hamrs/uninstall`
- When: the file is grepped for `mark_uninstalled hamrs`
- Then: the pattern is found

### Integration Tests

**T-17 — smoke install (dry run with mocks)**
- Given: mock `sudo` (no-op), mock `curl` (copies a fixture AppImage), mock `apt-get` (no-op), mock `uname` returning `x86_64`, mock `update-desktop-database` (no-op)
- When: `apps/hamrs/install` is run in a temp HAMAPPS_DIR
- Then: exit code 0; `/opt/hamrs/hamrs.AppImage` exists; `/usr/local/bin/hamrs` exists; `/usr/local/share/applications/hamrs.desktop` exists; `~/.local/share/ham-apps/installed/hamrs` exists

**T-18 — smoke uninstall (after install)**
- Given: state from T-17
- When: `apps/hamrs/uninstall` is run
- Then: exit code 0; `/opt/hamrs/` removed; `/usr/local/bin/hamrs` removed; `.desktop` file removed; `~/.local/share/ham-apps/installed/hamrs` removed

### Acceptance Tests

**T-19 — `ham-apps list` shows hamrs**
- Given: ham-apps project with `apps/hamrs/` populated
- When: `scripts/list-apps` is run
- Then: output includes `hamrs` with name `HAMRS`

**T-20 — `ham-apps install hamrs` succeeds (dry run)**
- Given: mock environment
- When: `scripts/install-app hamrs` is called
- Then: `apps/hamrs/install` is executed and exits 0

---

## Security & Compliance

### Threat Model

The install script downloads a binary (AppImage) from an S3 URL over HTTPS. Risks:

1. **Supply chain — unsigned AppImage**: HAMRS does not publish checksums or GPG signatures for AppImages. Mitigation: download over HTTPS with `curl -fsSL` (fail on HTTP errors, follow redirects, silent); note this limitation in the script comments.
2. **Hardcoded version**: version string is hardcoded with a TODO for dynamic detection. Using a pinned version reduces risk of pulling an unknown release, at the cost of needing manual updates.
3. **Privilege escalation via sudo**: `sudo` is used to write to `/opt/hamrs/`, `/usr/local/bin/`, and `/usr/local/share/applications/`. This is standard for system-wide app installs and follows the existing trustedqsl pattern.
4. **Temp dir injection**: temp dir is created with `mktemp -d` (random suffix) and cleaned up via `trap`. No user-supplied paths are used in the temp dir path.

### Security Controls

- Download uses HTTPS (S3 endpoint) with `curl -fsSL` — fails on non-2xx, follows redirects
- No credentials stored in scripts
- No user input passed to shell commands without quoting
- All variable references quoted (`"$ARCH"`, `"$HAMRS_URL"`, etc.)
- Temp dir uses `mktemp -d` (no predictable path)
- AppImage placed in `/opt/hamrs/` (root-owned, not world-writable)

### Compliance Requirements

- No personal data collected or stored
- OWASP A06 (Vulnerable Components): version is pinned, TODO comment for future dynamic detection
- OWASP A08 (Software Integrity): no checksum verification possible without upstream support — documented as known limitation

---

## Implementation Plan

### Phase 1: Static Files

#### Task 1.1 — Write metadata file
Create `apps/hamrs/metadata` with correct key=value fields.

#### Task 1.2 — Write description file
Create `apps/hamrs/description` with short first-line summary and longer description.

### Phase 2: Scripts

#### Task 2.1 — Write tests for install script
Create `tests/hamrs/test_install.sh` with shellcheck and smoke tests for the install script.

#### Task 2.2 — Write install script
Create `apps/hamrs/install` implementing the full install flow described in Design section.

#### Task 2.3 — Write tests for uninstall script
Create `tests/hamrs/test_uninstall.sh` with shellcheck and smoke tests for the uninstall script.

#### Task 2.4 — Write uninstall script
Create `apps/hamrs/uninstall` implementing the full uninstall flow.

---

## Deployment

### Deployment Steps

1. Files are delivered as part of the ham-apps git repository — no separate deployment step
2. Users run `ham-apps install hamrs` which invokes `apps/hamrs/install`
3. Script requires internet access to reach `hamrs-dist.s3.amazonaws.com`

### Rollback Plan

Run `ham-apps uninstall hamrs` to reverse the install. The uninstall script removes all created files. No system packages are installed by HAMRS install (only `libfuse2` which is a safe, commonly-installed package).

### Monitoring

Not applicable — this is a local desktop application install. The `mark_installed`/`mark_uninstalled` state files in `~/.local/share/ham-apps/installed/` serve as the install state tracking mechanism.

---

## Acceptance Criteria

- [ ] `apps/hamrs/metadata` exists with name, category=logging, website, tags, min-os fields
- [ ] `apps/hamrs/description` exists; first line is ≤120 chars; body describes POTA/Field Day use
- [ ] `apps/hamrs/install` passes `shellcheck` with no errors
- [ ] `apps/hamrs/install` uses `set -euo pipefail`
- [ ] `apps/hamrs/install` sources `scripts/utils`
- [ ] `apps/hamrs/install` has `trap` for temp dir cleanup
- [ ] `apps/hamrs/install` has `# TODO: dynamic version detection` comment near version string
- [ ] `apps/hamrs/install` detects architecture (x86_64, arm64, armv7l) and exits on unsupported arch
- [ ] `apps/hamrs/install` downloads AppImage from S3 URL over HTTPS
- [ ] `apps/hamrs/install` installs AppImage to `/opt/hamrs/hamrs.AppImage`
- [ ] `apps/hamrs/install` creates wrapper at `/usr/local/bin/hamrs`
- [ ] `apps/hamrs/install` creates `.desktop` file at `/usr/local/share/applications/hamrs.desktop`
- [ ] `apps/hamrs/install` calls `mark_installed hamrs`
- [ ] `apps/hamrs/uninstall` passes `shellcheck` with no errors
- [ ] `apps/hamrs/uninstall` uses `set -euo pipefail`
- [ ] `apps/hamrs/uninstall` sources `scripts/utils`
- [ ] `apps/hamrs/uninstall` removes `/opt/hamrs/`, wrapper, and `.desktop` file
- [ ] `apps/hamrs/uninstall` calls `mark_uninstalled hamrs`
- [ ] Smoke install test (T-17) passes
- [ ] Smoke uninstall test (T-18) passes

---

## References

- Reference implementation: `apps/trustedqsl/install`, `apps/trustedqsl/uninstall`
- Shared utilities: `scripts/utils`
- Category definitions: `data/categories`
- HAMRS download page: https://hamrs.app/
- HAMRS S3 distribution: `https://hamrs-dist.s3.amazonaws.com/hamrs-pro-<version>-linux-<arch>.AppImage`
- Confirmed S3 URLs (version 2.49.0):
  - `https://hamrs-dist.s3.amazonaws.com/hamrs-pro-2.49.0-linux-x86_64.AppImage`
  - `https://hamrs-dist.s3.amazonaws.com/hamrs-pro-2.49.0-linux-arm64.AppImage`
  - `https://hamrs-dist.s3.amazonaws.com/hamrs-pro-2.49.0-linux-armv7l.AppImage`
