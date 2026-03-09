# AUXO MCP Server

An [MCP](https://modelcontextprotocol.io) (Model Context Protocol) server that exposes [AUXO](https://on2it.net) Zero Trust and Case Management capabilities to AI assistants like Claude, VS Code Copilot, and other MCP-compatible clients.

## Features

**Zero Trust** — manage protect surfaces, locations, states, contacts, assets, security measures, and transaction flows.

**Case Management** — create, update, escalate, and close support cases/tickets.

Domains are automatically enabled based on which API tokens you provide.

## Prerequisites

You need an API token for each AUXO domain you want to use:

- **Zero Trust token** — required for protect surfaces, locations, states, measures, flows, contacts, and assets.
- **Tickets token** — required for case/ticket management.

You only need the token(s) for the domain(s) you intend to use; the server automatically enables features based on which tokens are provided. Tokens can be requested by contacting [ON2IT support](mailto:support@on2it.net).

## Installation

### Claude Desktop Extension

Download the latest `auxo-mcp-server.mcpb` from the [Releases](../../releases/latest) page and double-click it, or drag it into Claude Desktop settings. Claude will prompt you for your API tokens during setup.

### Pre-built Binaries

Download the binary for your platform from the [Releases](../../releases/latest) page:

| Platform      | Binary                              |
| ------------- | ----------------------------------- |
| macOS (ARM)   | `auxo-mcp-server-darwin-arm64`      |
| macOS (Intel) | `auxo-mcp-server-darwin-amd64`      |
| Linux (AMD64) | `auxo-mcp-server-linux-amd64`       |
| Linux (ARM64) | `auxo-mcp-server-linux-arm64`       |
| Windows       | `auxo-mcp-server-windows-amd64.exe` |

Make it executable (`chmod +x`) and place it on your `PATH`, or reference it directly in your MCP client configuration.

### Container

The container image is published to GitHub Container Registry:

```bash
docker pull ghcr.io/on2itsecurity/auxo-mcp:latest
```

Run it in HTTP mode (the container default):

```bash
docker run -p 8080:8080 \
  -e AUXO_ZT_TOKEN=your-token \
  -e AUXO_TICKET_TOKEN=your-ticket-token \
  ghcr.io/on2itsecurity/auxo-mcp:latest
```

### Build from Source

Requires Go 1.24+.

```bash
cd server
go build -o auxo-mcp-server .
```

## Configuration

Copy [`env.ini.dist`](env.ini.dist) as a reference for available environment variables.

| Variable                 | Description                              | Default                          |
| ------------------------ | ---------------------------------------- | -------------------------------- |
| `AUXO_ZT_TOKEN`          | Zero Trust API token                     | _(required for ZT features)_     |
| `AUXO_TICKET_TOKEN`      | Tickets/Case Management API token        | _(required for ticket features)_ |
| `AUXO_API_URL`           | AUXO API endpoint                        | `api.on2it.net`                  |
| `AUXO_SERVER_MODE`       | `STDIO` or `HTTP`                        | `STDIO`                          |
| `AUXO_SERVER_PORT`       | Port for HTTP mode                       | `8080`                           |
| `AUXO_ENABLE_ZERO_TRUST` | Explicitly enable/disable ZT domain      | auto-detected from token         |
| `AUXO_ENABLE_TICKETS`    | Explicitly enable/disable Tickets domain | auto-detected from token         |
| `AUXO_DEBUG`             | Enable debug logging for API calls       | `false`                          |

CLI flags `-mode` and `-port` take precedence over environment variables.

## MCP Client Configuration

### VS Code / Copilot

Add to `.vscode/mcp.json` in your workspace:

```json
{
  "servers": {
    "auxo": {
      "type": "stdio",
      "command": "path/to/auxo-mcp-server",
      "env": {
        "AUXO_ZT_TOKEN": "your-token",
        "AUXO_TICKET_TOKEN": "your-ticket-token"
      }
    }
  }
}
```

### Claude Desktop (manual)

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "auxo": {
      "command": "path/to/auxo-mcp-server",
      "env": {
        "AUXO_ZT_TOKEN": "your-token",
        "AUXO_TICKET_TOKEN": "your-ticket-token"
      }
    }
  }
}
```

### HTTP Mode

For clients that connect over HTTP/SSE:

```bash
auxo-mcp-server -mode HTTP -port 8080
```

Clients connect to `http://localhost:8080/sse`. Tokens and settings can also be passed as query parameters:

```
http://localhost:8080/sse?zt_token=TOKEN&ticket_token=TOKEN
```

> **Security considerations for HTTP mode:**
>
> - **Query parameter tokens are logged and cached.** Tokens in URLs may appear in server logs, browser history, proxy logs, and HTTP referrer headers. Prefer environment variables for token configuration whenever possible.
> - **No TLS by default.** The built-in HTTP server does not support TLS. If exposed beyond localhost, place it behind a reverse proxy (e.g. nginx, Caddy) with TLS termination to prevent tokens and API traffic from being transmitted in plain text.
> - **No authentication on the endpoint.** Anyone who can reach the `/sse` endpoint can use the server. Restrict access via firewall rules, network policies, or a reverse proxy with authentication.
> - **Bind to localhost only** when running locally. The server currently binds to all interfaces (`0.0.0.0`); use firewall rules to limit exposure.
>
> Query parameter overrides are primarily intended for local development and testing.

## Available Tools

### Zero Trust

| Tool                              | Description                                    |
| --------------------------------- | ---------------------------------------------- |
| `createProtectSurface`            | Create a new protect surface                   |
| `listProtectSurfaces`             | List all protect surfaces                      |
| `updateProtectSurface`            | Update an existing protect surface             |
| `deleteProtectSurface`            | Delete one or more protect surfaces            |
| `createLocation`                  | Create a new location                          |
| `listLocations`                   | List all locations                             |
| `updateLocation`                  | Update an existing location                    |
| `deleteLocation`                  | Delete one or more locations                   |
| `createState`                     | Create a new state                             |
| `listStates`                      | List all states                                |
| `updateState`                     | Update an existing state                       |
| `deleteState`                     | Delete one or more states                      |
| `listContacts`                    | List all contacts                              |
| `listAssets`                      | List all assets                                |
| `listMeasures`                    | Search security measures from the AUXO catalog |
| `listProtectSurfaceMeasures`      | List measures assigned to a protect surface    |
| `updateProtectSurfaceMeasure`     | Add or update measure implementation status    |
| `removeMeasureFromProtectSurface` | Remove a measure from a protect surface        |
| `createTransactionFlow`           | Create a flow between two protect surfaces     |
| `createExternalFlow`              | Create a flow to/from outside the organization |
| `listTransactionFlows`            | List all flows for a protect surface           |
| `deleteTransactionFlow`           | Delete a flow between two protect surfaces     |
| `deleteExternalFlow`              | Delete an external flow                        |

### Case Management

| Tool                       | Description                                    |
| -------------------------- | ---------------------------------------------- |
| `createCase`               | Create a new support case/ticket               |
| `getCase`                  | Get case details by ID                         |
| `updateCasePriority`       | Update case priority (1-4, where 1 is highest) |
| `updateCasePrimaryContact` | Update the primary contact email               |
| `updateCaseSubject`        | Update the case subject                        |
| `escalateCase`             | Escalate to higher priority attention          |
| `deescalateCase`           | De-escalate to normal handling                 |
| `addNoteToCase`            | Add a note/comment to a case                   |
| `closeCase`                | Request to close a case                        |

## License

See [LICENSE.md](LICENSE.md).
