# MCP Tool

Exposes a flow as a tool via the [Model Context Protocol](https://modelcontextprotocol.io/) (MCP). AI assistants like Claude Desktop, Claude Code, Cursor, and other MCP-compatible clients can discover and call your flow as a tool.

The coordinator exposes a single MCP endpoint at `/mcp` using the Streamable HTTP transport. All MCP Tool flows are registered as tools on this endpoint and automatically synced every 5 seconds.

For connecting AI clients, authentication setup, and testing, see the [MCP Server guide](/docs/guides/mcp-server). For a step-by-step example, see the [Build a Weather Tool for AI Assistants](/playbooks/mcp-weather-tool) playbook.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Tool name that AI clients see (required) |
| Description | string | Human-readable description of what the tool does (required) |
| Input Parameters | property list | Parameters the tool accepts - each with a name, type, description, and required flag (required) |
| Annotations | object | Optional behavioral hints (see below) |

The output **must** be [Sync Response](/docs/components/outputs/sync-response) - this is enforced automatically in the UI. The processed message is returned as the tool result to the AI client.

## Annotations

Annotations are advisory hints clients use to decide how to handle a call - typically whether to auto-approve it or prompt the user first. They are **not** a security boundary: clients should never trust them from untrusted servers. Defaults follow the [MCP specification](https://modelcontextprotocol.io/) so leaving them at their default values is a safe, conservative starting point.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Display Title | string | (empty) | Human-readable title shown by clients in place of the tool name |
| Read-Only | boolean | false | Tool does not modify its environment. Clients may auto-approve calls |
| Destructive | boolean | true | Tool may perform destructive updates. Only meaningful when Read-Only is false. Clients typically prompt before destructive calls |
| Idempotent | boolean | false | Repeated calls with the same arguments have no additional effect. Only meaningful when Read-Only is false |
| Open World | boolean | true | Tool interacts with external entities outside the server's control (e.g. web search). Disable for closed-domain tools like memory or local file access |

:::tip Publishing to a marketplace?
If you plan to list your server in a public marketplace like the Claude marketplace, set safety annotations on every tool: `Read-Only` for read operations, and `Destructive` for create/update/delete operations. Missing tool annotations are the single most common reason for marketplace rejection (~30% of submissions).
:::

## Error Handling

By default, successful tool executions return with a 200 status code. To signal errors or different HTTP status codes (like 404 for "not found" or 400 for "bad request"), set the `meta status_code` field in your flow:

```
meta status_code = 404
```

This is useful when you want to return semantic error codes to the AI client, such as:
- `404` when a requested resource (user, document, etc.) is not found
- `400` for invalid input parameters
- `403` for permission denied
- `500` for internal errors

The AI client will receive both the status code and your response message, allowing it to handle different error conditions appropriately.

:::tip
Write clear, specific descriptions for both the tool and its parameters. AI assistants use these descriptions to decide when and how to call your tool.
:::

## Tool Packs

For supported services, you can skip manual tool creation entirely. [Templates](/docs/guides/templates) let you deploy a full set of pre-built tools in one step - configure shared credentials, select the tools you need, and deploy them all at once.
