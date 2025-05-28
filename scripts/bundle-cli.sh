#!/bin/bash

# Bundle CLI binary into Wails app for macOS
# This script runs after Wails build to embed the CLI binary

set -e

echo "Bundling CLI binary into Wails app..."

# Determine the platform
PLATFORM=$(uname -s)
ARCH=$(uname -m)

# Convert architecture names
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
esac

# Set paths
CLI_BINARY="insta"
WAILS_APP_DIR="cmd/instaui/build/bin"

if [ "$PLATFORM" = "Darwin" ]; then
    APP_BUNDLE="$WAILS_APP_DIR/Insta-Infra UI.app"
    RESOURCES_DIR="$APP_BUNDLE/Contents/Resources"
    MACOS_DIR="$APP_BUNDLE/Contents/MacOS"
    
    # Check if app bundle exists
    if [ ! -d "$APP_BUNDLE" ]; then
        echo "Error: Wails app bundle not found at $APP_BUNDLE"
        exit 1
    fi
    
    # Check if CLI binary exists
    if [ ! -f "$CLI_BINARY" ]; then
        echo "Error: CLI binary not found at $CLI_BINARY"
        echo "Make sure to run 'make build' first"
        exit 1
    fi
    
    # Create Resources directory if it doesn't exist
    mkdir -p "$RESOURCES_DIR"
    
    # Copy CLI binary to Resources
    echo "Copying CLI binary to app bundle..."
    cp "$CLI_BINARY" "$RESOURCES_DIR/insta-cli"
    chmod +x "$RESOURCES_DIR/insta-cli"
    
    # Create a wrapper script that sets up the environment
    cat > "$RESOURCES_DIR/insta-wrapper.sh" << 'EOF'
#!/bin/bash

# Get the directory where this script is located (Resources directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_BINARY="$SCRIPT_DIR/insta-cli"

# Set up environment for the CLI
export PATH="/usr/local/bin:/opt/homebrew/bin:$PATH"

# Add common Docker/Podman paths
export PATH="/Applications/Docker.app/Contents/Resources/bin:$PATH"
export PATH="/usr/local/bin:$PATH"

# Execute the CLI with all arguments
exec "$CLI_BINARY" "$@"
EOF
    
    chmod +x "$RESOURCES_DIR/insta-wrapper.sh"
    
    echo "CLI binary bundled successfully!"
    echo "Binary location: $RESOURCES_DIR/insta-cli"
    echo "Wrapper script: $RESOURCES_DIR/insta-wrapper.sh"
    
elif [ "$PLATFORM" = "Linux" ]; then
    # For Linux, we'll put the binary alongside the main executable
    if [ ! -f "$CLI_BINARY" ]; then
        echo "Error: CLI binary not found at $CLI_BINARY"
        exit 1
    fi
    
    LINUX_BINARY_DIR="$WAILS_APP_DIR"
    cp "$CLI_BINARY" "$LINUX_BINARY_DIR/insta-cli"
    chmod +x "$LINUX_BINARY_DIR/insta-cli"
    
    echo "CLI binary bundled for Linux at $LINUX_BINARY_DIR/insta-cli"
    
elif [ "$PLATFORM" = "MINGW"* ] || [ "$PLATFORM" = "CYGWIN"* ]; then
    # For Windows
    CLI_BINARY="insta.exe"
    if [ ! -f "$CLI_BINARY" ]; then
        echo "Error: CLI binary not found at $CLI_BINARY"
        exit 1
    fi
    
    WINDOWS_BINARY_DIR="$WAILS_APP_DIR"
    cp "$CLI_BINARY" "$WINDOWS_BINARY_DIR/insta-cli.exe"
    
    echo "CLI binary bundled for Windows at $WINDOWS_BINARY_DIR/insta-cli.exe"
else
    echo "Warning: Unknown platform $PLATFORM, skipping CLI bundling"
fi

echo "CLI bundling complete!" 