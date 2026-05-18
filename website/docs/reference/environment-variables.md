---
sidebar_position: 2
---

# Environment Variables

All Qaynaq settings can be configured via environment variables.

## Runtime

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `ROLE` | string | `coordinator` | Node role: `coordinator` or `worker` |
| `GRPC_PORT` | uint | — | gRPC server port (required) |
| `HTTP_PORT` | uint | `8080` | HTTP port for web UI and REST API |
| `DISCOVERY_URI` | string | `localhost:50000` | Coordinator address for worker discovery |
| `DEBUG_MODE` | bool | `false` | Enable debug logging |

## Security

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `SECRET_KEY` | string | — | 32-byte encryption key, also used for signing JWT tokens |

## Authentication

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `AUTH_TYPE` | string | `none` | Authentication mode: `none`, `basic`, or `oauth2` |
| `AUTH_BASIC_USERNAME` | string | — | Username for basic auth |
| `AUTH_BASIC_PASSWORD` | string | — | Password for basic auth |
| `AUTH_OAUTH2_CLIENT_ID` | string | — | OAuth2 client ID |
| `AUTH_OAUTH2_CLIENT_SECRET` | string | — | OAuth2 client secret |
| `AUTH_OAUTH2_ISSUER_URL` | string | — | OIDC issuer URL (e.g. `https://accounts.google.com`, `http://localhost:8090/realms/qaynaq`). Qaynaq discovers the authorization, token, and userinfo endpoints from `<issuer>/.well-known/openid-configuration` at startup. |
| `AUTH_OAUTH2_SCOPES` | string | `openid,email,profile` | Comma-separated OAuth2 scopes |
| `AUTH_OAUTH2_ALLOWED_USERS` | string | — | Comma-separated allowed email addresses |
| `AUTH_OAUTH2_ALLOWED_DOMAINS` | string | — | Comma-separated allowed email domains |
| `AUTH_OAUTH2_SESSION_COOKIE_NAME` | string | `qaynaq_session` | Session cookie name |
| `AUTH_OAUTH2_ADMIN_USERS` | string | — | Comma-separated email patterns granted Admin role (supports `*@x.com`, `sales-*@x.com`). See [Access Control](/docs/guides/access-control). |
| `AUTH_OAUTH2_MCP_USERS` | string | — | Comma-separated email patterns granted MCP-only role. |
| `AUTH_OAUTH2_ROLE_ATTRIBUTE_PATH` | string | — | JMESPath expression against userinfo claims that returns `Admin`, `MCP`, or anything else (no role). |

See [Authentication](/docs/getting-started/authentication) for setup instructions.

## MCP

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `MCP_OAUTH_ENABLED` | bool | `false` | Expose OAuth 2.1 Authorization Server endpoints so MCP clients (Claude Desktop, Cursor, etc.) can authenticate via the standard MCP OAuth flow |

See [MCP OAuth](/docs/guides/mcp-oauth) for setup.

## Database

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DATABASE_DRIVER` | string | — | Database backend: `sqlite` or `postgres` |
| `DATABASE_URI` | string | — | Database connection string |

:::warning
If `DATABASE_DRIVER` and `DATABASE_URI` are not set, Qaynaq stores data in memory. All data is lost when the process stops.
:::

### SQLite

```bash
export DATABASE_DRIVER="sqlite"
export DATABASE_URI="file:./qaynaq.sqlite?_foreign_keys=1&mode=rwc"
```

### PostgreSQL

URL format:

```bash
export DATABASE_DRIVER="postgres"
export DATABASE_URI="postgres://qaynaq:yourpassword@localhost:5432/qaynaq?sslmode=disable"
```

DSN format:

```bash
export DATABASE_DRIVER="postgres"
export DATABASE_URI="host=localhost user=qaynaq password=yourpassword dbname=qaynaq port=5432 sslmode=disable"
```
