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
    BINARY_NAME="${BINARY_NAME}-web-${GOOS}-${GOARCH}"
else
    OUTPUT_DIR="."
    BINARY_NAME="${BINARY_NAME}"
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Print build information
echo "Building browser-based Web UI binary:"
echo "  OS: ${GOOS}"
echo "  ARCH: ${GOARCH}"
echo "  Output: ${OUTPUT_DIR}/${BINARY_NAME}"
echo "  Version: ${VERSION}"
echo "  Build Time: ${BUILD_TIME}"

# Step 1: Build React frontend
echo "Building React frontend..."
cd cmd/insta/frontend

# Check if node_modules exists, install dependencies if not
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi

# Build the frontend (includes favicon generation)
npm run build

# Step 2: Copy frontend build to embed location
echo "Copying frontend build to embed location..."
cd ../../..
mkdir -p cmd/insta/ui/dist
cp -r cmd/insta/frontend/dist/* cmd/insta/ui/dist/

# Step 3: Build Go binary with embedded frontend
echo "Building Go binary with embedded frontend..."
CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "${OUTPUT_DIR}/${BINARY_NAME}" ./cmd/insta

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

echo "Browser-based Web UI build complete: ${OUTPUT_DIR}/${BINARY_NAME}"
echo ""
echo "Usage:"
echo "  ${OUTPUT_DIR}/${BINARY_NAME}                    # Launch web UI (default)"
echo "  ${OUTPUT_DIR}/${BINARY_NAME} --ui               # Launch web UI with browser"
echo "  ${OUTPUT_DIR}/${BINARY_NAME} --port 9310       # Launch on specific port"
echo "  ${OUTPUT_DIR}/${BINARY_NAME} --no-browser       # Start server without opening browser"
echo "  ${OUTPUT_DIR}/${BINARY_NAME} -l                 # List services (CLI mode)"
echo "  ${OUTPUT_DIR}/${BINARY_NAME} postgres           # Start PostgreSQL (CLI mode)" 