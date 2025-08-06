#!/bin/sh
set -e

# Default installation directory
DEFAULT_INSTALL_DIR="/usr/local/bin"
GITHUB_ORG="data-catering"
REPO="insta-infra"
BINARY_NAME="insta"

# Detect OS and architecture
detect_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Convert architecture names
    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64)   ARCH="arm64" ;;
    esac
    
    # Convert OS names
    case "$OS" in
        darwin)  OS="darwin" ;;
        linux)   OS="linux" ;;
        msys*|mingw*|cygwin*|windows*) 
            OS="windows"
            BINARY_NAME="insta.exe"
            ;;
    esac
}

# Function to get the latest release URL
get_download_url() {
    ARTIFACT_NAME="insta-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        ARTIFACT_NAME="insta-windows-${ARCH}.exe"
    fi
    echo "https://github.com/${GITHUB_ORG}/${REPO}/releases/latest/download/${ARTIFACT_NAME}"
}

# Create temporary directory
TMP_DIR=$(mktemp -d)
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

# Main installation
echo "Detecting system information..."
detect_os_arch

echo "Operating System: $OS"
echo "Architecture: $ARCH"

DOWNLOAD_URL=$(get_download_url)
echo "Downloading insta-infra from: $DOWNLOAD_URL"

# Download the binary
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_NAME"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$BINARY_NAME"
else
    echo "Error: neither curl nor wget found. Please install either one and try again."
    exit 1
fi

# Make binary executable
chmod +x "$TMP_DIR/$BINARY_NAME"

# Determine install directory
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
if [ ! -w "$INSTALL_DIR" ]; then
    if command -v sudo >/dev/null 2>&1; then
        echo "Elevated privileges required to install to $INSTALL_DIR"
        sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    else
        echo "Error: Cannot write to $INSTALL_DIR and sudo is not available."
        echo "Please run this script with sudo or install manually."
        exit 1
    fi
else
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
fi

echo "Successfully installed insta-infra to $INSTALL_DIR/$BINARY_NAME"
echo "Run 'insta --ui' to start the Web UI or 'insta --help' for CLI usage"