---
sidebar_position: 11
---

# Access Control

When Qaynaq runs with OAuth2 authentication, every signed-in user gets full access to the management dashboard and the MCP server by default. This guide explains how to split users into two roles so you can hand teammates MCP-only access without exposing flow editing, secrets, and connections.

## Roles

Qaynaq recognizes two roles when OAuth2 auth is enabled:

- **Admin** - full access to the dashboard, the API at `/api/*`, and the MCP server at `/mcp`.
- **MCP** - sign-in succeeds and the user can complete the OAuth handshake from Claude Desktop, Cursor, or any other MCP client, but they cannot open the dashboard or call the API. After signing in they land on a small page with the MCP server URL and copy-paste snippets for the popular AI clients.

Users who authenticate but match neither role land on an informational "no access" page.

Basic auth is unchanged - the single configured user is always an admin. With `AUTH_TYPE=none` everyone is treated as admin.

## Configuration

Three optional env vars control role mapping:

- `AUTH_OAUTH2_ADMIN_USERS` - comma-separated email patterns granted Admin.
- `AUTH_OAUTH2_MCP_USERS` - comma-separated email patterns granted MCP-only access.
- `AUTH_OAUTH2_ROLE_ATTRIBUTE_PATH` - a [JMESPath](https://jmespath.org/) expression evaluated against the userinfo response your identity provider returns.

If all three are empty (the default), every user who passes the `AUTH_OAUTH2_ALLOWED_USERS` / `AUTH_OAUTH2_ALLOWED_DOMAINS` gate is treated as Admin. The role split is opt-in.

When the role config is set, the precedence is:

1. If `AUTH_OAUTH2_ROLE_ATTRIBUTE_PATH` returns the string `Admin` or `MCP`, that wins.
2. Otherwise the admin / MCP lists are consulted.
3. If neither matches, the user has no role and lands on the no-access page.

`AUTH_OAUTH2_ALLOWED_USERS` and `AUTH_OAUTH2_ALLOWED_DOMAINS` remain the outer gate: a user not in the allowlist is rejected before the role evaluator ever runs.

## Email patterns

Both list env vars accept exact emails and glob patterns:

```bash
# Exact match
AUTH_OAUTH2_ADMIN_USERS=alice@example.com

# Whole domain
AUTH_OAUTH2_MCP_USERS=*@example.com

# Prefix on the local part
AUTH_OAUTH2_MCP_USERS=sales-*@example.com

# Mix freely
AUTH_OAUTH2_ADMIN_USERS=alice@example.com,*@admin.example.com
```

Matching is case-insensitive and uses Go's `path/filepath.Match` semantics (`*` matches any sequence of non-separator characters; emails have no separator).

If the same email matches both lists, Admin wins.

## JMESPath against IdP claims

For identity providers that publish group or role claims in their userinfo response (Keycloak, Okta, Authentik, Auth0 with a custom claim, and so on), `AUTH_OAUTH2_ROLE_ATTRIBUTE_PATH` lets you derive roles from those claims:

```bash
AUTH_OAUTH2_ROLE_ATTRIBUTE_PATH="contains(groups[*], 'qaynaq-admins') && 'Admin' || (contains(groups[*], 'qaynaq-mcp') && 'MCP' || 'None')"
```

The expression must return the literal string `Admin`, `MCP`, or anything else (which Qaynaq treats as "no recognized role").

The expression is evaluated against whatever JSON your IdP returns from its userinfo endpoint. What claims are available is up to your IdP. For Keycloak, expose the `groups` claim by adding a group mapper to your client. For other providers, check what their userinfo endpoint returns and write the expression to match.

If your IdP does not return groups or roles in userinfo (Google, plain GitHub OAuth), use the static lists - they work for any provider.

## How role changes take effect

Role configuration is read from env on every request. If you restart Qaynaq with a different `AUTH_OAUTH2_ADMIN_USERS` value, sessions that were issued before the restart will see the new rules on their next request - no forced re-login.

The userinfo claims themselves are captured once at sign-in and stored in the session JWT. If a user's group membership changes upstream after they sign in, they keep the role implied by their original claims until their session expires (default 24 hours) or they sign out and back in.

## Sign-out

Sign-out clears the Qaynaq session cookie only. Your identity provider session is left alone, since it may be shared with other applications. If a user needs to switch IdP accounts, they sign out at the IdP separately.

## Behavior matrix

| | `/` (dashboard) | `/api/*` | `/mcp/oauth/authorize` | `/mcp` (after authorize) |
|---|---|---|---|---|
| Admin | allowed | allowed | allowed | allowed |
| MCP | redirects to `/mcp-access` | 403 | allowed | allowed |
| No role | redirects to `/no-access` | 403 | redirects to `/no-access` | n/a |
| Not signed in | redirects to `/login` | 401 | redirects to `/auth/login` | 401 |

## Group-based tool access

Roles decide who may reach the MCP endpoint at all. Groups decide which tools each caller sees once they are in. The accountant's assistant lists only the finance tools, the lawyer's only the legal ones, and each restricted tool also rejects calls from members of other groups.

Qaynaq does not manage group membership itself. Your identity provider already knows who is in `accounting` and who is in `legal`; Qaynaq reads that claim and applies it. There is one env var:

- `AUTH_OAUTH2_GROUPS_ATTRIBUTE_PATH` - a JMESPath expression evaluated against the userinfo claims that yields the user's group names, either a string array or a single string. For most providers it is simply `groups`.

```bash
AUTH_OAUTH2_GROUPS_ATTRIBUTE_PATH=groups
```

### Tagging tools and servers

- **Qaynaq tools**: in the flow builder, MCP tool flows get an **Access** picker next to the buffer selector. Pick one or more groups; empty means everyone.
- **External MCP servers**: on the MCP Servers page, each server has an access control (the people icon) that restricts every tool proxied from that server.

Restrictions apply to MCP clients authenticated through the OAuth flow, where Qaynaq knows which person is calling. Static API tokens and instances running without auth are not restricted, since there is no user identity to match against; treat those tokens as admin credentials.

### The groups registry

The **Settings → Groups** page lists every group Qaynaq knows about. It fills itself: each time a user signs in, the groups from their IdP claims are recorded. To tag a tool for a group before its first member has ever signed in, use **Import groups** and paste the names straight from your IdP admin console, one per line.

The Access picker only offers names from this registry, so a typo cannot silently lock a team out of a tool. Groups that have never been seen at a login carry a warning badge, and if a tagged name looks like a near-miss of a real observed group, the picker points that out.

The registry is a convenience for the UI only. Access decisions always use the groups from the caller's identity, never this list, so a stale registry can never grant or deny the wrong access.

### How group changes take effect

Each user's groups are captured from the IdP at sign-in and refreshed on every subsequent sign-in. When someone leaves a group in your IdP, their Qaynaq access updates the next time they sign in to Qaynaq. To cut off a user immediately, revoke their MCP session under **Settings → Sessions**.

## Tool call audit log

**Settings → Tool Calls** records every call made through the MCP endpoint: who called (the signed-in user's email, the API token's name, or anonymous), which tool, whether it succeeded, failed, or was denied by a group restriction, and how long it took. Denied calls are recorded with the caller's groups at the time of the call, so you can reconstruct why a call was rejected weeks later.
