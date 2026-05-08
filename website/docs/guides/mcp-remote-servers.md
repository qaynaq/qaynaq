---
sidebar_position: 4
---

# Remote MCP Servers (HTTP)

Qaynaq's `/mcp` endpoint can serve tools from your own flows ([MCP Server](/docs/guides/mcp-server)) AND from external MCP servers you register, all in the same connection. Register an external server in **MCP Servers** and within ~30 seconds its tools appear alongside your native tools, namespaced by server name.

Qaynaq supports two flavors of upstream MCP server: **remote** (HTTP URL, this guide) and **local** (CLI binaries Qaynaq runs as a child process, see [Local MCP Servers](/docs/guides/mcp-local-servers)). Both show up in the same MCP Servers list and route through the same `/mcp` endpoint.

Use this guide when an upstream service already publishes an HTTP MCP server and you want a single endpoint your AI assistants can connect to instead of one per service.

## When to Use the Proxy

- You want one `/mcp` connection in Claude Desktop / Cursor / Claude Code that exposes both your flows and a third-party MCP server (e.g., a vendor's hosted MCP, an internal team's MCP, GitHub's MCP).
- You want to put one of your existing OAuth Connections behind a third-party MCP server so token rotation is handled centrally.
- You want a static API key shared across multiple AI clients without baking it into each client's config.

If you only need to expose your own flows, see [MCP Server](/docs/guides/mcp-server).

## Setup

### 1. Open MCP Servers

Click **MCP Servers** in the sidebar (under the same group as Connections). The page lists every registered server with status badge, tool count, last sync time, and any last error.

### 2. Add a Server

Click **Add Server** and fill in the dialog:

| Field | Description |
|-------|-------------|
| **Name** | A short identifier for this server. Used to namespace tool names: a tool `send_message` from a server named `slack` shows up as `slack__send_message`. Keep it short and stable - changing it later breaks any client config that references the namespaced name. |
| **Server URL** | The upstream MCP server's HTTP endpoint, e.g. `https://mcp-server.example.com/mcp`. Streamable HTTP only. |
| **Authentication** | One of `No authentication`, `Bearer token / API key`, or `OAuth connection` (only shown if you have at least one OAuth connection set up). |

Then save. Within the next sync tick (≤30s) the server's tools appear on `/mcp`.

### 3. Authentication Modes

#### No authentication

Used for public or LAN-only MCP servers that don't require auth. No additional fields.

#### Bearer token / API key

Used for static credentials (long-lived API keys, personal access tokens). The token is encrypted at rest with the same key as the rest of Qaynaq.

| Field | Description |
|-------|-------------|
| **Token** | The credential value. Stored encrypted; never displayed back after save. |
| **Custom Header** | Optional. By default the value is sent as `Authorization: Bearer <token>`. Set this if the upstream expects a different header name (e.g., `X-API-Key`). |

#### OAuth connection

Used for upstream MCP servers that authenticate against a provider you've already connected to in **Connections** (Google, Slack, etc.). The proxy fetches the access token from your existing connection on every request, refreshes it automatically when needed, and force-refreshes + retries on a 401.

| Field | Description |
|-------|-------------|
| **Connection** | Pick one of your existing OAuth connections from the dropdown. The proxy will use that connection's access token as `Authorization: Bearer <token>` for every request to this MCP server. |

This is the right choice when the upstream MCP server is from a provider you've already gone through the OAuth dance for. Token refresh is fully automatic - if the upstream returns 401 because the token rotated upstream, the proxy force-refreshes and retries the same call without you noticing.

## Tool Namespacing and Collisions

Every upstream tool is exposed as `<server-name>__<tool-name>`. So `slack` server's `send_message` becomes `slack__send_message` on `/mcp`.

Native tools (your own flows) win on collisions. If you have a flow whose MCP tool name is `slack__send_message`, the upstream server's tool with the same namespaced name is dropped from the listing and a warning is logged. This keeps your locally-defined behavior authoritative.

## Recovery and Reliability

The proxy keeps each upstream server's MCP session open and reuses it across sync ticks. Two recovery mechanisms run automatically:

- **In-transport 401 retry** (OAuth connection auth only): if the upstream returns 401 on a tool call or list-tools, the proxy force-refreshes the connection's access token, replaces the Authorization header, and retries the same request once. This is silent - the calling AI client sees the successful response.
- **Circuit breaker** (all auth modes): if a server fails 3 syncs in a row (network error, 5xx, repeated 401 even after refresh), it goes into a 5-minute cooldown. The status badge flips to red and `last_error` is shown. After the cooldown elapses, one attempt is allowed through; success clears the breaker and flips the status back to active automatically, another failure restarts the cooldown. The breaker survives coordinator restarts: an errored server is still picked up by the next sync tick. To force an immediate retry, click the **Restart** icon on the server's row (or edit the server - any update also resets the breaker).

If a server stays errored: check the `last_error` shown on the server's row. Common causes:

- **`unauthorized (401)`** with `OAuth connection` auth: the connection's refresh token is invalid (user revoked access, expired beyond the rotation window). Re-authorize the connection in **Connections**, then click **Restart** on the MCP server row to retry immediately.
- **`unauthorized (401)`** with `Bearer token` auth: the static token was revoked or rotated upstream. Edit the server and paste a new token.
- **`connection refused`** / **`no such host`**: the URL is unreachable from the coordinator. Check connectivity.
- **`session terminated`**: the upstream rotated its session ID. Should auto-recover on the next sync.

## Trust Model

A few things worth knowing before you put this in front of a multi-tenant or internet-exposed Qaynaq:

- **Server URLs are admin-trusted.** Any user with access to the **MCP Servers** page can point a server at any URL the coordinator can reach, including internal addresses. There is no SSRF allowlist or blocklist today. If your Qaynaq deployment is single-tenant and every UI user is already a trusted admin, this is fine. For shared deployments, treat MCP server creation as an administrative action.
- **Upstream tools run with the configured auth.** If you point a server at an upstream you don't fully trust, that upstream sees whatever credential you configured. Use a scoped token whenever possible.
- **JSON responses are size-capped at 10 MiB** to prevent a malicious or buggy upstream from OOMing the coordinator with a multi-GB response. Server-Sent Event streams are not capped (they're long-lived by design); per-event payloads are bounded by mcp-go's internal decoder.
- **Static tokens are encrypted at rest** with Qaynaq's secret key. The plaintext value is held in coordinator memory while a server is registered (needed to pass to the upstream); this is the same trust model as Connections.

## Verifying

After adding a server, watch the row's status badge. It should flip to green within 30 seconds with a non-zero tool count. Connect from your AI client - the upstream tools appear under their namespaced names alongside your native tools.

You can also test the local `/mcp` endpoint with curl (see [MCP Server > Verifying](/docs/guides/mcp-server#verifying)) - the `tools/list` response should include the namespaced upstream tools.

## See Also

- [MCP Server](/docs/guides/mcp-server) - the local `/mcp` endpoint serving your flow-defined tools.
- [Connections](/docs/getting-started/authentication) - OAuth connections used by the `OAuth connection` auth mode.
- [MCP OAuth](/docs/guides/mcp-oauth) - using OAuth for clients connecting TO Qaynaq's `/mcp` endpoint (the inverse direction).
