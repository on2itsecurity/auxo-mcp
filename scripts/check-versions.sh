#!/usr/bin/env bash
set -euo pipefail

# Fail unless plugin/.claude-plugin/plugin.json, the npx pin in
# plugin/.mcp.json and extension/manifest.json carry the same version. On a tag
# build (GITHUB_REF_TYPE=tag), or when a tag is passed as the first argument,
# that version must also equal the tag without its leading "v".
#
# Usage: scripts/check-versions.sh [tag]

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

TAG="${1:-}"
if [ -z "$TAG" ] && [ "${GITHUB_REF_TYPE:-}" = "tag" ]; then
  TAG="${GITHUB_REF_NAME:-}"
fi

node -e '
  const fs = require("fs");
  const path = require("path");
  const [root, tag] = process.argv.slice(1);
  const read = (rel) => JSON.parse(fs.readFileSync(path.join(root, rel), "utf8"));

  const pin = read("plugin/.mcp.json").mcpServers.auxo.args.find((a) => /^@[^/]+\/auxo-mcp@/.test(a)) || "";
  const versions = {
    "plugin/.claude-plugin/plugin.json": read("plugin/.claude-plugin/plugin.json").version,
    "plugin/.mcp.json (npx pin)": pin.replace(/^@[^/]+\/auxo-mcp@/, ""),
    "extension/manifest.json": read("extension/manifest.json").version,
  };
  if (tag) versions[`git tag ${tag}`] = tag.replace(/^v/, "");

  for (const [source, v] of Object.entries(versions)) console.log(`${source.padEnd(36)} ${v || "(missing)"}`);

  if (!/^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$/.test(versions["plugin/.mcp.json (npx pin)"])) {
    console.error("ERROR: plugin/.mcp.json must pin @<scope>/auxo-mcp to an exact version");
    process.exit(1);
  }
  if (new Set(Object.values(versions)).size !== 1) {
    console.error("ERROR: versions disagree; run scripts/bump-version.sh <version>");
    process.exit(1);
  }
  console.log("Versions match");
' "$ROOT_DIR" "$TAG"
