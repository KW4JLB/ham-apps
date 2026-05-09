# ham-apps — Claude Code Context

## What this project is

`ham-apps` is an app manager for Amateur Radio software on Debian and Ubuntu, inspired by pi-apps. It provides a GUI (yad-based) and CLI for installing apps that are difficult to get through standard package managers.

## Repository layout

```
ham-apps          # main entry point (bash)
scripts/          # core logic scripts (bash)
  utils           # shared functions — source this in all scripts
  install-app     # runs apps/<name>/install
  uninstall-app   # runs apps/<name>/uninstall
  list-apps       # tabular list of available/installed apps
  update          # git-based self-update
gui/              # yad GUI scripts
  app-list        # main browser window
  app-details     # confirm + launch install/uninstall
apps/             # one directory per app
  <slug>/
    metadata      # key=value: name, category, website, tags, min-os
    description   # plain text; first line = short summary
    install       # bash install script
    uninstall     # bash uninstall script
    icon.png      # optional 64x64 or 128x128 icon
data/
  categories      # pipe-delimited category definitions
tests/            # shellcheck and smoke tests
version           # semver string
```

## Install state

Installed apps are tracked as empty files in `~/.local/share/ham-apps/installed/<slug>`. The `is_installed`, `mark_installed`, and `mark_uninstalled` functions in `scripts/utils` manage this.

## Key conventions

- All scripts source `scripts/utils` for logging (`info`, `success`, `warning`, `error`) and helpers
- App install scripts use `set -euo pipefail` and clean up temp dirs with `trap`
- No hardcoded version strings without a `# TODO: dynamic version detection` comment
- `sudo` is used inside app scripts, not by the core runner — so scripts can be linted/tested without root
- Categories are defined in `data/categories`; app metadata references them by id

## Tech stack

- Pure bash + yad (GUI) / zenity (fallback)
- No Python, no compiled code in core
- Target: Debian 11+ and Ubuntu 20.04+
