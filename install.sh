#!/bin/bash
# install.sh — Bootstrap installer for ham-apps
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/KW4JLB/ham-apps/main/install.sh | bash
#   bash install.sh [--help] [--dry-run]
#
# Environment overrides:
#   HAMAPPS_DIR   — installation directory (default: $HOME/ham-apps)
#   HAMAPPS_REPO  — git clone URL (default: https://github.com/KW4JLB/ham-apps.git)

set -euo pipefail

# ---------------------------------------------------------------------------
# Inline colour helpers (cannot source scripts/utils — this runs before clone)
# ---------------------------------------------------------------------------
info()    { echo -e "\e[34m[INFO]\e[0m  $*"; }
success() { echo -e "\e[32m[OK]\e[0m    $*"; }
warning() { echo -e "\e[33m[WARN]\e[0m  $*" >&2; }
error()   { echo -e "\e[31m[ERROR]\e[0m $*" >&2; }
die()     { error "$*"; exit 1; }

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
DRY_RUN=0

usage() {
  cat <<'EOF'
Usage: bash install.sh [OPTIONS]

Bootstrap ham-apps on a Debian 11+ or Ubuntu 20.04+ system.

Options:
  --help, -h     Show this help message and exit
  --dry-run      Print planned actions without executing them

Environment variables:
  HAMAPPS_DIR    Installation directory (default: $HOME/ham-apps)
  HAMAPPS_REPO   Git clone URL (default: https://github.com/KW4JLB/ham-apps.git)

Example:
  curl -fsSL https://raw.githubusercontent.com/KW4JLB/ham-apps/main/install.sh | bash
  HAMAPPS_DIR=/opt/ham-apps bash install.sh
EOF
}

for arg in "$@"; do
  case "$arg" in
    --help|-h)
      usage
      exit 0
      ;;
    --dry-run)
      DRY_RUN=1
      ;;
    *)
      die "Unknown option: $arg  (try --help)"
      ;;
  esac
done

# ---------------------------------------------------------------------------
# Dry-run command wrapper
# ---------------------------------------------------------------------------
run_cmd() {
  if [[ $DRY_RUN -eq 1 ]]; then
    echo "[DRY-RUN] would: $*"
  else
    "$@"
  fi
}

# ---------------------------------------------------------------------------
# OS detection
# ---------------------------------------------------------------------------
check_os() {
  local os_release="${HAMAPPS_TEST_OS_RELEASE:-/etc/os-release}"
  if [[ ! -f "$os_release" ]]; then
    die "Cannot detect OS: $os_release not found."
  fi

  local ID VERSION_ID
  # shellcheck source=/dev/null
  . "$os_release"

  case "${ID:-}" in
    debian)
      if [[ "${VERSION_ID:-0}" -lt 11 ]]; then
        die "Unsupported Debian version ${VERSION_ID}. Requires Debian 11 (bullseye) or later."
      fi
      ;;
    ubuntu)
      local major minor
      major="${VERSION_ID%%.*}"
      minor="${VERSION_ID##*.}"
      if [[ "$major" -lt 20 ]] || { [[ "$major" -eq 20 ]] && [[ "$minor" -lt 4 ]]; }; then
        die "Unsupported Ubuntu version ${VERSION_ID}. Requires Ubuntu 20.04 or later."
      fi
      ;;
    *)
      die "Unsupported OS: ${ID:-unknown}. ham-apps requires Debian 11+ or Ubuntu 20.04+."
      ;;
  esac
}

# ---------------------------------------------------------------------------
# Install directory validation (CWE-78: command injection prevention)
# ---------------------------------------------------------------------------
validate_install_dir() {
  local dir="$1"
  # Reject paths containing shell metacharacters
  if [[ "$dir" =~ [\$\(\)\`\;\|\&\<\>\"\'\\] ]]; then
    die "HAMAPPS_DIR contains invalid characters. Use a plain filesystem path."
  fi
  # Must be an absolute path or start with ~ (home-relative)
  if [[ "$dir" != /* && "$dir" != "~"* ]]; then
    die "HAMAPPS_DIR must be an absolute path (e.g. /home/user/ham-apps)."
  fi
}

# ---------------------------------------------------------------------------
# Dependency installation
# ---------------------------------------------------------------------------
install_deps() {
  local missing=()
  command -v git &>/dev/null || missing+=(git)
  command -v yad &>/dev/null || missing+=(yad)
  if [[ ${#missing[@]} -gt 0 ]]; then
    info "Installing missing packages: ${missing[*]}"
    run_cmd sudo apt-get update -qq
    run_cmd sudo apt-get install -y "${missing[@]}"
  else
    info "Dependencies already satisfied (git, yad)."
  fi
}

# ---------------------------------------------------------------------------
# Clone or skip
# ---------------------------------------------------------------------------
clone_repo() {
  local target="$1"
  local repo="${HAMAPPS_REPO:-https://github.com/KW4JLB/ham-apps.git}"
  if [[ -d "$target/.git" ]]; then
    warning "ham-apps already cloned at $target — skipping clone."
  elif [[ -d "$target" ]]; then
    die "Directory $target exists but is not a git repo. Remove it or set HAMAPPS_DIR to a different path."
  else
    info "Cloning ham-apps into $target ..."
    run_cmd git clone "$repo" "$target"
  fi
}

# ---------------------------------------------------------------------------
# PATH configuration
# ---------------------------------------------------------------------------
configure_path() {
  local dir="${HAMAPPS_DIR:-$HOME/ham-apps}"
  local export_line="export PATH=\"${dir}:\$PATH\""
  local bashrc="${BASHRC_FILE:-$HOME/.bashrc}"
  local zshrc="$HOME/.zshrc"

  if ! grep -qF "$export_line" "$bashrc" 2>/dev/null; then
    run_cmd bash -c "echo '$export_line' >> '$bashrc'"
    info "Added ham-apps to PATH in $bashrc"
  else
    info "ham-apps PATH already set in $bashrc — skipping."
  fi

  if [[ -f "$zshrc" ]]; then
    if ! grep -qF "$export_line" "$zshrc" 2>/dev/null; then
      run_cmd bash -c "echo '$export_line' >> '$zshrc'"
      info "Added ham-apps to PATH in $zshrc"
    else
      info "ham-apps PATH already set in $zshrc — skipping."
    fi
  fi
}

# ---------------------------------------------------------------------------
# GUI binary installation
# Downloads the pre-built ham-apps-gui binary from the latest GitHub release,
# or falls back to building from source if Go is available.
# ---------------------------------------------------------------------------
install_gui_binary() {
  local hamapps_dir="$1"
  local gui_dir="$hamapps_dir/gui"
  local gui_binary="$gui_dir/ham-apps-gui"

  # Detect architecture
  local arch
  case "$(uname -m)" in
    x86_64)          arch="amd64" ;;
    aarch64|arm64)   arch="arm64" ;;
    *)               arch="" ;;
  esac

  if [[ -z "$arch" ]]; then
    warning "Unsupported architecture: $(uname -m). Skipping GUI binary download."
    return 0
  fi

  # Fetch latest release tag from GitHub API
  local latest_tag
  latest_tag="$(curl -fsSL "https://api.github.com/repos/KW4JLB/ham-apps/releases/latest" \
    2>/dev/null | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

  if [[ -z "$latest_tag" ]]; then
    warning "Could not determine latest release tag from GitHub API."
  else
    local download_url="https://github.com/KW4JLB/ham-apps/releases/download/${latest_tag}/ham-apps-gui-linux-${arch}"
    info "Downloading GUI binary from $download_url ..."
    if run_cmd curl -fsSL --output "$gui_binary" "$download_url"; then
      run_cmd chmod 755 "$gui_binary"
      success "GUI binary installed at $gui_binary"
      return 0
    else
      warning "Download failed. Attempting to build from source..."
    fi
  fi

  # Fallback: build from source if Go is available
  if command -v go &>/dev/null; then
    info "Building GUI binary from source using 'make -C $gui_dir build' ..."
    if run_cmd make -C "$gui_dir" build; then
      success "GUI binary built from source at $gui_binary"
      return 0
    else
      warning "Build from source failed."
    fi
  fi

  error "ham-apps GUI binary could not be installed. See docs/getting-started/installation.md"
  exit 1
}

# ---------------------------------------------------------------------------
# Success banner
# ---------------------------------------------------------------------------
print_banner() {
  local dir="${HAMAPPS_DIR:-$HOME/ham-apps}"
  local bashrc="${BASHRC_FILE:-$HOME/.bashrc}"
  echo ""
  echo "============================================================"
  echo "  ham-apps installed successfully!"
  echo "============================================================"
  echo ""
  echo "  Installation directory : $dir"
  echo "  Added to PATH in       : $bashrc"
  echo ""
  echo "  To start using ham-apps:"
  echo "    Reload your shell:    source $bashrc"
  echo "    Launch the GUI:       ham-apps gui"
  echo "    List apps:            ham-apps list"
  echo ""
  echo "  Happy DXing! 73 de KW4JLB"
  echo "============================================================"
  echo ""
}

# ---------------------------------------------------------------------------
# Main — only runs when NOT in test mode
# ---------------------------------------------------------------------------
if [[ "${HAMAPPS_TEST_MODE:-0}" -eq 0 ]]; then
  HAMAPPS_DIR="${HAMAPPS_DIR:-$HOME/ham-apps}"
  HAMAPPS_REPO="${HAMAPPS_REPO:-https://github.com/KW4JLB/ham-apps.git}"

  if [[ $DRY_RUN -eq 1 ]]; then
    info "Dry-run mode enabled — no changes will be made."
    # Validate OS but skip if we have a test override
    os_release="${HAMAPPS_TEST_OS_RELEASE:-/etc/os-release}"
    if [[ -f "$os_release" ]]; then
      check_os
    fi
    validate_install_dir "$HAMAPPS_DIR"
    run_cmd sudo apt-get update -qq
    run_cmd sudo apt-get install -y git yad
    run_cmd git clone "${HAMAPPS_REPO}" "$HAMAPPS_DIR"
    run_cmd curl -fsSL --output "$HAMAPPS_DIR/gui/ham-apps-gui" "https://github.com/KW4JLB/ham-apps/releases/download/<tag>/ham-apps-gui-linux-<arch>"
    run_cmd bash -c "echo 'export PATH=\"${HAMAPPS_DIR}:\$PATH\"' >> ~/.bashrc"
    success "Dry-run complete. No changes made."
    exit 0
  fi

  check_os
  validate_install_dir "$HAMAPPS_DIR"
  install_deps
  clone_repo "$HAMAPPS_DIR"
  install_gui_binary "$HAMAPPS_DIR"
  configure_path
  print_banner
fi
