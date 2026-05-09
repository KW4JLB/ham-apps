# Script Style Guide

All scripts in ham-apps — core and app — follow these conventions.

## Shell settings

Every script starts with:

```bash
#!/usr/bin/env bash
set -euo pipefail
```

- `-e` — exit on error
- `-u` — treat unset variables as errors
- `-o pipefail` — propagate pipe failures

## Sourcing utils

Core scripts and app install/uninstall scripts source `scripts/utils` for shared helpers:

```bash
source "$HAMAPPS_DIR/scripts/utils"
```

## Logging helpers

Use these functions instead of `echo`:

| Function | When to use |
|----------|-------------|
| `info "…"` | Normal progress messages |
| `success "…"` | Confirm a step completed |
| `warning "…"` | Non-fatal issues the user should know |
| `error "…"` | Fatal errors (script will exit) |

## Temp directory pattern

```bash
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT
```

Always use this pattern when downloading or extracting files. The `trap` fires on exit regardless of success or failure.

## sudo usage

- App scripts use `sudo` only for the specific commands that need it (e.g. `apt install`, copying to `/usr/local/bin`)
- The core runner (`ham-apps`) never runs with sudo; it is up to each app script

## Version strings

If a version is hardcoded, add a comment:

```bash
VERSION="2.6.7"  # TODO: dynamic version detection
```

This flags it for review when the upstream releases a new version.

## shellcheck

All scripts must pass `shellcheck` with no errors. Run:

```bash
shellcheck scripts/utils scripts/install-app scripts/uninstall-app
shellcheck apps/*/install apps/*/uninstall
```
