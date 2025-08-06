#!/bin/bash
set -e

# Build the demo version
echo "Building demo version..."
npm run build -- --config vite.demo.config.js

# Create docs/demo/ui directory if it doesn't exist
mkdir -p ../../../docs/demo/ui

# Copy the built files to docs/demo/ui
echo "Copying files to docs/demo/ui..."
cp -r dist-demo/* ../../../docs/demo/ui/

echo "Demo build complete!"