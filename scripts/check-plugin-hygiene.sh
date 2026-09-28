#!/usr/bin/env bash
set -uo pipefail

# Assert the plugin folder meets the Claude plugin directory's file rules.
# Checks only files tracked by git, because those are what the directory reads.
#
# Usage: scripts/check-plugin-hygiene.sh [plugin-dir]   (default: plugin)

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PLUGIN_DIR="${1:-plugin}"
MAX_BYTES=$((256 * 1024))
cd "$ROOT_DIR" || exit 2

errors=0
fail() {
  echo "FAIL: $*"
  errors=$((errors + 1))
}

if [ ! -f "$PLUGIN_DIR/.claude-plugin/plugin.json" ]; then
  fail "$PLUGIN_DIR/.claude-plugin/plugin.json not found"
fi

# Folder names on the path to the plugin: letters, digits, dots, hyphens, underscores.
if ! [[ "$PLUGIN_DIR" =~ ^[A-Za-z0-9._/-]+$ ]]; then
  fail "plugin path $PLUGIN_DIR contains characters other than letters, digits, . - _"
fi

# git ls-files -s prints "<mode> <object> <stage>\t<path>".
while IFS=$'\t' read -r meta path; do
  mode="${meta%% *}"
  name="$(basename "$path")"

  case "$name" in
    .DS_Store | Thumbs.db | desktop.ini) fail "$path: OS metadata file" ;;
    .npmrc | bunfig.toml | uv.toml) fail "$path: package manager config file" ;;
    package.json | package-lock.json | npm-shrinkwrap.json | yarn.lock | pnpm-lock.yaml | bun.lock | bun.lockb)
      fail "$path: package manifest or lockfile" ;;
    *.mcpb | *.dxt | *.zip | *.exe | *.dll | *.so | *.dylib) fail "$path: bundle or binary" ;;
  esac
  case "/$path/" in
    */__MACOSX/*) fail "$path: __MACOSX entry" ;;
  esac

  case "$mode" in
    120000) fail "$path: symbolic link" ;;
    160000) fail "$path: git submodule" ;;
  esac
  [ "$mode" = "100755" ] && fail "$path: executable file (binaries and scripts are not expected)"

  if [ -f "$path" ] && [ ! -L "$path" ]; then
    size=$(wc -c <"$path" | tr -d ' ')
    [ "$size" -ge "$MAX_BYTES" ] && fail "$path: $size bytes, limit is under $MAX_BYTES"
    if head -c 200 "$path" | grep -q "^version https://git-lfs.github.com/spec/"; then
      fail "$path: Git LFS pointer"
    fi
    # Text only: a NUL byte means a binary file.
    if [ "$size" -gt 0 ] && [ "$(tr -d '\000' <"$path" | wc -c | tr -d ' ')" != "$size" ]; then
      fail "$path: binary content"
    fi
  fi
done < <(git ls-files -s -- "$PLUGIN_DIR")

# Untracked OS files in the working tree would be caught by git add; report them too.
while IFS= read -r junk; do
  fail "$junk: OS metadata file in working tree"
done < <(find "$PLUGIN_DIR" \( -name .DS_Store -o -name Thumbs.db -o -name desktop.ini -o -name __MACOSX \) 2>/dev/null)

# .gitattributes: no export-ignore or export-subst anywhere, and no filter
# attributes at the repository root, above the plugin folder or inside it.
while IFS= read -r attrs; do
  if grep -Eq '(^|[[:space:]])-?export-(ignore|subst)' "$attrs"; then
    fail "$attrs: uses export-ignore or export-subst"
  fi
  dir="$(dirname "$attrs")"
  case "$dir" in
    . | "$PLUGIN_DIR" | "$PLUGIN_DIR"/*) grep -Eq '(^|[[:space:]])filter=' "$attrs" && fail "$attrs: uses filter" ;;
  esac
done < <(git ls-files -- '.gitattributes' '*/.gitattributes')

# Credentials: token-looking values next to AUXO token names, and common
# secret formats, anywhere in the repository.
secret_hits="$(git grep -nIE \
  -e '(AUXO_ZT_TOKEN|AUXO_TICKET_TOKEN|zt_token|ticket_token)["'\'']?[[:space:]]*[:=][[:space:]]*["'\'']?[A-Za-z0-9_.+/=-]{24,}' \
  -e 'eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}' \
  -e '(npm_[A-Za-z0-9]{36}|gh[pousr]_[A-Za-z0-9]{36})' \
  -- . ':!scripts/check-plugin-hygiene.sh' || true)"
if [ -n "$secret_hits" ]; then
  fail "possible credentials in the repository:"
  echo "$secret_hits"
fi

if [ "$errors" -gt 0 ]; then
  echo "$errors hygiene check(s) failed"
  exit 1
fi
echo "Plugin hygiene checks passed for $PLUGIN_DIR"
