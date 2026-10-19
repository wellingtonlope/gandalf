#!/usr/bin/env bash
set -euo pipefail

# install.sh - Install gandalf without Go
# Usage:
#   curl -sSL https://raw.githubusercontent.com/wellingtonlope/gandalf/main/install.sh | bash
#   curl -sSL https://raw.githubusercontent.com/wellingtonlope/gandalf/main/install.sh | bash -s -- --version v0.1.0
#   ./install.sh --dir /usr/local/bin --version latest

REPO="wellingtonlope/gandalf"
BINARY="gandalf"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

usage() {
  echo "Usage: $0 [--dir <install_dir>] [--version <tag>]"
  echo "  --dir     Install directory (default: /usr/local/bin)"
  echo "  --version Tag to install (default: latest, e.g. v0.1.0)"
  echo "  Env: INSTALL_DIR, VERSION"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dir) INSTALL_DIR="$2"; shift 2 ;;
    --version) VERSION="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; usage; exit 1 ;;
  esac
done

detect_os_arch() {
  local os arch
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)
  case "$os" in
    linux) os="linux" ;;
    darwin) os="darwin" ;;
    *) echo "Unsupported OS: $os" >&2; exit 1 ;;
  esac
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) echo "Unsupported arch: $arch" >&2; exit 1 ;;
  esac
  echo "${os}-${arch}"
}

resolve_version() {
  if [[ "$VERSION" == "latest" ]]; then
    # Follow GitHub redirect to get latest tag
    local url
    url=$(curl -sI -L "https://github.com/${REPO}/releases/latest" | tr -d '\r' | grep -i "^location:" | tail -1 | awk '{print $2}')
    if [[ -z "$url" ]]; then
      echo "Could not resolve latest version" >&2
      exit 1
    fi
    basename "$url"
  else
    echo "$VERSION"
  fi
}

main() {
  local platform tag asset url checksums_url tmpdir expected actual
  platform=$(detect_os_arch)
  tag=$(resolve_version)
  asset="${BINARY}-${platform}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
  checksums_url="https://github.com/${REPO}/releases/download/${tag}/checksums.txt"

  echo "Installing ${BINARY} ${tag} for ${platform}..."
  echo "  URL: ${url}"
  echo "  Dir: ${INSTALL_DIR}"

  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  if ! curl -fsSL "$url" -o "$tmpdir/$asset"; then
    echo "Download failed: $url" >&2
    echo "Check https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  if ! curl -fsSL "$checksums_url" -o "$tmpdir/checksums.txt"; then
    echo "Failed to download checksums: $checksums_url" >&2
    exit 1
  fi

  expected=$(awk -v asset="$asset" '$2 == asset { print $1 }' "$tmpdir/checksums.txt")
  if [[ ! "$expected" =~ ^[[:xdigit:]]{64}$ ]]; then
    echo "Could not find a unique SHA-256 checksum for $asset in checksums.txt" >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$tmpdir/$asset" | awk '{ print $1 }')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$tmpdir/$asset" | awk '{ print $1 }')
  else
    echo "Neither sha256sum nor shasum is available to verify the download" >&2
    exit 1
  fi

  if [[ "$(printf '%s' "$actual" | tr '[:upper:]' '[:lower:]')" != "$(printf '%s' "$expected" | tr '[:upper:]' '[:lower:]')" ]]; then
    echo "SHA-256 mismatch for $asset" >&2
    exit 1
  fi
  echo "SHA-256 verified for $asset"

  tar -xzf "$tmpdir/$asset" -C "$tmpdir"
  local bin_path="$tmpdir/${BINARY}-${platform}"
  # Some archives contain just the binary; handle both
  if [[ ! -f "$bin_path" ]]; then
    # fallback: archive extracts directly to gandalf
    bin_path=$(find "$tmpdir" -type f -name "${BINARY}*" | head -1)
  fi
  # Single binary archives extract to gandalf-${platform}; rename to gandalf
  if [[ -f "$tmpdir/${BINARY}-${platform}" ]]; then
    mv "$tmpdir/${BINARY}-${platform}" "$tmpdir/${BINARY}"
    bin_path="$tmpdir/${BINARY}"
  fi

  chmod +x "$bin_path"

  mkdir -p "$INSTALL_DIR"
  if [[ -w "$INSTALL_DIR" ]]; then
    mv "$bin_path" "$INSTALL_DIR/${BINARY}"
  else
    echo "Need sudo to write to $INSTALL_DIR"
    sudo mv "$bin_path" "$INSTALL_DIR/${BINARY}"
  fi

  echo "Installed to $INSTALL_DIR/${BINARY}"
  "$INSTALL_DIR/${BINARY}" --version || "$INSTALL_DIR/${BINARY}" -v || true
  echo "Try: gandalf --help  (ensure $INSTALL_DIR is in your PATH)"
}

main "$@"
