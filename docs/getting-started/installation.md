# Installation

## Requirements

- Debian 11+ or Ubuntu 20.04+
- `git` and `curl` — installed automatically by the install script if missing

The GUI is a self-contained binary built with [Go + Fyne](https://fyne.io/). No additional GUI toolkit or system library is required.

## One-line install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/KW4JLB/ham-apps/main/install | bash
```

This script:

1. Checks your OS version (Debian 11+ or Ubuntu 20.04+ required)
2. Installs `git` and `curl` if they are missing
3. Clones ham-apps to `~/ham-apps`
4. Downloads the latest pre-built GUI binary from [GitHub Releases](https://github.com/KW4JLB/ham-apps/releases/latest) for your architecture (`amd64` or `arm64`)
5. Adds ham-apps to your `PATH` in `~/.bashrc` (and `~/.zshrc` if present)

### Options

```bash
# Install to a custom directory
HAMAPPS_DIR=/opt/ham-apps curl -fsSL https://raw.githubusercontent.com/KW4JLB/ham-apps/main/install | bash

# Preview what will happen without making changes
bash install --dry-run

# Help
bash install --help
```

## Manual install

Clone the repository and add it to your PATH:

```bash
git clone https://github.com/KW4JLB/ham-apps.git ~/ham-apps
echo 'export PATH="$HOME/ham-apps:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Then download the GUI binary for your architecture from the [latest release](https://github.com/KW4JLB/ham-apps/releases/latest) and place it at `~/ham-apps/gui/ham-apps-gui`:

```bash
# Example for amd64:
curl -fsSL -o ~/ham-apps/gui/ham-apps-gui \
  https://github.com/KW4JLB/ham-apps/releases/latest/download/ham-apps-gui-linux-amd64
chmod 755 ~/ham-apps/gui/ham-apps-gui
```

## Building from source

Requires Go 1.26+ and the Fyne build dependencies:

```bash
sudo apt-get install -y pkg-config libgl1-mesa-dev libx11-dev \
  libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev libxxf86vm-dev

git clone https://github.com/KW4JLB/ham-apps.git ~/ham-apps
cd ~/ham-apps/gui
make build
```

The binary is written to `~/ham-apps/gui/ham-apps-gui`.

## Verify the install

```bash
ham-apps --version
```

## Updating ham-apps

```bash
ham-apps update
```

This runs a `git pull` in the installation directory. To update the GUI binary, re-run the install script or download a new binary from the [releases page](https://github.com/KW4JLB/ham-apps/releases).

## Uninstalling

```bash
rm -rf ~/ham-apps
# remove the export line from ~/.bashrc
```

Installed app state is tracked in `~/.local/share/ham-apps/installed/`. Removing that directory clears the state but does not uninstall the apps themselves — run `ham-apps uninstall <slug>` for each app first.
