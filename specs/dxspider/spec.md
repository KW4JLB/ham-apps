# DXSpider App Entry for ham-apps

## Overview

### Purpose
Add DXSpider as an installable app in the ham-apps manager. DXSpider is a full DX cluster node written in Perl that connects to the worldwide DX spotting network. This spec covers all files required to integrate DXSpider as a first-class ham-apps app entry: metadata, description, install/uninstall scripts, a new `dx-cluster` category, and associated tests.

### Scope
- New app slug: `dxspider`
- New category: `dx-cluster` added to `data/categories`
- Files created: `apps/dxspider/metadata`, `apps/dxspider/description`, `apps/dxspider/install`, `apps/dxspider/uninstall`
- Tests created: `tests/test-dxspider-metadata`, `tests/test-dxspider-scripts`
- No GUI changes, no changes to core scripts, no changes to existing apps

### Background
DXSpider is not available through standard apt repositories. It is installed by cloning from the GitHub mirror at `https://github.com/f1evm/dxspider`, requires a dedicated `sysop` system user, Perl CPAN module dependencies installable via apt, a configuration file, and a systemd service unit. The install intentionally enables but does not start the service, as the operator must configure the cluster call sign and settings first.

---

## Requirements

### Functional Requirements

| ID   | Requirement | Priority |
|------|-------------|----------|
| FR-1 | `data/categories` contains a new line: `dx-cluster\|DX Cluster\|DX cluster node and spotting network software` | Critical |
| FR-2 | `apps/dxspider/metadata` exists with correct key=value fields: name, category, website, tags, min-os | Critical |
| FR-3 | `apps/dxspider/description` exists; first line is a short summary ≤120 chars | Critical |
| FR-4 | `apps/dxspider/install` installs Perl dependencies via apt | Critical |
| FR-5 | `apps/dxspider/install` creates the `sysop` system user if it does not already exist | Critical |
| FR-6 | `apps/dxspider/install` clones or updates the DXSpider repo to `/home/sysop/spider` | Critical |
| FR-7 | `apps/dxspider/install` installs a systemd service unit at `/etc/systemd/system/dxspider.service` | Critical |
| FR-8 | `apps/dxspider/install` enables the service (does NOT start it) | Critical |
| FR-9 | `apps/dxspider/install` calls `mark_installed dxspider` | Critical |
| FR-10 | `apps/dxspider/uninstall` stops and disables the service if it exists | Critical |
| FR-11 | `apps/dxspider/uninstall` removes `/etc/systemd/system/dxspider.service` | Critical |
| FR-12 | `apps/dxspider/uninstall` removes `/home/sysop/spider` with a warning about data loss | Critical |
| FR-13 | `apps/dxspider/uninstall` emits a `warning` log line advising operator to manually remove the `sysop` user if desired | High |
| FR-14 | `apps/dxspider/uninstall` removes installed Perl apt packages if safe to do so | High |
| FR-15 | `apps/dxspider/uninstall` calls `mark_uninstalled dxspider` | Critical |

### Non-Functional Requirements

| ID    | Requirement | Target |
|-------|-------------|--------|
| NFR-1 | Both scripts pass `shellcheck -x` with zero errors | Zero warnings/errors |
| NFR-2 | Both scripts use `set -euo pipefail` | Present in every script |
| NFR-3 | Both scripts source `$HAMAPPS_DIR/scripts/utils` for logging | Sourced at top of each script |
| NFR-4 | Install script uses `trap` for temp dir cleanup when downloading/cloning | Present |
| NFR-5 | Scripts use `sudo` internally; never run as root at the top level | No `require_root` usage |
| NFR-6 | No hardcoded version strings without a `# TODO: dynamic version detection` comment | All versions commented |
| NFR-7 | Tests follow exactly the pattern of `tests/test-hamrs-metadata` and `tests/test-hamrs-scripts` | Pattern matched |
| NFR-8 | min-os set to `Debian 11, Ubuntu 20.04` | Exact string |

### Constraints
- DXSpider is not in apt; must be installed via git clone
- The `sysop` user creation must be idempotent (skip if already exists)
- The systemd service must be enabled but NOT started (operator must configure call sign first)
- The uninstall must not unconditionally delete the `sysop` user (data safety: operator may have other uses for it)
- All test files must be standalone bash scripts that print `PASS:` / `FAIL:` lines
- No icon.png is required (optional per project layout)
- Perl package names must be verified against Debian 11 / Ubuntu 20.04 apt availability

---

## Design

### Architecture

This is a pure data-and-script addition following the existing ham-apps app entry pattern. No core scripts are modified.

```
data/
  categories           ← append dx-cluster line
apps/
  dxspider/
    metadata           ← new file
    description        ← new file
    install            ← new executable bash script
    uninstall          ← new executable bash script
tests/
  test-dxspider-metadata   ← new standalone test script
  test-dxspider-scripts    ← new standalone test script
```

### Install Script Design

```
1. set -euo pipefail
2. Resolve HAMAPPS_DIR from script path
3. source $HAMAPPS_DIR/scripts/utils
4. Install Perl dependencies via apt:
   - perl
   - libnet-telnet-perl
   - libdigest-sha-perl
   - libdata-dumper-simple-perl  (Debian/Ubuntu availability note below)
5. Create sysop system user if not exists:
   id -u sysop &>/dev/null || sudo useradd -r -m -s /bin/bash sysop
6. Clone or update repo to /home/sysop/spider:
   - If /home/sysop/spider/.git exists: sudo -u sysop git -C /home/sysop/spider pull
   - Else: sudo -u sysop git clone https://github.com/f1evm/dxspider /home/sysop/spider
   - Use trap for cleanup of partial clone on failure
   - After clone: warning "DXSpider cloned from https://github.com/f1evm/dxspider — verify the remote and review code before starting the service."
7. Write systemd service unit to /etc/systemd/system/dxspider.service
8. sudo systemctl daemon-reload
9. sudo systemctl enable dxspider
10. warning: "DXSpider installed but NOT started. Configure /home/sysop/spider/local/cluster.cfg before starting."
11. mark_installed dxspider
12. success "DXSpider installed."
```

Note on `libdata-dumper-simple-perl`: this package may not be available on all target distros. The install script should attempt it with `apt-get install -y --no-install-recommends` and fall back gracefully, or use the core `Data::Dumper` (which ships with Perl). The spec uses the set: `perl libnet-telnet-perl libdigest-sha-perl` as the minimum reliable set; `libdata-dumper-simple-perl` is attempted separately with a warning on failure.

Note on `sysop` system user shell: The `sysop` account is created with `-s /bin/bash` because DXSpider's Perl scripts may invoke shell commands during cluster operation. This is intentional and follows DXSpider's official setup documentation. The account has no password set, so direct login is not possible without sudo. Operators are advised that the `sysop` account has a full bash shell environment.

### Uninstall Script Design

```
1. set -euo pipefail
2. Resolve HAMAPPS_DIR
3. source utils
4. Stop + disable service if active:
   sudo systemctl is-active dxspider && sudo systemctl stop dxspider || true
   sudo systemctl is-enabled dxspider && sudo systemctl disable dxspider || true
5. Remove service unit and reload:
   sudo rm -f /etc/systemd/system/dxspider.service
   sudo systemctl daemon-reload
6. warning "Removing /home/sysop/spider — this deletes all cluster data."
7. sudo rm -rf /home/sysop/spider
8. warning "The 'sysop' system user was NOT removed. Remove manually with: sudo userdel -r sysop"
9. Remove DXSpider-specific Perl packages only (do NOT remove `perl` — it is a core system dependency):
   sudo apt-get remove -y libnet-telnet-perl libdigest-sha-perl || true
   (use || true so partial removal does not abort; never remove perl itself)
10. mark_uninstalled dxspider
11. success "DXSpider removed."
```

### Systemd Service Unit Content

```ini
[Unit]
Description=DXSpider DX Cluster Node
After=network.target

[Service]
Type=simple
User=sysop
WorkingDirectory=/home/sysop/spider
ExecStart=/usr/bin/perl /home/sysop/spider/cluster.pl
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Metadata File Content

```
name=DXSpider
category=dx-cluster
website=https://www.dxcluster.org/
tags=dx-cluster,dx-spots,packet,cluster,node,perl
min-os=Debian 11, Ubuntu 20.04
```

### Description File Content

```
Open-source DX cluster node software that connects to the worldwide DX spotting network.

DXSpider is a full-featured DX cluster node written in Perl. It allows amateur radio
operators to post and receive DX spots, connect to the worldwide cluster network, and
run a full cluster node serving other stations. Requires manual configuration of the
cluster call sign in /home/sysop/spider/local/cluster.cfg before starting the service.
```

---

## Test Specification

### Unit Tests — test-dxspider-metadata

**Pattern**: follows `tests/test-hamrs-metadata` exactly.

| # | Given | When | Then |
|---|-------|------|------|
| 1 | repo exists | check metadata file | `apps/dxspider/metadata` exists |
| 2 | metadata exists | grep for name | contains `name=DXSpider` |
| 3 | metadata exists | grep for category | contains `category=dx-cluster` |
| 4 | data/categories exists | grep for category id | `dx-cluster\|` line present |
| 5 | metadata exists | grep for website | contains `website=https://www.dxcluster.org/` |
| 6 | metadata exists | grep for tags | `tags=` with at least one tag |
| 7 | metadata exists | grep for min-os | `min-os=` field present |
| 8 | repo exists | check description file | `apps/dxspider/description` exists |
| 9 | description exists | read first line | non-empty, ≤120 chars |
| 10 | description exists | grep for DX cluster | mentions "DX cluster" or "DXSpider" |
| 11 | repo exists | check install | exists and is executable |
| 12 | repo exists | check uninstall | exists and is executable |

### Integration Tests — test-dxspider-scripts

**Pattern**: follows `tests/test-hamrs-scripts` exactly.

**Install script tests**:

| # | Given | When | Then |
|---|-------|------|------|
| 1 | install exists | grep | contains `set -euo pipefail` |
| 2 | install exists | grep | sources `scripts/utils` |
| 3 | install exists | grep | contains `trap` for temp/clone cleanup |
| 4 | install exists | grep | references `https://github.com/f1evm/dxspider` clone URL |
| 5 | install exists | grep | references `/home/sysop/spider` install location |
| 6 | install exists | grep | creates `sysop` user with `useradd` |
| 7 | install exists | grep | references `/etc/systemd/system/dxspider.service` |
| 8 | install exists | grep | calls `systemctl enable dxspider` |
| 9 | install exists | grep | calls `mark_installed dxspider` |
| 10 | install exists | grep | does NOT contain `systemctl start dxspider` (must not autostart) |
| 11 | shellcheck installed | shellcheck -x install | zero errors |

**Uninstall script tests**:

| # | Given | When | Then |
|---|-------|------|------|
| 12 | uninstall exists | grep | contains `set -euo pipefail` |
| 13 | uninstall exists | grep | sources `scripts/utils` |
| 14 | uninstall exists | grep | references `/home/sysop/spider` removal |
| 15 | uninstall exists | grep | references `/etc/systemd/system/dxspider.service` removal |
| 16 | uninstall exists | grep | calls `systemctl stop` and `systemctl disable` |
| 17 | uninstall exists | grep | contains `warning` about manual sysop user removal |
| 18 | uninstall exists | grep | calls `mark_uninstalled dxspider` |
| 19 | shellcheck installed | shellcheck -x uninstall | zero errors |

### Acceptance Tests

Given the full app entry is installed in the repo:
- `bash tests/test-dxspider-metadata` exits 0 with all PASS lines
- `bash tests/test-dxspider-scripts` exits 0 with all PASS lines
- `shellcheck -x apps/dxspider/install` produces zero output
- `shellcheck -x apps/dxspider/uninstall` produces zero output
- `grep 'dx-cluster|' data/categories` finds the line
- `HAMAPPS_DIR=$PWD bash ham-apps list` shows dxspider without errors

---

## Security & Compliance

### Threat Model

| Threat | Impact | Mitigation |
|--------|--------|------------|
| Git clone from untrusted mirror | Code injection during install | Use known-good mirror `https://github.com/f1evm/dxspider`; emit `warning` log line after clone advising operator to verify remote and review code before starting service |
| `sysop` user with `/bin/bash` shell | Wider post-exploitation surface if cluster Perl process is compromised | User is system user (`-r`); no password set; no sudo rights granted; `/bin/bash` required by DXSpider for shell invocations during cluster operation |
| Perl modules installed globally via apt | Supply chain risk | Use only distro-provided apt packages, not CPAN |
| `sudo rm -rf /home/sysop/spider` in uninstall | Data loss | Warn operator before removal via `warning` log line |
| Service runs as `sysop` | Isolation boundary | `User=sysop` in systemd unit; limited permissions |
| Removing `perl` in uninstall | Breaks system package manager | Do NOT remove `perl` package; only remove DXSpider-specific modules |

### Security Controls
- No credentials, tokens, or passwords in any script
- No CPAN/pip installs — apt packages only for Perl dependencies
- `sysop` user created as system account (no login shell password)
- `set -euo pipefail` prevents silent failure and unbound variable use
- `trap` ensures no partial clone artifacts remain on failure
- Service unit uses `Restart=on-failure` only (not `always`) to prevent loops on misconfiguration

### Compliance Requirements
- All scripts must pass `shellcheck -x` (project-enforced linting)
- No hardcoded credentials (project convention)
- Temp dirs cleaned via `trap` (project convention)

---

## Implementation Plan

### Phases

**Phase 1 — Test Files (TDD red phase)**
1.1 Write `tests/test-dxspider-metadata` (all tests fail — files don't exist yet)
1.2 Write `tests/test-dxspider-scripts` (all tests fail — scripts don't exist yet)

**Phase 2 — Static Files**
2.1 Add `dx-cluster` line to `data/categories`
2.2 Create `apps/dxspider/metadata`
2.3 Create `apps/dxspider/description`

**Phase 3 — Scripts**
3.1 Create `apps/dxspider/install` (executable)
3.2 Create `apps/dxspider/uninstall` (executable)

**Phase 4 — Validation**
4.1 Run `bash tests/test-dxspider-metadata` — all pass
4.2 Run `bash tests/test-dxspider-scripts` — all pass
4.3 Run `shellcheck -x apps/dxspider/install` — zero errors
4.4 Run `shellcheck -x apps/dxspider/uninstall` — zero errors

### Configuration

No environment configuration is required. The install script is self-contained. The systemd unit references `/home/sysop/spider/cluster.pl` (the DXSpider main entry point per the upstream repo) and runs as the `sysop` user.

---

## Deployment

### Deployment Steps
1. Merge PR to main branch
2. Existing users run `ham-apps update` (git pull)
3. Users run `ham-apps install dxspider`
4. Operator configures `/home/sysop/spider/local/cluster.cfg` with their call sign
5. Operator starts service: `sudo systemctl start dxspider`

### Rollback Plan
- `ham-apps uninstall dxspider` reverses all changes except the `sysop` user (per design)
- `data/categories` change is additive; removing the line restores prior state

### Monitoring
- `systemctl status dxspider` for service health
- `journalctl -u dxspider -f` for logs

---

## Acceptance Criteria

- [ ] `data/categories` contains line `dx-cluster|DX Cluster|DX cluster node and spotting network software`
- [ ] `apps/dxspider/metadata` exists with name=DXSpider, category=dx-cluster, website=https://www.dxcluster.org/, tags (≥1), min-os=Debian 11, Ubuntu 20.04
- [ ] `apps/dxspider/description` exists; first line ≤120 chars; mentions DX cluster
- [ ] `apps/dxspider/install` is executable and passes shellcheck -x with zero errors
- [ ] `apps/dxspider/install` installs Perl deps via apt, creates sysop user, clones to /home/sysop/spider, writes systemd unit, enables (not starts) service, calls mark_installed dxspider
- [ ] `apps/dxspider/uninstall` is executable and passes shellcheck -x with zero errors
- [ ] `apps/dxspider/uninstall` stops+disables service, removes unit file, removes /home/sysop/spider with warning, warns about sysop user, calls mark_uninstalled dxspider
- [ ] `bash tests/test-dxspider-metadata` exits 0, all lines PASS
- [ ] `bash tests/test-dxspider-scripts` exits 0, all lines PASS
- [ ] No modifications to any existing app entries or core scripts

---

## References

- Existing pattern: `/home/parallels/git/kw4jlb/ham-apps/apps/hamrs/install` (git clone pattern, trap, mark_installed)
- Existing pattern: `/home/parallels/git/kw4jlb/ham-apps/apps/wsjtx/install` (simple apt install pattern)
- Test pattern: `/home/parallels/git/kw4jlb/ham-apps/tests/test-hamrs-metadata`
- Test pattern: `/home/parallels/git/kw4jlb/ham-apps/tests/test-hamrs-scripts`
- Shared utils: `/home/parallels/git/kw4jlb/ham-apps/scripts/utils`
- Categories file: `/home/parallels/git/kw4jlb/ham-apps/data/categories`
- DXSpider upstream: https://github.com/f1evm/dxspider
- DXSpider official site: https://www.dxcluster.org/
