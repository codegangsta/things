#!/bin/bash
# Build and install the Things CLI Callback URL handler app
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUNDLE_DIR="$PROJECT_DIR/cmd/bundle"
APP_NAME="ThingsCLICallback.app"
INSTALL_DIR="$HOME/Applications"

echo "Building Things CLI Callback handler..."

# Create output directory if needed
mkdir -p "$INSTALL_DIR"

# Remove existing app if present
if [ -d "$INSTALL_DIR/$APP_NAME" ]; then
    echo "Removing existing installation..."
    rm -rf "$INSTALL_DIR/$APP_NAME"
fi

# Compile AppleScript to app bundle
echo "Compiling AppleScript..."
osacompile -o "$INSTALL_DIR/$APP_NAME" "$BUNDLE_DIR/url-handler.applescript"

# Copy Info.plist (overwrites default)
echo "Configuring app bundle..."
cp "$BUNDLE_DIR/Info.plist" "$INSTALL_DIR/$APP_NAME/Contents/Info.plist"

# Register with Launch Services
echo "Registering URL handler..."
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f "$INSTALL_DIR/$APP_NAME"

echo ""
echo "Installation complete!"
echo "The things-cli:// URL scheme is now registered."
echo ""
echo "Installed to: $INSTALL_DIR/$APP_NAME"
