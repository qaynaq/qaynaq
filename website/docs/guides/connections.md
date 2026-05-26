---
sidebar_position: 6
---

# Connections (OAuth Providers)

Qaynaq Connections store OAuth credentials for third-party services. Once a connection
is authorized, you can attach it to MCP server entries or any component that needs an
access token. The coordinator handles refresh centrally so your workers never see
refresh tokens.

## Supported providers

| Provider | Slug | Notes |
|---|---|---|
| Google | `google` | Google Workspace APIs (Calendar, Drive, Gmail, Sheets, Docs, Slides) |
| Slack | `slack` | User tokens (`user_scope`); supports rotation |
| Slack MCP | `slack_mcp` | Official Slack remote MCP (`https://mcp.slack.com/mcp`). PKCE required. |
| Asana | `asana` | Regular Asana API |
| Asana MCP | `asana_mcp` | Asana's MCP server. Scopeless. Uses `resource=https://mcp.asana.com/v2`. |
| Atlassian (Jira & Confluence) | `atlassian` | OAuth 2.0 (3LO). Requires Cloud ID. |
| GitHub | `github` | Classic OAuth App. Tokens don't expire (no refresh). |
| GitHub MCP | `github_mcp` | OAuth 2.1 + PKCE. |
| HubSpot | `hubspot` | 30-min access tokens, rotating refresh. |
| Linear | `linear` | OAuth 2.1, `actor=user`. |
| Notion | `notion` | OAuth 2.1. Hosted MCP at `https://mcp.notion.com/mcp`. |
| Sentry | `sentry` | PKCE recommended. |
| Shopify | `shopify` | Per-shop URL. PKCE required. |

## Creating a connection

1. Open the **Connections** page in the Qaynaq UI.
2. Click **New Connection**.
3. Pick a provider, enter a connection name, paste OAuth Client ID and Client Secret.
4. For Shopify: enter the shop subdomain (e.g., `my-store` for `my-store.myshopify.com`).
5. For Atlassian: enter the Cloud ID (UUID — see provider notes below for how to find it).
6. Pick scopes. Asana MCP has none — that's expected.
7. Click **Authorize**. A popup walks you through the provider's consent flow.

The redirect URI to register in your OAuth app is shown in the form:
`https://YOUR-QAYNAQ-HOST/connections/oauth/callback`

## Per-provider setup notes

### Google

Create an OAuth Client ID at [Google Cloud Console](https://console.cloud.google.com/apis/credentials).
Application type: **Web application**. Refresh tokens are issued via
`access_type=offline` + `prompt=consent`.

### Slack

Create an app at [api.slack.com/apps](https://api.slack.com/apps). Use the
**Basic Information** page for Client ID / Client Secret. Qaynaq uses `user_scope=`
to issue user tokens (not bot tokens). Token rotation is supported automatically when
your app has it enabled.

### Slack MCP

Slack publishes an official remote MCP server at `https://mcp.slack.com/mcp`. The
Qaynaq `slack_mcp` provider connects to it via OAuth 2.1 + PKCE — a different
endpoint pair than regular `slack`:

- Authorize: `https://slack.com/oauth/v2_user/authorize`
- Token: `https://slack.com/api/oauth.v2.user.access`

Use the same Slack App you create at [api.slack.com/apps](https://api.slack.com/apps).
The user scopes you check in Qaynaq must also be requested in the app's OAuth & Permissions
page (under **User Token Scopes**). The MCP-published scope set includes Slack canvases
and a richer search surface (`search:read.users`, `search:read.mpim`, `search:read.im`).

Unlike regular `slack`, the MCP endpoint returns access tokens at the top level of the
OAuth response — no `authed_user` nesting — so this provider does not need any special
token-extraction handling.

PKCE is required by the OAuth 2.1 spec and Slack's MCP metadata. Qaynaq generates the
S256 challenge automatically.

### Asana (regular)

Create an app at [Asana Developer Console](https://app.asana.com/0/my-apps). Use the
default `default` scope for full access to user resources. Refresh tokens rotate on
every refresh — Qaynaq handles this automatically.

### Asana MCP

The MCP-enabled Asana app authorizes against the Asana MCP server at
`https://mcp.asana.com/v2`. Notes:

- Set the app to **MCP-enabled** in the Asana developer console.
- Configure workspace distribution to either "Specific workspaces" (dev) or "Any workspace".
- Asana MCP rejects any `scope=` parameter — Qaynaq sends the authorize URL without scopes.
- Tokens are valid for 1 hour and refresh automatically via the central refresh path.

The V1 Beta endpoint (`https://mcp.asana.com/sse`) was deprecated on 2026-05-11.
This integration uses V2 only.

### Atlassian (Jira & Confluence)

Create an OAuth 2.0 (3LO) integration at
[Atlassian Developer Console](https://developer.atlassian.com/console/myapps/).
Add the Jira and/or Confluence APIs.

**Cloud ID** — required to construct API URLs like `https://api.atlassian.com/ex/jira/{cloudid}/...`:

- Find it by visiting `https://YOUR-SITE.atlassian.net/_edge/tenant_info` in your browser.
  The response is JSON with a `cloudId` field. Copy the UUID.
- If you have access to multiple Atlassian sites, enter whichever site's Cloud ID this
  connection should address. To use a different site, create a separate connection.

**`offline_access` scope** is required for refresh tokens. Without it, Atlassian issues
a 1-hour token with no way to refresh, and the connection will fail after expiry.

### GitHub (OAuth App)

Create an OAuth App at [github.com/settings/developers](https://github.com/settings/developers).
This is the **OAuth App** flow, not the GitHub App flow. OAuth App tokens don't expire
upstream, so Qaynaq treats them as non-refreshing — there's no refresh endpoint to call.

### GitHub MCP

GitHub's official remote MCP server uses OAuth 2.1 + PKCE. Same OAuth App registration
as `github`, but Qaynaq automatically generates a PKCE `code_verifier` for each authorize
flow and sends `code_challenge_method=S256`. No setup difference for the user.

### HubSpot

Create a public app at [HubSpot Developer Portal](https://developers.hubspot.com/).
Access tokens are short (30 minutes); the central refresh path handles rotation
transparently. Token endpoint uses HubSpot's 2026-03 API.

### Linear

Create an OAuth application at
[Linear Application Settings](https://linear.app/settings/api/applications/new).
Qaynaq sends `actor=user` so the issued token represents the authorizing user (vs.
an app actor token, which would require the client_credentials grant).

For long-lived API key access without OAuth refresh, you can also configure Linear's
remote MCP server in **MCP Servers → Add Server** and use `auth_type=token` with a
Linear API key.

### Notion

Create an integration at [notion.so/my-integrations](https://www.notion.so/my-integrations).
Make it a **public integration** with OAuth enabled. Notion's permissions are configured
in the Notion UI, not via OAuth scopes — the scopes shown in Qaynaq's picker are
informational.

### Sentry

Create an OAuth application at
[Sentry → API Applications](https://sentry.io/settings/account/api/applications/).
PKCE is enabled automatically.

### Shopify

Create an app in the [Shopify Partner Dashboard](https://partners.shopify.com/).
**Shop domain** — Qaynaq's authorize URL is per-shop:
`https://{shop}.myshopify.com/admin/oauth/authorize`. Enter the shop subdomain
(e.g., `my-store` for `my-store.myshopify.com`) when creating the connection.
PKCE is enabled automatically.

The shop subdomain is validated against `^[a-z0-9][a-z0-9-]*$` to prevent URL
injection. Uppercase, dots, and special characters are rejected.

## Refresh handling

Refresh runs in the coordinator. Workers fetch a fresh access token via gRPC every
time they need one. The refresh path:

- **Standard providers** (Slack, Asana, Atlassian, HubSpot, Linear, Notion, Sentry,
  Shopify, Google): refresh tokens rotate on every refresh; the new refresh token
  replaces the old.
- **Google special case**: Google sometimes omits the refresh token on subsequent
  refreshes. Qaynaq preserves the original refresh token in that case.
- **GitHub OAuth App**: tokens don't expire upstream. Qaynaq skips the refresh
  call entirely; the stored access token is returned as-is until the user revokes
  the app or rotates their secret.
- **Slack non-rotating apps**: tokens with no expiry and no refresh token are treated
  as non-expiring (returned as-is until revoked).

## Health and recovery

When a refresh fails, the connection row records the upstream error, the time
it started failing, and a consecutive-failure counter. The connections page
shows a red **Failing** badge with a tooltip describing the error, and the
dashboard renders an alert listing every broken connection.

After three consecutive failures the refresh job backs off to one retry per
hour instead of every five minutes. Some OAuth providers revoke refresh tokens
under repeated failed exchanges, and a broken connection almost always needs a
user-driven re-authorize to recover anyway. The badge switches to
**Re-authorize** once the connection is in backoff.

The error state clears automatically on the next successful refresh, or
immediately when you re-authorize.

## Re-authorization

If a connection's tokens become invalid (revoked, secret rotated upstream, scope
changed), use the refresh icon on the connection row to re-authorize. The flow
preserves the connection name and provider; you can also update the Client ID,
Client Secret, and scopes during re-auth.

## Adding a provider Qaynaq doesn't support yet

The 13 providers above cover the most common remote MCP services as of May 2026. If
you need a provider not on the list (Keycloak, GitLab, custom OIDC, etc.), open an
issue with the provider's auth/token URLs and any quirks (scope mode, special URL
parameters, refresh behavior). Most standard OAuth 2.0/2.1 providers can be added
in roughly 30 LOC.
