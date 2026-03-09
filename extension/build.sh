#!/bin/bash
set -euo pipefail

# Build AUXO MCP Server Desktop Extension (.mcpb)
#
# Prerequisites:
#   - Compiled binaries in ../dist/
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

# Build missing binaries locally if Go is available
build_binary_if_missing() {
  local goos="$1" goarch="$2" output="$3"
  if [ -f "$output" ]; then
    return 0
  fi
  if ! command -v go &>/dev/null; then
    echo "Warning: $output not found and Go not available to compile it"
    return 1
  fi
  echo "Compiling $(basename "$output") locally..."
  mkdir -p "$DIST_DIR"
  (cd "$PROJECT_DIR/server" && CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -o "$output" .)
}

for arch in $DARWIN_ARCHS; do
  build_binary_if_missing darwin "$arch" "$DIST_DIR/auxo-mcp-server-darwin-${arch}"
done
for arch in $WINDOWS_ARCHS; do
  build_binary_if_missing windows "$arch" "$DIST_DIR/auxo-mcp-server-windows-${arch}.exe"
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

# Ensure binaries are executable
chmod +x "$BUILD_DIR/bin/"* 2>/dev/null || true

# Create .mcpb (ZIP archive)
rm -f "$OUTPUT"
(cd "$BUILD_DIR" && zip -r "$OUTPUT" .)

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
