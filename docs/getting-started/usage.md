# Usage

## GUI

Run `ham-apps` with no arguments to open the graphical browser. It requires `yad`. If `yad` is not available, `zenity` is used as a fallback.

```bash
ham-apps
```

From the GUI you can browse apps by category, read descriptions, and launch installs or uninstalls.

## CLI

| Command | Description |
|---------|-------------|
| `ham-apps list` | List all available apps and their install status |
| `ham-apps list installed` | List only installed apps |
| `ham-apps install <slug>` | Install an app |
| `ham-apps uninstall <slug>` | Remove an app |
| `ham-apps update` | Update ham-apps itself from GitHub |

### Examples

```bash
# See what's available
ham-apps list

# Install WSJT-X
ham-apps install wsjtx

# Install Direwolf
ham-apps install direwolf

# Remove Fldigi
ham-apps uninstall fldigi
```

## Install state

Installed apps are tracked as empty marker files in `~/.local/share/ham-apps/installed/<slug>`. You can inspect this directory directly:

```bash
ls ~/.local/share/ham-apps/installed/
```
