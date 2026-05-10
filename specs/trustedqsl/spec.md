# Trusted QSL — App Addition Spec

## Overview

### Purpose
Add Trusted QSL (`trustedqsl`) as a new installable application in the ham-apps app manager. Trusted QSL is an open-source tool that provides digital signature support for the ARRL Logbook of The World (LoTW) QSL confirmation system. It is the required client for signing and uploading ADIF log files to LoTW.

### Scope
Create the four required files under `apps/trustedqsl/`:
- `metadata` — key=value app descriptor
- `description` — plain text description (first line is short summary)
- `install` — bash install script
- `uninstall` — bash uninstall script

No GUI, no core script, and no `data/categories` changes are needed. The `tools` category already exists.

### Background
The `trustedqsl` package is available in the official Debian 11+ and Ubuntu 20.04+ apt repositories. The preferred install path is therefore `sudo apt-get install -y trustedqsl`, matching the canonical pattern used by `apps/wsjtx/install`. After installation, `mark_installed trustedqsl` is called; after removal, `mark_uninstalled trustedqsl` is called.

---

## Requirements

### Functional Requirements

| ID   | Requirement                                                                 | Priority |
|------|-----------------------------------------------------------------------------|----------|
| FR-1 | `apps/trustedqsl/metadata` must contain valid key=value fields              | Critical |
| FR-2 | `apps/trustedqsl/description` first line must be the short summary          | Critical |
| FR-3 | `apps/trustedqsl/install` must install the `trustedqsl` apt package         | Critical |
| FR-4 | `apps/trustedqsl/install` must call `mark_installed trustedqsl` on success  | Critical |
| FR-5 | `apps/trustedqsl/uninstall` must remove the `trustedqsl` apt package        | Critical |
| FR-6 | `apps/trustedqsl/uninstall` must call `mark_uninstalled trustedqsl` on success | Critical |
| FR-7 | `ham-apps list` must include trustedqsl in its output after files are added  | High     |

### Non-Functional Requirements

| ID    | Requirement                                                              | Target        |
|-------|--------------------------------------------------------------------------|---------------|
| NFR-1 | Scripts must pass `shellcheck` with no errors                            | Zero errors   |
| NFR-2 | Scripts must use `set -euo pipefail`                                     | Mandatory     |
| NFR-3 | Scripts must source `scripts/utils` for logging and helpers              | Mandatory     |
| NFR-4 | Scripts must use `sudo` internally (not called as root by core runner)   | Mandatory     |
| NFR-5 | No hardcoded version strings in scripts without a TODO comment           | Mandatory     |
| NFR-6 | Temp dirs (if any) cleaned up with `trap`                                | Mandatory     |

### Constraints

- No Python, no compiled code in core scripts
- Target OS: Debian 11+ and Ubuntu 20.04+
- Package name in apt repos: `trustedqsl`
- App slug must match directory name: `trustedqsl`
- `min-os` in metadata must be Debian 11, Ubuntu 20.04 (or equivalent expression)
- Category must reference an existing category id from `data/categories` — use `tools`

---

## Design

### Architecture

The new app follows the identical structure and conventions of all existing apps, specifically `apps/wsjtx/` as the canonical reference:

```
apps/
  trustedqsl/
    metadata     # key=value: name, category, website, tags, min-os
    description  # plain text; line 1 = short summary
    install      # bash: set -euo pipefail; source utils; apt-get install; mark_installed
    uninstall    # bash: set -euo pipefail; source utils; apt-get remove; mark_uninstalled
```

The `HAMAPPS_DIR` variable is resolved using `readlink -f` traversal from the script's own path (three levels up: script → app-dir → apps → repo root). This pattern is used verbatim in all existing install/uninstall scripts.

### Data Models

**metadata file** (key=value, no quotes):
```
name=Trusted QSL
category=tools
website=https://sourceforge.net/projects/trustedqsl/
tags=lotw,qsl,digital-signature,logging
min-os=Debian 11, Ubuntu 20.04
```

**description file** (plain text):
```
Open-source tool for digital signatures supporting the LoTW QSL system.
Trusted QSL (tQSL) is the required client for creating and uploading digitally
signed ADIF log files to the ARRL Logbook of The World (LoTW) QSL confirmation
system. It supports certificate management, log signing, and direct LoTW upload.
```

**install script** (bash):
```bash
#!/bin/bash
# Install Trusted QSL

set -euo pipefail
HAMAPPS_DIR="$(dirname "$(dirname "$(dirname "$(readlink -f "$0")")")")"
source "$HAMAPPS_DIR/scripts/utils"

info "Installing Trusted QSL..."
sudo apt-get install -y trustedqsl
mark_installed trustedqsl
success "Trusted QSL installed."
```

**uninstall script** (bash):
```bash
#!/bin/bash
# Uninstall Trusted QSL

set -euo pipefail
HAMAPPS_DIR="$(dirname "$(dirname "$(dirname "$(readlink -f "$0")")")")"
source "$HAMAPPS_DIR/scripts/utils"

info "Removing Trusted QSL..."
sudo apt-get remove -y trustedqsl
mark_uninstalled trustedqsl
success "Trusted QSL removed."
```

### API Design

Not applicable — this is a bash app entry, not an API.

---

## Test Specification

All tests are bash scripts placed in `tests/` following the existing naming convention (`test-trustedqsl-*`). Tests use the same pass/fail counter pattern as `tests/test-utils-helpers`.

### Unit Tests — `tests/test-trustedqsl-metadata`

Tests that verify static file correctness without executing external commands.

**Given** the `apps/trustedqsl/` directory exists,
**When** each required file is read,
**Then** the fields and content conform to the spec.

Scenarios:
1. `metadata` file exists at `apps/trustedqsl/metadata`
2. `metadata` contains `name=Trusted QSL`
3. `metadata` contains `category=tools`
4. `metadata` contains `website=` starting with `https://`
5. `metadata` contains `tags=` with at least one tag
6. `metadata` contains `min-os=` field
7. `description` file exists at `apps/trustedqsl/description`
8. `description` first line is non-empty (short summary)
9. `install` file exists and is executable
10. `uninstall` file exists and is executable

### Integration Tests — `tests/test-trustedqsl-scripts`

Tests that verify script structure and shellcheck compliance without running apt or sudo.

**Given** the install and uninstall scripts exist,
**When** each script is inspected and linted,
**Then** all structural requirements are met.

Scenarios:
1. `install` contains `set -euo pipefail`
2. `install` sources `scripts/utils`
3. `install` calls `sudo apt-get install -y trustedqsl`
4. `install` calls `mark_installed trustedqsl`
5. `install` passes `shellcheck` with zero errors
6. `uninstall` contains `set -euo pipefail`
7. `uninstall` sources `scripts/utils`
8. `uninstall` calls `sudo apt-get remove -y trustedqsl`
9. `uninstall` calls `mark_uninstalled trustedqsl`
10. `uninstall` passes `shellcheck` with zero errors

### Acceptance Tests

**Given** the `trustedqsl` app files are in place,
**When** `scripts/list-apps` is executed,
**Then** `trustedqsl` appears in the output.

---

## Security and Compliance

### Threat Model

- The install script runs `sudo apt-get install` — this is identical to all other apt-based apps; trust is delegated to the system's configured apt sources.
- No credentials, API keys, or tokens are involved.
- No user input is processed by these scripts; all inputs are fixed strings.
- The `mark_installed` function creates a file in `~/.local/share/ham-apps/installed/` — no privilege escalation beyond the `sudo apt-get` call.

### Security Controls

- `set -euo pipefail` prevents silent failures and unbound variable usage.
- `sudo` is scoped only to the `apt-get` call; no blanket root execution.
- No external downloads with unverified checksums — apt handles package verification via GPG-signed repos.
- No eval of user-supplied input.

### Compliance Requirements

- OWASP A03 (Injection): no user input evaluated as code — met.
- OWASP A05 (Security Misconfiguration): no hardcoded credentials — met.
- OWASP A08 (Software Integrity): apt's GPG verification handles supply chain — met.

---

## Implementation Plan

### Phases

**Phase 1 — Test Files** (TDD: write failing tests first)
- Task 1.1: Write `tests/test-trustedqsl-metadata` (static file checks — all fail until files exist)
- Task 1.2: Write `tests/test-trustedqsl-scripts` (shellcheck + structural checks)

**Phase 2 — App Files** (implementation to make tests pass)
- Task 2.1: Create `apps/trustedqsl/metadata`
- Task 2.2: Create `apps/trustedqsl/description`
- Task 2.3: Create `apps/trustedqsl/install`
- Task 2.4: Create `apps/trustedqsl/uninstall`

### Configuration

No configuration required. The app relies entirely on apt's package management.

---

## Deployment

### Deployment Steps

1. Merge PR containing `apps/trustedqsl/` to `main`.
2. Users run `ham-apps update` (git pull) to receive the new app.
3. Users run `ham-apps install trustedqsl` or use the GUI to install.

### Rollback Plan

Remove `apps/trustedqsl/` directory from the repository and push a revert commit.

### Monitoring

Not applicable — this is a local app manager with no server-side telemetry.

---

## Acceptance Criteria

- [ ] `apps/trustedqsl/metadata` exists with all required fields (name, category, website, tags, min-os)
- [ ] `apps/trustedqsl/description` exists; first line is the short summary
- [ ] `apps/trustedqsl/install` is executable, uses `set -euo pipefail`, sources utils, installs via apt, calls `mark_installed`
- [ ] `apps/trustedqsl/uninstall` is executable, uses `set -euo pipefail`, sources utils, removes via apt, calls `mark_uninstalled`
- [ ] Both scripts pass `shellcheck` with zero errors
- [ ] `tests/test-trustedqsl-metadata` exists and passes after app files are created
- [ ] `tests/test-trustedqsl-scripts` exists and passes after app files are created
- [ ] `trustedqsl` appears in `scripts/list-apps` output

---

## References

- Canonical pattern: `apps/wsjtx/install`, `apps/wsjtx/uninstall`, `apps/wsjtx/metadata`
- Shared utilities: `scripts/utils` (`mark_installed`, `mark_uninstalled`, `info`, `success`, `warning`, `error`)
- Category registry: `data/categories` (id `tools` already defined)
- Existing tests: `tests/test-utils-helpers` (pass/fail counter pattern to follow)
- Upstream: https://sourceforge.net/projects/trustedqsl/
