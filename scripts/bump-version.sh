#!/usr/bin/env bash
set -euo pipefail

# Set the release version in every committed file that carries it:
#   plugin/.claude-plugin/plugin.json   "version"
#   plugin/.mcp.json                    the exact npx pin @on2it/auxo-mcp@<version>
#   extension/manifest.json             "version"
#
# Usage: scripts/bump-version.sh <version>   (semver without a leading "v")
#
# The npm package.json files are not touched; the release workflow sets their
# version from the git tag (see scripts/npm-prepare.sh). Requires node.

if [ $# -ne 1 ]; then
  echo "Usage: $0 <version>" >&2
  exit 2
fi
VERSION="$1"
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "ERROR: $VERSION is not a semver version without a leading v" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

node -e '
  const fs = require("fs");
  const path = require("path");
  const [root, version] = process.argv.slice(1);
  const edit = (rel, fn) => {
    const file = path.join(root, rel);
    const data = JSON.parse(fs.readFileSync(file, "utf8"));
    fn(data);
    fs.writeFileSync(file, JSON.stringify(data, null, 2) + "\n");
    console.log(`${rel}: ${version}`);
  };

  edit("plugin/.claude-plugin/plugin.json", (m) => { m.version = version; });
  edit("extension/manifest.json", (m) => { m.version = version; });
  edit("plugin/.mcp.json", (m) => {
    const args = m.mcpServers.auxo.args;
    const i = args.findIndex((a) => /^@[^/]+\/auxo-mcp@/.test(a));
    if (i === -1) throw new Error("no @<scope>/auxo-mcp@<version> argument in plugin/.mcp.json");
    args[i] = args[i].replace(/@[^@]+$/, "@" + version);
  });
' "$ROOT_DIR" "$VERSION"
