#!/bin/bash
#
# Debafr Installer
# Installs Debafr to ~/.local/bin
#
# Usage:
#   Local:  ./install.sh [OPTIONS]
#   Remote: curl -LsSf https://.../install.sh | bash -s -- [OPTIONS]
#
# Options:
#   -v, --version VERSION  Install specific version (default: latest)
#   -n, --dry-run          Show what would be done without making changes
#   -u, --uninstall        Remove Debafr
#   -q, --quiet            Suppress non-error output
#   -h, --help             Show this help message
#
# Examples:
#   curl -LsSf https://.../install.sh | bash
#   curl -LsSf https://.../install.sh | bash -s -- -v v0.2.0
#   curl -LsSf https://.../install.sh | bash -s -- --uninstall
#

set -euo pipefail

#==============================================================================
# Configuration
#==============================================================================
DEFAULT_VERSION="v0.2.0"
HOME_BIN_DIR="${HOME}/.local/bin"
REPO_URL="https://github.com/dimns/debafr"

#==============================================================================
# Global variables
#==============================================================================
QUIET=false
DRY_RUN=false
UNINSTALL=false
VERSION=""
COLOR_RED=""
COLOR_GREEN=""
COLOR_YELLOW=""
COLOR_RESET=""

#==============================================================================
# Cleanup handler
#==============================================================================
cleanup() {
  local exit_code=$?
  if [[ -f "${TMP_DIR:-}/debafr.tar.gz" ]]; then
    rm -f "${TMP_DIR}/debafr.tar.gz" 2>/dev/null || true
  fi
  if [[ -n "${TMP_DIR:-}" && -d "${TMP_DIR:-}" ]]; then
    rm -rf "$TMP_DIR" 2>/dev/null || true
  fi
  exit $exit_code
}

trap cleanup EXIT INT TERM

#==============================================================================
# Utility functions
#==============================================================================
init_colors() {
  if [[ -t 1 ]]; then
    COLOR_RED=$'\033[31m'
    COLOR_GREEN=$'\033[32m'
    COLOR_YELLOW=$'\033[33m'
    COLOR_RESET=$'\033[0m'
  fi
}

log() {
  if [[ "$QUIET" != "true" ]]; then
    echo "$@"
  fi
}

info() {
  log "${COLOR_GREEN}INFO${COLOR_RESET}: $*"
}

warn() {
  log "${COLOR_YELLOW}WARNING${COLOR_RESET}: $*" >&2
}

err() {
  echo "${COLOR_RED}ERROR${COLOR_RESET}: $*" >&2
  exit 1
}

check_cmd() {
  command -v "$1" >/dev/null 2>&1
}

need_cmd() {
  if ! check_cmd "$1"; then
    err "required command not found: '$1'"
  fi
}

show_help() {
  cat <<HELP
Debafr Installer

Usage:
  Local:  $(basename "$0") [OPTIONS]
  Remote: curl -LsSf https://.../install.sh | bash -s -- [OPTIONS]

Options:
  -v, --version VERSION  Install specific version (default: ${DEFAULT_VERSION})
  -n, --dry-run          Show what would be done without making changes
  -u, --uninstall        Remove Debafr from the system
  -q, --quiet            Suppress non-error output
  -h, --help             Show this help message

Examples:
  $(basename "$0")                              # Install latest version
  $(basename "$0") -v v0.2.0                    # Install specific version
  $(basename "$0") --uninstall                  # Remove Debafr
  $(basename "$0") --dry-run                    # Preview installation

  curl -LsSf https://.../install.sh | bash        # Quick install latest
  curl -LsSf https://.../install.sh | bash -s -- -v v0.2.0  # Install specific version
  curl -LsSf https://.../install.sh | bash -s -- --uninstall  # Remove
HELP
}

#==============================================================================
# Argument parsing
#==============================================================================
parse_args() {
  # Support passing args via environment variable for piped installations
  # Usage: INSTALL_DEBAFR_ARGS="-v v0.2.0" curl ... | bash
  local env_args=()
  if [[ -n "${INSTALL_DEBAFR_ARGS:-}" ]]; then
    # shellcheck disable=SC2206
    env_args=($INSTALL_DEBAFR_ARGS)
  fi

  # Combine env args with positional args (prefer positional)
  local all_args=()
  if [[ ${#env_args[@]} -gt 0 && $# -eq 0 ]]; then
    all_args=("${env_args[@]}")
  else
    all_args=("$@")
  fi

  while [[ ${#all_args[@]} -gt 0 ]]; do
    case "${all_args[0]}" in
      -v|--version)
        VERSION="${all_args[1]:-}"
        if [[ -z "$VERSION" ]]; then
          err "option '$1' requires a version argument"
        fi
        all_args=("${all_args[@]:2}")
        ;;
      -n|--dry-run)
        DRY_RUN=true
        all_args=("${all_args[@]:1}")
        ;;
      -u|--uninstall)
        UNINSTALL=true
        all_args=("${all_args[@]:1}")
        ;;
      -q|--quiet)
        QUIET=true
        all_args=("${all_args[@]:1}")
        ;;
      -h|--help)
        show_help
        exit 0
        ;;
      -*)
        err "unknown option: '${all_args[0]}'"
        ;;
      *)
        err "unexpected argument: '${all_args[0]}'"
        ;;
    esac
  done

  # Set default version if not specified
  if [[ -z "$VERSION" && "$UNINSTALL" != "true" ]]; then
    VERSION="$DEFAULT_VERSION"
  fi
}

#==============================================================================
# Platform detection
#==============================================================================
detect_platform() {
  local os arch

  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    linux*)  os="linux" ;;
    darwin*) os="macos" ;;
    *)       err "unsupported OS: $(uname -s)" ;;
  esac

  case "$(uname -m)" in
    x86_64)          arch="amd64" ;;
    aarch64|arm64)   arch="arm64" ;;
    armv7l|armhf)    arch="arm" ;;
    *)               err "unsupported architecture: $(uname -m)" ;;
  esac

  echo "${os}_${arch}"
}

#==============================================================================
# Pre-flight checks
#==============================================================================
preflight_checks() {
  # Check for required commands
  need_cmd curl
  need_cmd tar
  need_cmd chmod
  need_cmd rm
  need_cmd mkdir

  # Check if running as root
  if [[ "$(id -u)" -eq 0 ]]; then
    warn "Running as root is not recommended. Installation target is user home directory."
  fi

  # Check write access to HOME_BIN_DIR
  if [[ -d "$HOME_BIN_DIR" && ! -w "$HOME_BIN_DIR" ]]; then
    err "no write permission to $HOME_BIN_DIR"
  fi
}

#==============================================================================
# PATH check
#==============================================================================
check_path() {
  case ":${PATH}:" in
    *:"$HOME_BIN_DIR":*)
      return 0
      ;;
  esac

  warn "$HOME_BIN_DIR is not in your PATH"
  log ""
  log "To use Debafr, add the following to your shell profile:"
  log "  export PATH=\"$HOME_BIN_DIR:\$PATH\""
  log ""
}

#==============================================================================
# Download functions
#==============================================================================
get_download_url() {
  local platform="$1"
  local version="$2"
  local version_clean="${version#v}"

  echo "${REPO_URL}/releases/download/${version}/debafr_${version_clean}_${platform}.tar.gz"
}

download_file() {
  local url="$1"
  local output="$2"

  if [[ "$DRY_RUN" == "true" ]]; then
    info "[DRY-RUN] Would download: $url"
    return 0
  fi

  if ! curl -sLf -o "$output" "$url"; then
    err "failed to download from $url"
  fi
}

#==============================================================================
# Verify checksum (if checksum file exists)
#==============================================================================
verify_checksum() {
  local tarball="$1"
  local version="$2"
  local platform="$3"
  local checksum_url="${REPO_URL}/releases/download/${version}/debafr_${version#v}_checksums.txt"
  local expected_checksum
  local actual_checksum

  if [[ "$DRY_RUN" == "true" ]]; then
    info "[DRY-RUN] Would verify checksum"
    return 0
  fi

  # Try to fetch checksum file (optional)
  if expected_checksum=$(curl -sLf "$checksum_url" 2>/dev/null | grep "${platform}.tar.gz" | awk '{print $1}'); then
    actual_checksum=$(sha256sum "$tarball" 2>/dev/null | awk '{print $1}') || \
    actual_checksum=$(shasum -a 256 "$tarball" 2>/dev/null | awk '{print $1}') || \
    err "no checksum tool available (need sha256sum or shasum)"

    if [[ "$expected_checksum" != "$actual_checksum" ]]; then
      err "checksum mismatch! file may be corrupted or tampered with"
    fi

    info "Checksum verified ✓"
  else
    warn "Checksum file not found, skipping verification"
  fi
}

#==============================================================================
# Installation
#==============================================================================
install_debafr() {
  local platform url tarball_dir

  platform=$(detect_platform)
  url=$(get_download_url "$platform" "$VERSION")

  log "Installing Debafr ${VERSION} for ${platform}..."
  log ""

  # Create temporary directory
  if [[ "$DRY_RUN" != "true" ]]; then
    TMP_DIR=$(mktemp -d)
    tarball_dir="$TMP_DIR"
  else
    tarball_dir="/tmp"
    info "[DRY-RUN] Would create temporary directory"
  fi

  local tarball="${tarball_dir}/debafr.tar.gz"

  # Download
  info "Downloading..."
  download_file "$url" "$tarball"

  if [[ "$DRY_RUN" == "true" ]]; then
    info "[DRY-RUN] Would verify checksum"
    info "[DRY-RUN] Would extract to $HOME_BIN_DIR"
    info "[DRY-RUN] Would set execute permissions"
    return 0
  fi

  # Verify checksum
  verify_checksum "$tarball" "$VERSION" "$platform"

  # Create target directory
  if ! mkdir -p "$HOME_BIN_DIR"; then
    err "unable to create directory: $HOME_BIN_DIR"
  fi

  # Extract
  info "Extracting..."
  if ! tar -xzf "$tarball" -C "$HOME_BIN_DIR" debafr; then
    # Try without specific path (some archives have different structure)
    if ! tar -xzf "$tarball" -C "$HOME_BIN_DIR"; then
      err "failed to extract archive"
    fi
  fi

  # Set permissions
  info "Setting permissions..."
  if ! chmod +x "${HOME_BIN_DIR}/debafr"; then
    err "failed to set execute permissions"
  fi

  # Verify installation
  if ! [[ -x "${HOME_BIN_DIR}/debafr" ]]; then
    err "installation verification failed: binary not executable"
  fi

  log ""
  info "Debafr ${VERSION} successfully installed!"
  log ""
  log "Usage: debafr"
  log ""

  # Check PATH
  check_path
}

#==============================================================================
# Uninstallation
#==============================================================================
uninstall_debafr() {
  local binary="${HOME_BIN_DIR}/debafr"

  log "Uninstalling Debafr..."
  log ""

  if [[ ! -f "$binary" ]]; then
    warn "Debafr not found at $binary"
    return 0
  fi

  if [[ "$DRY_RUN" == "true" ]]; then
    info "[DRY-RUN] Would remove: $binary"
    return 0
  fi

  if ! rm -f "$binary"; then
    err "failed to remove $binary"
  fi

  info "Debafr removed successfully!"
  log ""
  log "You may also want to remove $HOME_BIN_DIR from PATH if no longer needed."
}

#==============================================================================
# Main
#==============================================================================
main() {
  init_colors
  parse_args "$@"

  if [[ "$UNINSTALL" == "true" ]]; then
    preflight_checks
    uninstall_debafr
  else
    preflight_checks
    install_debafr
  fi
}

main "$@"
