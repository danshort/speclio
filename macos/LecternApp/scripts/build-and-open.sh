#!/usr/bin/env bash
# Builds the Lectern .app bundle via package.sh, then reveals it in Finder.
#
# Usage:  scripts/build-and-open.sh [version]   (default version: 0.0.0-dev)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PKG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"   # macos/LecternApp
VERSION="${1:-0.0.0-dev}"

# Point at Xcode's toolchain (needed for actool to compile the .icon) unless the
# caller already set DEVELOPER_DIR.
export DEVELOPER_DIR="${DEVELOPER_DIR:-/Applications/Xcode.app/Contents/Developer}"

"$SCRIPT_DIR/package.sh" "$VERSION"
open "$PKG_DIR/dist"
