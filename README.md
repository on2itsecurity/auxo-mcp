# AUXO MCP Server

An [MCP](https://modelcontextprotocol.io) (Model Context Protocol) server that exposes [AUXO](https://on2it.net) Zero Trust and Case Management capabilities to AI assistants like Claude, VS Code Copilot, and other MCP-compatible clients.

## Features

**Zero Trust** — manage protect surfaces, locations, states, contacts, assets, security measures, and transaction flows. Run interactive Zero Trust readiness assessments.

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

> [!NOTE]
> On macOS the extension manifest launches the bundled binary through `/bin/sh -c 'chmod +x "$0" 2>/dev/null; exec "$0"'`. This is a deliberate workaround: Claude Desktop's extension extractor does not preserve the executable bit on bundled binaries, so the wrapper restores it before exec'ing the server. It runs no other commands and touches no files outside the extension directory.

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
| `listProtectSurfaces`             | List protect surfaces (with optional filters)  |
| `getProtectSurface`               | Get full details of a protect surface by ID    |
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

### Readiness Assessments

| Tool                         | Description                                                          |
| ---------------------------- | -------------------------------------------------------------------- |
| `startReadinessAssessment`   | Start an assessment — opens an interactive questionnaire (MCP App)   |
| `getReadinessQuestions`      | Get the readiness questionnaire (strategical/tactical/operational)   |
| `createReadinessAssessment`  | Submit a completed readiness assessment                              |
| `listReadinessAssessments`   | List all readiness assessments                                       |
| `getReadinessAssessment`     | Get full details of a readiness assessment by ID                     |
| `deleteReadinessAssessment`  | Delete a readiness assessment by ID                                  |

In clients that support [MCP Apps](https://modelcontextprotocol.io/extensions/apps/overview) (Claude Desktop, claude.ai), `startReadinessAssessment` renders an interactive ON2IT-branded questionnaire panel: answer each question with a current and ambition maturity level, review, submit, and get an instant gap analysis. In other clients the same tool falls back to a conversational interview.

The `run-readiness-assessment` MCP prompt turns the assistant into an interviewer that guides you through the assessment (interactive panel where supported; otherwise guided question-by-question, or a quick draft from a free-form description of your organization) and submits the result after your confirmation.

### Case Management

| Tool                       | Description                                    |
| -------------------------- | ---------------------------------------------- |
| `createCase`               | Create a new support case/ticket               |
| `getCases`                 | List all cases/tickets                         |
| `getCase`                  | Get case details by ID                         |
| `updateCasePriority`       | Update case priority (1-4, where 1 is highest) |
| `updateCasePrimaryContact` | Update the primary contact email               |
| `updateCaseSubject`        | Update the case subject                        |
| `escalateCase`             | Escalate to higher priority attention          |
| `deescalateCase`           | De-escalate to normal handling                 |
| `addNoteToCase`            | Add a note/comment to a case                   |
| `closeCase`                | Request to close a case                        |

## Example Prompts

Some prompts to get started:

- *"List all protect surfaces and show me which ones are in Zero Trust focus."*
- *"Run a Zero Trust readiness assessment for my organization."*
- *"Create a priority 3 information request case titled 'Firewall rule review' for jane.doe@example.com, asking ON2IT to review the outbound rules on the DMZ segment."*

## Privacy

The server communicates only with the configured AUXO API endpoint (`api.on2it.net` by default) and sends only the data needed to execute the requested tool call. It does not collect telemetry, does not read conversation history, and stores nothing locally; API tokens are handled by your MCP client (Claude Desktop stores them in the OS keychain). See the [ON2IT Privacy & Cookie Notice](https://on2it.net/privacy/) for how ON2IT processes data on its platform.

## License

See [LICENSE.md](LICENSE.md).
