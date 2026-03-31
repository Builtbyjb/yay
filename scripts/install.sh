#! /usr/bin/env bash

set -e

# Configuration
APP_NAME="yay"
APP_VERSION="0.1.0"
INSTALL_DIR="$HOME/.local/bin"


main() {
    go build -o "$INSTALL_DIR/$APP_NAME" "./cmd/main.go"
    chmod +x "$INSTALL_DIR/$APP_NAME"
    # xattr -d "com.apple.quarantine" "$INSTALL_DIR/$APP_NAME"
    echo "Installed $APP_NAME version $APP_VERSION to $INSTALL_DIR"
}

main
