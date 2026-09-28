# AUXO by ON2IT

Work with the [AUXO](https://on2it.net) platform from Claude. The plugin runs the AUXO MCP server on your computer and adds two skills that guide Claude through ON2IT's way of working.

## What it does

- **Zero Trust:** list, create, update and delete protect surfaces, locations and states; browse the AUXO security measure catalog and track measure assignment, implementation and evidence per protect surface; map transaction flows between protect surfaces and to or from outside the organisation; view contacts and assets.
- **Readiness assessments:** run an interactive Zero Trust readiness assessment, submit it and review earlier assessments.
- **Case management:** look up, create and update support cases with ON2IT, add notes, escalate, de-escalate and request closure.

The `protect-surface-design` skill walks Claude through defining a protect surface: identify the DAAS element, check existing protect surfaces and locations, create the protect surface, assign measures and map transaction flows. The `auxo-case-management` skill covers opening and handling cases. Both skills tell Claude to confirm with you before it creates, changes, escalates, closes or deletes anything.

## Requirements

- An AUXO account.
- A Zero Trust API token, a Tickets API token, or both. Request them from ON2IT support at [support@on2it.net](mailto:support@on2it.net). Each token enables only its own tools, so you can use one domain without the other.
- Node.js 18 or later on your computer, with `npx` on your `PATH`. The plugin uses `npx` to fetch and start the server.

## Install and configure

1. Install the plugin from the Claude plugin directory, or in Claude Code run `/plugin` and add it from a marketplace that lists it.
2. When the plugin is enabled, Claude asks for three settings:
   - **Zero Trust API Token** (optional, stored in your system's secure credential store)
   - **Tickets API Token** (optional, stored in your system's secure credential store)
   - **API URL** (optional, default `api.on2it.net`)
3. Leave a token empty to turn off that domain. To change a value later in Claude Code, run `/plugin configure auxo`, or open the plugin in `/plugin` and choose **Configure options**.
4. In Claude Code, run `/mcp` to check that the `auxo` server is connected.

## What it runs, sends and fetches

- **Runs:** `npx -y @on2it/auxo-mcp@<version>`, with the version pinned exactly in `.mcp.json`. The first start downloads `@on2it/auxo-mcp` and the one platform package that matches your computer, for example `@on2it/auxo-mcp-darwin-arm64`, from the public npm registry (`registry.npmjs.org`). npm keeps them in its local cache. The platform package contains the compiled AUXO MCP server, built from the open source code in [github.com/on2itsecurity/auxo-mcp](https://github.com/on2itsecurity/auxo-mcp) and published from GitHub Actions with npm provenance.
- **Talks to Claude:** the server runs locally and communicates with Claude over standard input and output only. It opens no network port.
- **Sends:** requests go only to the configured AUXO API endpoint (`api.on2it.net` by default) over HTTPS, authenticated with your tokens. Each request carries only the data needed for the tool call Claude makes, for example the fields of a protect surface you asked to create.
- **Stores:** nothing. The plugin and server keep no files, logs or caches of their own and send no telemetry. Tokens are held by Claude in your system's secure credential store and passed to the server as environment variables.

## Supported surfaces

The plugin starts a local server, so it works where Claude can run programs on your computer:

- Claude Code
- Cowork sessions that run on your computer

claude.ai chat in the browser and the Claude mobile apps do not start local servers. There the skills load, but the AUXO tools are not available.

On Windows, Claude Code starts `npx` directly. If the server does not start, run `npx -y @on2it/auxo-mcp@<version>` in a terminal to check that Node.js and `npx` work, and report the result to [support@on2it.net](mailto:support@on2it.net).

## Privacy Policy

ON2IT processes data sent to the AUXO platform as described in the [ON2IT Privacy & Cookie Notice](https://on2it.net/privacy/).

## Support and license

Report problems at [github.com/on2itsecurity/auxo-mcp/issues](https://github.com/on2itsecurity/auxo-mcp/issues) or e-mail [support@on2it.net](mailto:support@on2it.net). Security issues go to [responsibledisclosure@on2it.net](mailto:responsibledisclosure@on2it.net). Released under the MIT license, see `LICENSE`.
