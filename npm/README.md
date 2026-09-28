# npm packaging (maintainers)

The AUXO MCP server is a Go binary. To run it through `npx`, for the Claude plugin in `plugin/` and for other MCP clients, it is published to npm in the same way as esbuild and Biome: one package per platform holding the binary, plus one small launcher package that depends on all of them.

## Layout

```
npm/
  auxo-mcp/                    @on2it/auxo-mcp, the launcher
    bin/auxo-mcp.js            finds and runs the platform binary
  platforms/<os>-<cpu>/        @on2it/auxo-mcp-<os>-<cpu>, one per platform
    bin/auxo-mcp-server[.exe]  copied in at release time, not committed
```

| Go build (`dist/`)                  | npm package                    |
| ----------------------------------- | ------------------------------ |
| `auxo-mcp-server-darwin-arm64`      | `@on2it/auxo-mcp-darwin-arm64` |
| `auxo-mcp-server-darwin-amd64`      | `@on2it/auxo-mcp-darwin-x64`   |
| `auxo-mcp-server-linux-amd64`       | `@on2it/auxo-mcp-linux-x64`    |
| `auxo-mcp-server-linux-arm64`       | `@on2it/auxo-mcp-linux-arm64`  |
| `auxo-mcp-server-windows-amd64.exe` | `@on2it/auxo-mcp-win32-x64`    |

The release also builds `windows-arm64`, which is not published to npm. Windows on Arm can run the x64 package under emulation only if Node itself is x64; a native arm64 Node reports `win32-arm64` and gets the "no binary installed" error. Add a `win32-arm64` package if there is demand.

How it fits together:

- Each platform package sets `os` and `cpu`, so npm installs only the one that matches the machine and skips the rest. Their `files` list holds only the binary.
- The launcher lists every platform package in `optionalDependencies`, pinned to its own exact version.
- `bin/auxo-mcp.js` resolves `@on2it/auxo-mcp-${process.platform}-${process.arch}`, restores the executable bit if the tarball lost it, and spawns the binary with `stdio: "inherit"`, passing arguments and environment through. It forwards `SIGINT` and `SIGTERM` and exits with the server's exit code or signal. It never writes to stdout, which is the MCP channel.
- No package has dependencies beyond this, and none has install or lifecycle scripts.
- There is no `.npmrc` in the repository. The Claude plugin directory blocks a plugin that uses `npx` when the repository sets a custom registry.

The committed `package.json` files carry the placeholder version `0.0.0-dev`. `scripts/npm-prepare.sh <version> <dist-dir>` copies the binaries in and writes the real version into all six files, including the `optionalDependencies` pins. It also holds the npm scope (`NPM_SCOPE`); the launcher reads the scope from its own package name, so changing the scope means editing `NPM_SCOPE`, the pin in `plugin/.mcp.json` and the docs.

After running `npm-prepare.sh` locally, restore the committed files with `git checkout npm/`.

## Release flow

The `npm` job in `.github/workflows/release.yml` runs on every `v*` tag, after the binaries are built:

1. `scripts/check-versions.sh <tag>` fails the job unless `plugin/.claude-plugin/plugin.json`, the pin in `plugin/.mcp.json` and `extension/manifest.json` all equal the tag.
2. `scripts/npm-prepare.sh` prepares the six packages at the tag's version.
3. The five platform packages are published first and the launcher last, each with `npm publish --access public --provenance`. A package whose version is already on the registry is skipped, so re-running a failed job is safe. Prerelease versions (with a `-`) are published under the `next` dist-tag instead of `latest`.

See [CONTRIBUTING.md](../CONTRIBUTING.md#releasing) for the full release procedure.

## First publish with NPM_TOKEN

npm trusted publishing can only be configured for a package that already exists, so the first release uses a token:

1. Make sure the `@on2it` organisation exists on npmjs.com and that the maintainers are members with publish rights. If the scope is not available, change `NPM_SCOPE` as described above.
2. As an organisation member, create a **granular access token** with read and write permission on the `@on2it` scope (or a classic **Automation** token), and "bypass two-factor authentication" enabled so CI can publish.
3. Add it to the GitHub repository as the Actions secret `NPM_TOKEN` (Settings, Secrets and variables, Actions).
4. Tag the release. The workflow passes the secret to npm as `NODE_AUTH_TOKEN`.
5. Confirm that all six packages appear on npmjs.com at the new version and show the provenance badge.

## Switch to trusted publishing

Once the packages exist:

1. On npmjs.com, open each of the six packages, go to **Settings**, **Trusted publishing**, and add a GitHub Actions publisher with organisation `on2itsecurity`, repository `auxo-mcp` and workflow filename `release.yml` (leave environment empty).
2. Optionally set **Publishing access** to "Require two-factor authentication and disallow tokens" so only the workflow can publish.
3. Delete the `NPM_TOKEN` repository secret and revoke the token on npmjs.com.

The workflow already grants `id-token: write` and installs npm 11, which is required for trusted publishing. With no token present, npm authenticates through OIDC.

## Deprecate or withdraw a bad version

Prefer releasing a fixed version: bump with `scripts/bump-version.sh`, merge and tag. The plugin pins an exact version, so users only move when `plugin/.mcp.json` on `main` points at the new one.

To warn anyone who installs the bad version, deprecate the launcher and the platform packages:

```bash
V=0.2.0
MSG="Broken release, use 0.2.1 or later"
for p in auxo-mcp auxo-mcp-darwin-arm64 auxo-mcp-darwin-x64 auxo-mcp-linux-x64 auxo-mcp-linux-arm64 auxo-mcp-win32-x64; do
  npm deprecate "@on2it/$p@$V" "$MSG"
done
```

Undo with an empty message: `npm deprecate "@on2it/auxo-mcp@$V" ""`.

`npm unpublish` is possible only within 72 hours of publishing and only when nothing depends on the version. Never unpublish a version that a released `plugin/.mcp.json` pins, since existing installs would stop starting. The version number can never be reused.
