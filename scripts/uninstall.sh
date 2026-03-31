#!/usr/bin/env bash

set -e

INSTALL_DIR="$HOME/.local/bin"

main() {
    rm -rf "$INSTALL_DIR/yay"
    echo "Uninstalled yay from $INSTALL_DIR"
}

main
