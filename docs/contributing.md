# Contributing

Contributions are welcome — new apps, bug fixes, documentation improvements, and shell script cleanup.

## Adding a new app

See [Adding an App](dev/adding-an-app.md) for the full walkthrough.

The short version: create `apps/<slug>/` with `metadata`, `description`, `install`, and `uninstall`, then open a pull request.

## Reporting bugs

Open an issue on GitHub and include:

- Your OS and version (`lsb_release -a`)
- The app slug and command you ran
- The full terminal output

## Pull request guidelines

- One app or fix per PR
- `install` and `uninstall` must both be present and tested
- Scripts must pass `shellcheck`
- Keep commit messages short and descriptive

## Development environment

No special tooling is required. Clone the repo and work directly with bash. For GUI changes, install `yad`:

```bash
sudo apt install yad shellcheck
```

Run the linter against your changes:

```bash
shellcheck scripts/* apps/*/install apps/*/uninstall
```
