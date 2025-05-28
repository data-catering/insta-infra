#!/bin/bash

# Script to prepare resources for the Wails UI build
# This copies the necessary files from cmd/insta/resources to cmd/instaui/resources
# so they can be embedded in the UI binary

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SOURCE_DIR="$PROJECT_ROOT/cmd/insta/resources"
TARGET_DIR="$PROJECT_ROOT/cmd/instaui/resources"

echo "Preparing UI resources..."
echo "Source: $SOURCE_DIR"
echo "Target: $TARGET_DIR"

# Remove existing target directory if it exists
if [ -d "$TARGET_DIR" ]; then
    echo "Removing existing resources directory..."
    rm -rf "$TARGET_DIR"
fi

# Create target directory
echo "Creating resources directory..."
mkdir -p "$TARGET_DIR"

# Copy docker-compose files
echo "Copying docker-compose files..."
cp "$SOURCE_DIR/docker-compose.yaml" "$TARGET_DIR/"
cp "$SOURCE_DIR/docker-compose-persist.yaml" "$TARGET_DIR/"

# Copy data directory
echo "Copying data directory..."
cp -r "$SOURCE_DIR/data" "$TARGET_DIR/"

echo "UI resources prepared successfully!"
echo "Resources copied to: $TARGET_DIR" 