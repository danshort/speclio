#!/usr/bin/env bash
# Packages LecternApp as a .app bundle, code-signs it, and zips it for the
# Homebrew cask.
#
# Signing is Developer-ID-ready: by default it ad-hoc signs (required for Apple
# Silicon to run; fine for local dev). Set SIGN_IDENTITY to a "Developer ID
# Application: ..." identity for a distributable, notarizable build. Notarization
# is a separate gated step in the release workflow (needs network + credentials).
#
# Usage:   scripts/package.sh [version]            (default version: 0.0.0-dev)
# Signed:  SIGN_IDENTITY="Developer ID Application: …" scripts/package.sh 1.2.3
set -euo pipefail

VERSION="${1:-0.0.0-dev}"
APP_NAME="Lectern"
EXEC_NAME="LecternApp"
PKG_DIR="$(cd "$(dirname "$0")/.." && pwd)"   # macos/LecternApp
DIST="$PKG_DIR/dist"
APP="$DIST/$APP_NAME.app"

echo "==> swift build -c release"
swift build -c release --package-path "$PKG_DIR"
BIN="$(swift build -c release --package-path "$PKG_DIR" --show-bin-path)/$EXEC_NAME"

echo "==> assembling $APP_NAME.app"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$BIN" "$APP/Contents/MacOS/$EXEC_NAME"
sed "s/__VERSION__/$VERSION/g" "$PKG_DIR/Resources/Info.plist" > "$APP/Contents/Info.plist"

SIGN_IDENTITY="${SIGN_IDENTITY:--}"
if [ "$SIGN_IDENTITY" = "-" ]; then
    echo "==> ad-hoc code-signing (set SIGN_IDENTITY=<Developer ID> for distribution)"
    codesign --force --deep --options runtime --sign - "$APP"
else
    echo "==> code-signing with: $SIGN_IDENTITY (hardened runtime, timestamped)"
    codesign --force --deep --options runtime --timestamp --sign "$SIGN_IDENTITY" "$APP"
fi
codesign --verify --strict --verbose=2 "$APP"

echo "==> zipping for cask"
ZIP="$DIST/$APP_NAME-$VERSION.zip"
rm -f "$ZIP"
( cd "$DIST" && ditto -c -k --keepParent "$APP_NAME.app" "$(basename "$ZIP")" )

echo ""
echo "Built:  $APP"
echo "Zip:    $ZIP"
echo "SHA256: $(shasum -a 256 "$ZIP" | cut -d' ' -f1)"
