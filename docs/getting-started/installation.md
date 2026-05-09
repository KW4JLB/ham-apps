# Installation

## Requirements

- Debian 11+ or Ubuntu 20.04+
- `yad` for the GUI
- `git` for self-updates

Install the GUI dependency:

```bash
sudo apt install yad git
```

## Clone and install

```bash
git clone https://github.com/KW4JLB/ham-apps.git ~/ham-apps
```

Add ham-apps to your PATH:

```bash
echo 'export PATH="$HOME/ham-apps:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Verify the install:

```bash
ham-apps --version
```

## Updating ham-apps

ham-apps updates itself from the GitHub repository:

```bash
ham-apps update
```

## Uninstalling ham-apps

Remove the clone and the PATH entry:

```bash
rm -rf ~/ham-apps
# remove the export line from ~/.bashrc
```

Installed apps are tracked in `~/.local/share/ham-apps/installed/`. Deleting that directory will clear the install state but will not uninstall the apps themselves — run `ham-apps uninstall <slug>` for each app first.
