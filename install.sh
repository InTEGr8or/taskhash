#!/usr/bin/env bash
set -e

TASKHASH_VERSION="${TASKHASH_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$PWD}"
TASKHASH_BIN="$INSTALL_DIR/taskhash"

echo "Installing taskhash v$TASKHASH_VERSION..."

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    mingw*|msys*|cygwin) OS="windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Try to get latest release info
if [ "$TASKHASH_VERSION" = "latest" ]; then
    echo "Fetching latest version..."
    TASKHASH_VERSION=$(curl -s https://api.github.com/repos/bizkite-co/taskhash/releases/latest 2>/dev/null | grep '"tag_name"' | sed 's/.*"v\?\([^"]*\)".*/\1/') || TASKHASH_VERSION="v1.0.0"
fi

TASKHASH_URL="https://github.com/bizkite-co/taskhash/releases/download/$TASKHASH_VERSION/taskhash_${OS}_${ARCH}"

if command -v curl &> /dev/null; then
    DOWNLOAD_CMD="curl -fsSL"
elif command -v wget &> /dev/null; then
    DOWNLOAD_CMD="wget -qO-"
else
    echo "Error: curl or wget required"
    exit 1
fi

# Download binary
echo "Downloading from $TASKHASH_URL..."
if [ -f "$TASKHASH_BIN" ]; then
    mv "$TASKHASH_BIN" "${TASKHASH_BIN}.old"
fi

$DOWNLOAD_CMD "$TASKHASH_URL" -o "$TASKHASH_BIN" || {
    echo "Failed to download. Building from source..."
    if ! command -v go &> /dev/null; then
        echo "Error: Go required to build from source"
        exit 1
    fi
    go build -o "$TASKHASH_BIN" taskhash.go 2>/dev/null || {
        echo "Error: Could not build. Make sure you're in the taskhash source directory."
        exit 1
    }
}

chmod +x "$TASKHASH_BIN"
rm -f "${TASKHASH_BIN}.old"

echo "taskhash installed at $TASKHASH_BIN"

# Initialize in the current repo
echo ""
echo "Initializing taskhash..."
"$TASKHASH_BIN" init

echo ""
echo "Done! Run 'taskhash --help' for usage."
