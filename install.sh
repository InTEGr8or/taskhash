#!/usr/bin/env bash
set -e

REPO="${REPO:-InTEGr8or/taskhash}"
TASKHASH_VERSION="${TASKHASH_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$PWD}"
TASKHASH_BIN="$INSTALL_DIR/taskhash"

echo "Installing taskhash v$TASKHASH_VERSION to $INSTALL_DIR..."

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
    TASKHASH_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | grep '"tag_name"' | sed 's/.*"v\?\([^"]*\)".*/\1/') || TASKHASH_VERSION="v1.0.0"
fi

TASKHASH_URL="https://github.com/$REPO/releases/download/$TASKHASH_VERSION/taskhash_${OS}_${ARCH}"

download_binary() {
    echo "Downloading from $TASKHASH_URL..."
    if command -v curl &> /dev/null; then
        curl -fsSL "$TASKHASH_URL" -o "$TASKHASH_BIN"
    elif command -v wget &> /dev/null; then
        wget -q "$TASKHASH_URL" -O "$TASKHASH_BIN"
    else
        echo "Error: curl or wget required"
        return 1
    fi
}

build_from_source() {
    echo "Building from source..."
    local tmpdir=$(mktemp -d)
    git clone --depth 1 "https://github.com/$REPO.git" "$tmpdir/taskhash" 2>/dev/null || {
        echo "Error: git required to build from source"
        rm -rf "$tmpdir"
        return 1
    }
    (cd "$tmpdir/taskhash" && go build -o "$TASKHASH_BIN" .)
    rm -rf "$tmpdir"
}

# Backup existing
if [ -f "$TASKHASH_BIN" ]; then
    cp "$TASKHASH_BIN" "${TASKHASH_BIN}.old"
fi

# Try download first, fall back to source
download_binary || build_from_source

chmod +x "$TASKHASH_BIN"
rm -f "${TASKHASH_BIN}.old"

echo "taskhash installed at $TASKHASH_BIN"

# Initialize in the current repo
echo ""
echo "Initializing taskhash..."
"$TASKHASH_BIN" init

echo ""
echo "Done! Run './taskhash --help' for usage."
