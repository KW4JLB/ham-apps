# ham-apps

An app manager for Amateur Radio software on Debian and Ubuntu.

Inspired by [pi-apps](https://github.com/Botspot/pi-apps), ham-apps makes it easy to discover, install, and remove amateur radio applications that are hard to get through standard package managers — things like WSJT-X (latest release), Direwolf (built from source), Winlink clients, and more.

## Why ham-apps?

Most amateur radio software isn't in the official Debian/Ubuntu repositories. What is there is often out of date. ham-apps:

- Installs software at current upstream release versions
- Handles build-from-source apps automatically
- Tracks what's installed so you can cleanly uninstall
- Ships a `yad`-based GUI for users who prefer point-and-click

## Quick start

```bash
git clone https://github.com/KW4JLB/ham-apps.git ~/ham-apps
echo 'export PATH="$HOME/ham-apps:$PATH"' >> ~/.bashrc
source ~/.bashrc
ham-apps
```

See [Installation](getting-started/installation.md) for full details.

## App categories

| Category | Description |
|----------|-------------|
| digital-modes | FT8, PSK31, RTTY, and soundcard modes |
| logging | Contest and general logging |
| packet-aprs | AX.25 packet and APRS |
| satellite | Tracking and satellite SDR |
| sdr | Software-defined radio |
| tools | Rig control, antenna analysis |
| contest | Contest-focused logging |
| mapping | APRS maps and geographic tools |
| winlink | Radio email via Winlink |
