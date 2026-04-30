---
sidebar_position: 4
---

# MCP OAuth

Qaynaq can act as an OAuth 2.1 Authorization Server for MCP clients, following the [MCP authorization spec](https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization). This lets clients like Claude Desktop and Cursor authenticate users through the same identity provider that protects the rest of the application, instead of pasting a static API token.

When enabled, Qaynaq:

- Publishes OAuth metadata at `/.well-known/oauth-authorization-server` and `/.well-known/oauth-protected-resource`.
- Accepts dynamic client registration at `/mcp/oauth/register` (RFC 7591), so each MCP client provisions its own credentials.
- Delegates the actual end-user login to your existing app authentication. The user signs in once via your IdP (Okta, Auth0, Keycloak, Google, ...) and the same session is reused for MCP authorization.
- Issues short-lived JWT access tokens (1 hour) and rotating refresh tokens (30 days) that are scoped to the MCP endpoint.

Static API tokens continue to work side by side, so headless and CI clients are unaffected.

## Prerequisites

You need application authentication enabled (`AUTH_TYPE=oauth2` is recommended). With `AUTH_TYPE=none`, MCP clients can still complete the flow but every authorization is bound to the same anonymous identity.

See [Authentication](/docs/getting-started/authentication) for IdP setup.

## Enabling

Set:

```bash
MCP_OAUTH_ENABLED=true
```

Restart the coordinator. The new endpoints are mounted alongside `/mcp`.

If you toggle MCP authentication on in **Settings > Authentication**, requests without a valid token receive a `401` with a `WWW-Authenticate` header pointing to the metadata document. MCP clients use that header to discover the authorization server automatically.

## Connecting Clients

The MCP spec requires clients to drive the OAuth flow themselves. Most modern clients support it natively. The first time a user connects, a browser window opens for sign-in. After that, refresh tokens keep the session alive without further prompts.

### What the user sees

1. The MCP client opens the browser at Qaynaq's `/mcp/oauth/authorize` endpoint.
2. If the user is not signed in to Qaynaq, they're redirected through the configured login flow first.
3. The first time this MCP client is used, Qaynaq's web UI renders a **consent page** listing the client name and the permissions it's asking for. The user clicks **Allow** to continue or **Cancel** to abort.
4. On allow, Qaynaq records the consent and redirects back to the MCP client's local callback. Subsequent connections (and refresh-token rotations) skip the consent page.
5. To reset consent, click **Revoke consent** next to the client in **Settings > OAuth Clients**: the next connection prompts again. (Refresh tokens are revoked at the same time so the client immediately sees `invalid_grant` and walks the flow from scratch.)

### Claude Desktop / Claude Code

Add the connector with the bare `/mcp` URL. No token is required:

```bash
claude mcp add qaynaq -- npx mcp-remote http://localhost:8080/mcp
```

`mcp-remote` opens the browser, walks the user through the IdP login, and stores the issued tokens locally.

### Cursor

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

### Claude Desktop "Advanced settings" fields

Claude Desktop's "Add custom connector" dialog has optional **OAuth Client ID** and **OAuth Client Secret** fields under Advanced settings. **Leave both empty.** Qaynaq supports [Dynamic Client Registration](https://datatracker.ietf.org/doc/html/rfc7591), so Claude registers itself automatically the first time it connects. The Advanced fields are only needed for MCP servers that do not support DCR.

## Managing Sessions and Clients

Two pages under **Settings** in the web UI cover the OAuth lifecycle:

### Settings > Sessions

One row per active refresh token. Each row shows the client name, the user's email, when the session started, and when the refresh token expires. Click the trash icon to **revoke a session**: the refresh token is invalidated immediately, and the user is asked to log in again the next time their access token expires (within the next hour). The client registration is kept, so the next login reuses the same client_id without going through registration again.

### Settings > OAuth Clients

One row per registered MCP client (Claude Desktop, Cursor, etc.). Each row shows the client_id, when it registered, when it was last used, and whether the user has consented.

Two actions:

- **Revoke consent** (only shown when the user has consented). Drops the consent row and all refresh tokens for the client. The registration is kept; the next connection prompts for consent again. Use this to reset a user's session without forcing the client to re-register.
- **Delete client** (trash icon). Removes the registration entirely along with consents and refresh tokens. The next connection runs the full OAuth flow from scratch including dynamic client registration.

:::warning Quit the MCP client before deleting
Some MCP clients (notably `mcp-remote` / Claude Code CLI) cache the deleted `client_id` on disk and keep retrying with it, producing a reconnection loop. Before deleting a client, quit it on the user's machine, then delete it here. For routine sign-out without this risk, use **Revoke consent** or **Settings > Sessions** instead.
:::

Use **Sessions** for routine "sign someone out" actions; **Revoke consent** when you also want to force the consent prompt again; **Delete client** for cleaning up unused or stale registrations.

## When to Use OAuth vs Static Tokens

Pick **OAuth** when:

- You run an IdP (Okta, Auth0, Keycloak, Google) and want one identity per user.
- You want per-user revocation and audit, not per-token.
- Your MCP clients are interactive (Claude Desktop, Cursor, IDE extensions).

Stick with **static tokens** when:

- The caller is non-interactive (CI, scripts, internal automations) and cannot complete a browser flow.
- You run Qaynaq without an IdP and want zero-config access.

Both mechanisms are accepted simultaneously. You can mix and match per client.

## Troubleshooting

**Client gets `401` without a metadata pointer.** Make sure `MCP_OAUTH_ENABLED=true` is set on the coordinator process and the endpoint is reachable on the host the client is talking to.

**`redirect_uri does not match any registered URI`.** Some clients regenerate their redirect URI between runs. Either delete the existing client in **Settings > OAuth Clients** so it re-registers, or pin a stable redirect on the client side.

**`PKCE verification failed`.** Older MCP clients omit PKCE. Upgrade the client - the spec mandates PKCE for OAuth flows.

**Stale MCP client credentials page.** The MCP client (commonly `mcp-remote`) is presenting a `client_id` Qaynaq has never seen. Usually means the client cached registration data from a previous database. Clear the local cache and reconnect:

```bash
pkill -f mcp-remote
rm -rf ~/.mcp-auth/mcp-remote-*
```

**Browser keeps reopening "ECONNREFUSED" on `localhost:<port>`.** The MCP client process died but its cached callback port survived. Same recovery as above.
