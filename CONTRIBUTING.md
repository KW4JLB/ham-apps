# Contributing to ham-apps

## Adding a New App

Create a directory under `apps/` named with a short lowercase slug (e.g. `apps/js8call/`).

It must contain four files:

### `metadata`
Key=value pairs:
```
name=JS8Call
category=digital-modes
website=http://js8call.com
tags=js8,digital,weak-signal
min-os=Ubuntu 20.04
```

**Required fields:** `name`, `category`, `website`

### `description`
Plain text. First line is the short summary shown in the app list. Remaining lines are the full description shown in the detail view.

### `install`
A bash script that installs the app. Guidelines:
- Start with `set -euo pipefail`
- Source `$HAMAPPS_DIR/scripts/utils` for logging helpers
- Use `sudo` only where needed
- Clean up temp files with a `trap` on EXIT
- Do not assume a specific working directory

### `uninstall`
A bash script that cleanly removes the app. Should undo everything `install` does.

## Script Style

- Use `info`, `success`, `warning`, and `error` from `scripts/utils` for output
- No hardcoded version strings without a `# TODO: dynamic version detection` comment
- Test on a clean Debian/Ubuntu install if possible

## App Icons

Optional: add an `icon.png` (64x64 or 128x128) to the app directory. If absent, a category default is used.
