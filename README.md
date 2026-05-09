# ham-apps

An app manager for Amateur Radio software on Debian and Ubuntu.

Inspired by [pi-apps](https://github.com/Botspot/pi-apps), ham-apps makes it easy to discover, install, and remove amateur radio applications that are hard to get through standard package managers — things like WSJT-X (latest release), Direwolf (built from source), Winlink clients, and more.

## Requirements

- Debian 11+ or Ubuntu 20.04+
- `yad` for the GUI (`sudo apt install yad`)
- `git` for self-updates

## Install

```bash
git clone <repo-url> ~/ham-apps
echo 'export PATH="$HOME/ham-apps:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Usage

```bash
ham-apps                   # launch GUI browser
ham-apps list              # list all apps and status
ham-apps list installed    # list only installed apps
ham-apps install wsjtx     # install an app
ham-apps uninstall wsjtx   # remove an app
ham-apps update            # update ham-apps itself
```

## App Categories

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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Adding a new app means creating a directory under `apps/` with four files: `install`, `uninstall`, `description`, and `metadata`.

## License

MIT
