#!/usr/bin/env bash
set -euo pipefail

REPO="kacheo/devlog"
BINARY="devlog"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux"  ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Fetch latest release version from GitHub API
if [ -z "${VERSION:-}" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest release version." >&2
  exit 1
fi

ARCHIVE="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
CHECKSUM_URL="${BASE_URL}/checksums.txt"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading devlog ${VERSION} for ${OS}/${ARCH}..."
curl -fsSL "$ARCHIVE_URL" -o "${TMP}/${ARCHIVE}"
curl -fsSL "$CHECKSUM_URL" -o "${TMP}/checksums.txt"

# Verify checksum
cd "$TMP"
if command -v sha256sum >/dev/null 2>&1; then
  grep "$ARCHIVE" checksums.txt | sha256sum --check --status
elif command -v shasum >/dev/null 2>&1; then
  grep "$ARCHIVE" checksums.txt | shasum -a 256 --check --status
else
  echo "Warning: no sha256sum or shasum found — skipping checksum verification." >&2
fi

tar -xzf "$ARCHIVE"

# Install
if [ -w "$INSTALL_DIR" ]; then
  cp "$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo cp "$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo "devlog ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo "Run 'devlog --version' to verify."
