---
sidebar_position: 3
---

# MCP Server

Qaynaq's coordinator includes a built-in [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server at `/mcp`. Any flow with the [MCP Tool](/docs/components/inputs/mcp-tool) input becomes a tool that AI assistants can discover and call. Tools sync automatically every 5 seconds.

## Connecting Clients

The endpoint is available at `http://<coordinator-host>:<port>/mcp`. Clients that only support stdio transport can use [`mcp-remote`](https://www.npmjs.com/package/mcp-remote) as a bridge.

### Claude Code

```bash
claude mcp add qaynaq -- npx mcp-remote http://localhost:8080/mcp
```

### Claude Desktop

Add in **Settings > Connectors**, or add to your configuration file:

```json
{
  "mcpServers": {
    "qaynaq": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp"]
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json` in your project:

```json
{
  "mcpServers": {
    "qaynaq": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp"]
    }
  }
}
```

## Authentication

The MCP endpoint can be protected with API tokens, managed through the **Settings** page in the web UI.

:::warning
Token management requires application authentication to be enabled first (basic or OAuth2). See [Authentication](/docs/getting-started/authentication).
:::

### Setup

1. Navigate to **Settings** in the web UI.
2. Toggle **Require authentication** under MCP Authentication.
3. Click **Create Token** and give it a name (e.g., "Claude Desktop").
4. Copy the token - it is only shown once.

### Passing the Token

Append the token as a query parameter:

```bash
claude mcp add qaynaq -- npx mcp-remote "http://localhost:8080/mcp?token=at_your_token_here"
```

Or in a JSON configuration:

```json
{
  "mcpServers": {
    "qaynaq": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp?token=at_your_token_here"]
    }
  }
}
```

Clients that support custom headers can use the Authorization header instead:

```
Authorization: Bearer at_your_token_here
```

:::tip
Create a separate token for each client. You can revoke access per client and track usage via the last-used timestamp in Settings.
:::

## Verifying

Test the endpoint with curl:

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
```

If authentication is enabled, add the token:

```bash
curl -X POST "http://localhost:8080/mcp?token=at_your_token_here" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
```

## Multiple Tools

Each MCP Tool flow becomes a separate tool on the same `/mcp` endpoint. Create as many flows as needed - they all appear together when a client connects.

To quickly deploy a full set of tools for a supported service, use [MCP Tool Packs](/docs/guides/mcp-tool-packs). Packs let you configure shared credentials once and create multiple tools at once.

For a step-by-step example, see the [Build a Weather Tool for AI Assistants](/playbooks/mcp-weather-tool) playbook.
