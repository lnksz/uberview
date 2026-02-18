#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly LOCAL_BIN_DIR="$HOME/.local/bin"
readonly BINARY_PATH="$LOCAL_BIN_DIR/uberview"
readonly USER_SYSTEMD_DIR="$HOME/.config/systemd/user"
readonly USER_UNIT_PATH="$USER_SYSTEMD_DIR/uberview.service"
readonly REPO_UNIT_PATH="$SCRIPT_DIR/systemd/uberview.service"

cd "$SCRIPT_DIR"

mkdir -p "$LOCAL_BIN_DIR"
echo "Installing uberview to $LOCAL_BIN_DIR"
GOBIN="$LOCAL_BIN_DIR" go install -trimpath -ldflags="-s -w"
if command -v upx >/dev/null 2>&1; then
	echo "Compressing $BINARY_PATH with upx"
	upx -q "$BINARY_PATH" >/dev/null 2>&1
fi
echo "Successfully installed to $BINARY_PATH"

mkdir -p "$USER_SYSTEMD_DIR"
if [ ! -f "$USER_UNIT_PATH" ] || ! diff -q "$REPO_UNIT_PATH" "$USER_UNIT_PATH" >/dev/null 2>&1; then
	cp "$REPO_UNIT_PATH" "$USER_UNIT_PATH"
	echo "Updated $USER_UNIT_PATH"
	systemctl --user daemon-reload
fi

echo "Restarting uberview.service"
systemctl --user restart uberview.service
