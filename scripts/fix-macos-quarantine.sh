#!/bin/bash

# Fix macOS Gatekeeper quarantine issue for Insta-Infra UI
# Run this script in the same directory as "Insta-Infra UI.app"

set -e

APP_NAME="Insta-Infra UI.app"

echo "🔧 Fixing macOS Gatekeeper issue for $APP_NAME"
echo ""

# Check if app exists
if [ ! -d "$APP_NAME" ]; then
    echo "❌ Error: $APP_NAME not found in current directory"
    echo "Please run this script in the same directory as the app"
    exit 1
fi

echo "📱 Found app: $APP_NAME"

# Remove quarantine attributes
echo "🧹 Removing quarantine attributes..."
if xattr -cr "$APP_NAME"; then
    echo "✅ Quarantine attributes removed successfully"
else
    echo "⚠️  Warning: Could not remove quarantine attributes"
    echo "   You may need to run with sudo or use the right-click method"
fi

echo ""
echo "🚀 Attempting to open the app..."
if open "$APP_NAME"; then
    echo "✅ App opened successfully!"
    echo ""
    echo "🎉 The app should now run without Gatekeeper warnings."
    echo "   If you still see warnings, try right-clicking the app and selecting 'Open'"
else
    echo "❌ Could not open app automatically"
    echo ""
    echo "📋 Manual steps:"
    echo "   1. Right-click on '$APP_NAME'"
    echo "   2. Select 'Open' from the context menu"
    echo "   3. Click 'Open' in the security dialog"
fi

echo ""
echo "ℹ️  For more help, visit: https://github.com/data-catering/insta-infra" 