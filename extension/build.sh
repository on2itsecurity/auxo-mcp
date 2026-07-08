#!/bin/bash
set -euo pipefail

# Build AUXO MCP Server Desktop Extension (.mcpb)
#
# Prerequisites:
#   - Go (binaries are always rebuilt from ../server; existing ../dist/
#     binaries are only reused when Go is not installed)
#   - npm (for mcpb CLI, optional - script can also create the ZIP directly)
#
# Usage:
#   ./build.sh                  # Build for macOS ARM64 + Windows AMD64 (most common)
#   ./build.sh --all            # Build with all platform binaries
#   ./build.sh --darwin-only    # Build for macOS only

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_DIR/dist"
BUILD_DIR="$SCRIPT_DIR/build"
OUTPUT="$SCRIPT_DIR/auxo-mcp-server.mcpb"

# Clean previous build
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/bin"

# Copy manifest and icon
cp "$SCRIPT_DIR/manifest.json" "$BUILD_DIR/"
if [ -f "$SCRIPT_DIR/icon.png" ]; then
  cp "$SCRIPT_DIR/icon.png" "$BUILD_DIR/"
  echo "Icon included: icon.png"
else
  echo "Warning: icon.png not found in $SCRIPT_DIR — extension will have no icon"
fi

# Determine which binaries to include
case "${1:-}" in
  --all)
    echo "Building with all platform binaries..."
    DARWIN_ARCHS="amd64 arm64"
    WINDOWS_ARCHS="amd64 arm64"
    ;;
  --darwin-only)
    echo "Building for macOS only..."
    DARWIN_ARCHS="amd64 arm64"
    WINDOWS_ARCHS=""
    ;;
  *)
    echo "Building for macOS ARM64 + Windows AMD64 (most common)..."
    DARWIN_ARCHS="arm64"
    WINDOWS_ARCHS="amd64"
    ;;
esac

# Always rebuild binaries so the bundle never packages stale ones; falls back
# to an existing binary only when Go is not installed.
build_binary() {
  local goos="$1" goarch="$2" output="$3"
  if ! command -v go &>/dev/null; then
    if [ -f "$output" ]; then
      echo "Warning: Go not available, reusing existing $output"
      return 0
    fi
    echo "Warning: $output not found and Go not available to compile it"
    return 1
  fi
  echo "Compiling $(basename "$output")..."
  mkdir -p "$DIST_DIR"
  (cd "$PROJECT_DIR/server" && CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -o "$output" .)
}

for arch in $DARWIN_ARCHS; do
  build_binary darwin "$arch" "$DIST_DIR/auxo-mcp-server-darwin-${arch}"
done
for arch in $WINDOWS_ARCHS; do
  build_binary windows "$arch" "$DIST_DIR/auxo-mcp-server-windows-${arch}.exe"
done

# Check if we can create a universal macOS binary
DARWIN_BINS=()
for arch in $DARWIN_ARCHS; do
  bin="$DIST_DIR/auxo-mcp-server-darwin-${arch}"
  if [ ! -f "$bin" ]; then
    echo "Warning: $bin not found, skipping"
    continue
  fi
  DARWIN_BINS+=("$bin")
done

if [ ${#DARWIN_BINS[@]} -eq 2 ] && command -v lipo &>/dev/null; then
  echo "Creating universal macOS binary..."
  lipo -create "${DARWIN_BINS[@]}" -output "$BUILD_DIR/bin/auxo-mcp-server"
elif [ ${#DARWIN_BINS[@]} -ge 1 ]; then
  echo "Copying macOS binary: ${DARWIN_BINS[0]}"
  cp "${DARWIN_BINS[0]}" "$BUILD_DIR/bin/auxo-mcp-server"
else
  echo "Warning: No macOS binaries found"
fi

# Copy Windows binary
for arch in $WINDOWS_ARCHS; do
  bin="$DIST_DIR/auxo-mcp-server-windows-${arch}.exe"
  if [ -f "$bin" ]; then
    echo "Copying Windows binary: $bin"
    cp "$bin" "$BUILD_DIR/bin/auxo-mcp-server.exe"
    break  # Only need one Windows binary in the bundle
  else
    echo "Warning: $bin not found, skipping"
  fi
done

# Binaries must be executable inside the archive (zip preserves these bits);
# upstream steps like CI artifact download are known to strip them.
chmod 755 "$BUILD_DIR/bin/"* 2>/dev/null || {
  echo "ERROR: no binaries in $BUILD_DIR/bin to package" >&2
  exit 1
}

# Create .mcpb (ZIP archive)
rm -f "$OUTPUT"
(cd "$BUILD_DIR" && zip -r "$OUTPUT" .)

# Verify the packaged binaries kept their executable bit; a bundle without it
# installs fine but fails at spawn time with "Permission denied".
if ! command -v zipinfo &>/dev/null; then
  echo "ERROR: zipinfo not available to verify bundle permissions" >&2
  exit 1
fi
for entry in bin/auxo-mcp-server bin/auxo-mcp-server.exe; do
  perms="$(zipinfo "$OUTPUT" "$entry" 2>/dev/null | awk 'NR==1 {print $1}')" || true
  if [ -z "$perms" ]; then
    continue  # entry not in this bundle variant (e.g. --darwin-only)
  fi
  case "$perms" in
    -??x*) ;;
    *)
      echo "ERROR: $entry is packaged without its executable bit ($perms)." >&2
      echo "The extension would fail with 'Failed to spawn process: Permission denied'." >&2
      exit 1
      ;;
  esac
done

echo ""
echo "Built: $OUTPUT"
echo "Contents:"
(cd "$BUILD_DIR" && find . -type f | sort)
echo ""
echo "Size: $(du -h "$OUTPUT" | cut -f1)"
echo ""
echo "To install: Double-click the .mcpb file or drag it into Claude Desktop settings."

# Clean up build directory
rm -rf "$BUILD_DIR"
