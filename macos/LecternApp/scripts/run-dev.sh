#!/usr/bin/env bash
# Builds and runs the Lectern app in debug mode (no .app bundle) for fast
# iteration. Ctrl-C to quit.
#
# Usage:  scripts/run-dev.sh
set -euo pipefail

PKG_DIR="$(cd "$(dirname "$0")/.." && pwd)"   # macos/LecternApp

# Point at Xcode's toolchain unless the caller already set DEVELOPER_DIR.
export DEVELOPER_DIR="${DEVELOPER_DIR:-/Applications/Xcode.app/Contents/Developer}"

swift run --package-path "$PKG_DIR"
