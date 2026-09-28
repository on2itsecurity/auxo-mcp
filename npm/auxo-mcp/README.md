# @on2it/auxo-mcp

The [AUXO](https://on2it.net) MCP server by ON2IT, packaged for npm. It gives MCP clients such as Claude Code access to AUXO Zero Trust (protect surfaces, locations, states, measures, transaction flows, readiness assessments) and AUXO case management.

Most people use it through the **AUXO by ON2IT** Claude plugin, which runs this package for you. See the [repository README](https://github.com/on2itsecurity/auxo-mcp#readme) for every install option.

## Run

```bash
AUXO_ZT_TOKEN=your-zt-token AUXO_TICKET_TOKEN=your-ticket-token npx -y @on2it/auxo-mcp@<version>
```

The server speaks MCP over stdio. Each token enables its own tools; leave one out to disable that domain. `AUXO_API_URL` defaults to `api.on2it.net`. Tokens are available from [support@on2it.net](mailto:support@on2it.net).

Example MCP client configuration:

```json
{
  "mcpServers": {
    "auxo": {
      "command": "npx",
      "args": ["-y", "@on2it/auxo-mcp@<version>"],
      "env": {
        "AUXO_ZT_TOKEN": "your-zt-token",
        "AUXO_TICKET_TOKEN": "your-ticket-token"
      }
    }
  }
}
```

## How it works

This package contains only a small launcher script. The server itself is a Go binary published in one package per platform, installed automatically as an optional dependency:

| Platform              | Package                         |
| --------------------- | ------------------------------- |
| macOS, Apple silicon  | `@on2it/auxo-mcp-darwin-arm64`  |
| macOS, Intel          | `@on2it/auxo-mcp-darwin-x64`    |
| Linux, x64            | `@on2it/auxo-mcp-linux-x64`     |
| Linux, arm64          | `@on2it/auxo-mcp-linux-arm64`   |
| Windows, x64          | `@on2it/auxo-mcp-win32-x64`     |

The launcher finds the binary for the current platform and runs it with the same arguments and environment. It has no dependencies of its own and no install scripts. If you install with `--no-optional` or `--omit=optional`, the binary is missing and the launcher exits with an error that says so.

Requires Node.js 18 or later.

## Privacy

The server sends requests only to the configured AUXO API endpoint and collects no telemetry. See the [ON2IT Privacy & Cookie Notice](https://on2it.net/privacy/).

## License

MIT
