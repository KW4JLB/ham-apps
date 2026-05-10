# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

`ham-apps` is an app manager for Amateur Radio software on Debian and Ubuntu, inspired by pi-apps. It provides a GUI (yad-based) and CLI for installing apps that are difficult to get through standard package managers.

## Commands

**Run all tests:**
```bash
for t in tests/test-*; do bash "$t"; done
```

**Run a single test:**
```bash
bash tests/test-utils-helpers
```

**Lint a script with shellcheck:**
```bash
shellcheck -x scripts/utils
shellcheck -x apps/wsjtx/install
```

**Lint all scripts at once:**
```bash
find . -path './.git' -prune -o -type f -print | xargs grep -lE '^#!/bin/bash' | xargs shellcheck -x
```

**Smoke-test the CLI (no GUI, no root):**
```bash
HAMAPPS_DIR="$PWD" bash ham-apps list
HAMAPPS_DIR="$PWD" bash ham-apps --version
```

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

## How tests work

Each file under `tests/` is a standalone bash script that prints `PASS:` / `FAIL:` lines and exits non-zero if any fail. Tests are written TDD-first (red phase) — they target specific functions or files that must exist. Tests use `grep` / `source` / `shellcheck -x` rather than a framework. No test runner is installed; just `bash tests/<name>` directly.

When adding a new app, add a `tests/test-<slug>-metadata` and `tests/test-<slug>-scripts` following the pattern of existing tests.

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

## Agentic workflow

**IMPORTANT CONSTRAINT**: When orchestration is invoked (e.g. "orchestrate this", "run the orchestrator", "implement this feature"), always use the project-local agent at `.claude/agents/orchestrator.md`. Never fall back to the built-in Claude orchestrator agent. This project's orchestrator is the authoritative workflow for plan → spec → review → decompose → TDD implement.
