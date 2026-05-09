# Adding an App

Create a directory under `apps/` using a short lowercase slug (e.g. `apps/js8call/`). It must contain four files.

## `metadata`

Key=value pairs, one per line:

```
name=JS8Call
category=digital-modes
website=http://js8call.com
tags=js8,digital,weak-signal
min-os=Ubuntu 20.04
```

**Required:** `name`, `category`, `website`  
**Optional:** `tags`, `min-os`

Valid categories are defined in `data/categories`.

## `description`

Plain text. The **first line** is the short summary shown in the app list. Remaining lines form the full description shown in the detail view.

```
JS8Call — keyboard-to-keyboard HF digital messaging based on FT8
JS8Call extends the WSJT-X FT8 modem to support free-form messaging,
store-and-forward relaying, and APRS gateway capability over HF.
```

## `install`

A bash script that installs the app. Required structure:

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$HAMAPPS_DIR/scripts/utils"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

info "Installing JS8Call..."

# download, build, or install steps here

success "JS8Call installed."
```

Guidelines:

- Use `info`, `success`, `warning`, `error` from `scripts/utils` for all output
- Clean up temp directories with a `trap` on EXIT
- Use `sudo` only for the specific steps that need it
- Do not hardcode version strings without a `# TODO: dynamic version detection` comment
- Do not assume a specific working directory

## `uninstall`

A bash script that cleanly removes everything `install` created. Use the same structure and helpers.

## `icon.png` (optional)

A 64×64 or 128×128 PNG icon. If absent, a category default icon is used.

## Checklist before submitting

- [ ] `metadata` has `name`, `category`, and `website`
- [ ] `install` uses `set -euo pipefail` and cleans up temp files
- [ ] `uninstall` reverses the install
- [ ] Tested on a clean Debian or Ubuntu install
- [ ] No hardcoded version strings (or each has a TODO comment)
