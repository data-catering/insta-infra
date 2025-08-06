#!/bin/bash

# Exit on error
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Make sure we're in the root directory
cd "$(dirname "$0")/.."

# Version from git tag or default
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")}
BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}
BINARY_NAME=${BINARY_NAME:-insta}

# Release directory
RELEASE_DIR="release"
mkdir -p "$RELEASE_DIR"

# Function to print colored messages
print_status() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to calculate SHA256
calculate_sha256() {
    local file="$1"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        shasum -a 256 "$file" | cut -d' ' -f1
    else
        sha256sum "$file" | cut -d' ' -f1
    fi
}

# Function to build binaries for all platforms
build_all_platforms() {
    print_status "Building binaries for all platforms..."
    
    # Define platforms to build
    local platforms="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"
    
    for platform in $platforms; do
        local goos=${platform%/*}
        local goarch=${platform##*/}
        local output_name="${BINARY_NAME}-${goos}-${goarch}"
        
        # Add .exe extension for Windows
        if [ "$goos" = "windows" ]; then
            output_name="${output_name}.exe"
        fi
        
        print_status "Building for ${goos}/${goarch}..."
        
        # Build the binary
        CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
            -ldflags "-s -w -X main.version=$VERSION -X main.buildTime=$BUILD_TIME" \
            -o "${RELEASE_DIR}/${output_name}" ./cmd/insta
        
        # Apply UPX compression if available and supported
        if command -v upx >/dev/null 2>&1; then
            # Skip UPX on macOS (causes notarization issues)
            if [ "$goos" != "darwin" ]; then
                print_status "Compressing ${output_name} with UPX..."
                upx -q --best --lzma "${RELEASE_DIR}/${output_name}"
            else
                print_warning "Skipping UPX compression for ${output_name} (macOS)"
            fi
        else
            print_warning "UPX not found, skipping compression for ${output_name}"
        fi
    done
}

# Function to create checksums file
create_checksums() {
    print_status "Creating checksums file..."
    
    {
        echo "# SHA256 Checksums for insta-infra ${VERSION}"
        echo "# Generated on $(date -u)"
        echo ""
        
        # Calculate checksums for all binaries
        for file in "${RELEASE_DIR}"/insta-*; do
            if [ -f "$file" ]; then
                local filename=$(basename "$file")
                local checksum=$(calculate_sha256 "$file")
                echo "${checksum}  ${filename}"
            fi
        done
    } > "${RELEASE_DIR}/checksums.txt"
}

# Main build process
main() {
    print_status "Starting simplified release build process..."
    print_status "Version: ${VERSION}"
    print_status "Build Time: ${BUILD_TIME}"
    
    # Build binaries for all platforms
    build_all_platforms
    
    # Create checksums file
    create_checksums
    
    print_status "Release build completed successfully!"
    print_status "Release files are in ${RELEASE_DIR}/"
    
    # List the generated files
    echo ""
    echo "Generated files:"
    ls -la "${RELEASE_DIR}/"
}

# Run the main function
main 