#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-latest}"
REPO="sicario-labs/kode"

if [ "$VERSION" = "latest" ]; then
  TAG="$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)"
else
  TAG="v$VERSION"
fi

echo "Installing Kode $TAG..."

# Detect OS and arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY="kode-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
  BINARY="${BINARY}.exe"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$BINARY"
INSTALL_DIR="${KODE_INSTALL_DIR:-/usr/local/bin}"

echo "Downloading $BINARY..."
curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/kode"
chmod +x "$INSTALL_DIR/kode"

echo "Kode $TAG installed to $INSTALL_DIR/kode"

# Optional: TUI bundle
if command -v bun &>/dev/null; then
  echo "Installing TUI bundle..."
  TUI_URL="https://github.com/$REPO/releases/download/$TAG/tui-bundle.tar.gz"
  TUI_DIR="${KODE_TUI_DIR:-$HOME/.kode/tui}"
  mkdir -p "$TUI_DIR"
  curl -fsSL "$TUI_URL" | tar xz -C "$TUI_DIR"
  echo "TUI assets extracted to $TUI_DIR"
  echo "To use the TUI, set KODE_TUI_DIR=$TUI_DIR"
else
  echo "Skipping TUI bundle — bun not found. Install bun first: npm install -g bun"
fi

echo ""
echo "  Run: kode init"
echo "  Run: kode --help"
