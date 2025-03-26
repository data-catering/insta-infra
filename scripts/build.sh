#!/bin/bash
set -e

# Make sure we're in the root directory
cd "$(dirname "$0")/.."

# Default values
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}
BINARY_NAME=${BINARY_NAME:-insta}
RELEASE=${RELEASE:-false}

# Set output directory based on RELEASE flag
if [ "$RELEASE" = "true" ]; then
    OUTPUT_DIR="release"
else
    OUTPUT_DIR="."
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Build the binary
echo "Building for ${GOOS}/${GOARCH}..."
go build -ldflags "${LDFLAGS}" -o "${OUTPUT_DIR}/${BINARY_NAME}" ./cmd/insta

# Apply UPX compression if available and supported
if command -v upx >/dev/null 2>&1; then
    # Skip UPX on macOS ARM64
    if [ "$GOOS" != "darwin" ] || [ "$GOARCH" != "arm64" ]; then
        echo "Compressing with UPX..."
        upx -q --best --lzma "${OUTPUT_DIR}/${BINARY_NAME}"
    else
        echo "Skipping UPX compression on macOS ARM64"
    fi
else
    echo "UPX not found, skipping compression"
fi

echo "Build complete: ${OUTPUT_DIR}/${BINARY_NAME}" 