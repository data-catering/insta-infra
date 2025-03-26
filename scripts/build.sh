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

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Build the binary
echo "Building for ${GOOS}/${GOARCH}..."
go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/insta

# Apply UPX compression if available and supported
if command -v upx >/dev/null 2>&1; then
    # Skip UPX on macOS ARM64
    if [ "$GOOS" != "darwin" ] || [ "$GOARCH" != "arm64" ]; then
        echo "Compressing with UPX..."
        upx -q --best --lzma "${BINARY_NAME}"
    else
        echo "Skipping UPX compression on macOS ARM64"
    fi
else
    echo "UPX not found, skipping compression"
fi

echo "Build complete: ${BINARY_NAME}" 