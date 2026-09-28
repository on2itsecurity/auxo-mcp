#!/usr/bin/env bash
set -euo pipefail

# Prepare the npm packages in npm/ for publishing.
#
# Usage: scripts/npm-prepare.sh <version> <dist-dir>
#
#   <version>   semver without a leading "v", for example 0.2.0
#   <dist-dir>  directory holding the binaries built by the release workflow,
#               named auxo-mcp-server-<goos>-<goarch>[.exe]
#
# The script copies each binary into its platform package under
# npm/platforms/<os>-<cpu>/bin/, copies the license next to it, and writes the
# version and npm scope into all six package.json files, including the exact
# optionalDependencies pins in the launcher package.
#
# Every binary must be present unless NPM_PREPARE_ALLOW_MISSING=1 is set; CI
# and local tests use that to prepare only the current platform.
#
# Requires node (used to edit package.json files).

# npm scope for every package. Change it here if the scope is unavailable; the
# launcher derives the scope from its own package name at runtime.
NPM_SCOPE="@on2it"

# Go build target -> npm platform package directory (Go amd64 is Node x64).
PLATFORMS=(
  "darwin-arm64:darwin-arm64"
  "darwin-amd64:darwin-x64"
  "linux-amd64:linux-x64"
  "linux-arm64:linux-arm64"
  "windows-amd64:win32-x64"
)

if [ $# -ne 2 ]; then
  echo "Usage: $0 <version> <dist-dir>" >&2
  exit 2
fi

VERSION="$1"
DIST_DIR="$2"

if [[ "$VERSION" == v* ]]; then
  echo "ERROR: pass the version without a leading v (got $VERSION)" >&2
  exit 2
fi
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "ERROR: $VERSION is not a valid semver version" >&2
  exit 2
fi
if [ ! -d "$DIST_DIR" ]; then
  echo "ERROR: dist directory $DIST_DIR does not exist" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
NPM_DIR="$ROOT_DIR/npm"
LAUNCHER_DIR="$NPM_DIR/auxo-mcp"

# set_package_json <file> <name>: set name and version in a package.json file.
set_package_json() {
  node -e '
    const fs = require("fs");
    const [file, name, version] = process.argv.slice(1);
    const pkg = JSON.parse(fs.readFileSync(file, "utf8"));
    pkg.name = name;
    pkg.version = version;
    fs.writeFileSync(file, JSON.stringify(pkg, null, 2) + "\n");
  ' "$1" "$2" "$VERSION"
}

missing=0
optional_deps=()
for entry in "${PLATFORMS[@]}"; do
  go_target="${entry%%:*}"
  npm_platform="${entry#*:}"
  pkg_dir="$NPM_DIR/platforms/$npm_platform"
  pkg_name="$NPM_SCOPE/auxo-mcp-$npm_platform"
  ext=""
  [[ "$npm_platform" == win32-* ]] && ext=".exe"

  src="$DIST_DIR/auxo-mcp-server-$go_target$ext"
  dest="$pkg_dir/bin/auxo-mcp-server$ext"

  rm -rf "$pkg_dir/bin"
  if [ -f "$src" ]; then
    mkdir -p "$pkg_dir/bin"
    cp "$src" "$dest"
    chmod 755 "$dest"
    echo "Copied $(basename "$src") -> ${dest#"$ROOT_DIR"/}"
  else
    echo "WARNING: $src not found" >&2
    missing=$((missing + 1))
  fi

  cp "$ROOT_DIR/LICENSE.md" "$pkg_dir/LICENSE"
  set_package_json "$pkg_dir/package.json" "$pkg_name"
  optional_deps+=("$pkg_name")
done

if [ "$missing" -gt 0 ] && [ "${NPM_PREPARE_ALLOW_MISSING:-}" != "1" ]; then
  echo "ERROR: $missing binaries missing from $DIST_DIR (set NPM_PREPARE_ALLOW_MISSING=1 to allow)" >&2
  exit 1
fi

# Launcher: name, version and exact pins on every platform package.
set_package_json "$LAUNCHER_DIR/package.json" "$NPM_SCOPE/auxo-mcp"
node -e '
  const fs = require("fs");
  const [file, version, ...deps] = process.argv.slice(1);
  const pkg = JSON.parse(fs.readFileSync(file, "utf8"));
  pkg.optionalDependencies = Object.fromEntries(deps.sort().map((d) => [d, version]));
  fs.writeFileSync(file, JSON.stringify(pkg, null, 2) + "\n");
' "$LAUNCHER_DIR/package.json" "$VERSION" "${optional_deps[@]}"

echo "Prepared $NPM_SCOPE/auxo-mcp and ${#PLATFORMS[@]} platform packages at version $VERSION"
