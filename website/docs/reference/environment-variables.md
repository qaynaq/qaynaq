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
| `AUTH_OAUTH2_AUTHORIZATION_URL` | string | — | OAuth2 authorization endpoint |
| `AUTH_OAUTH2_TOKEN_URL` | string | — | OAuth2 token endpoint |
| `AUTH_OAUTH2_REDIRECT_URL` | string | — | OAuth2 redirect URL (e.g., `http://localhost:8080/auth/callback`) |
| `AUTH_OAUTH2_SCOPES` | string | — | Comma-separated OAuth2 scopes |
| `AUTH_OAUTH2_USER_INFO_URL` | string | — | OAuth2 user info endpoint |
| `AUTH_OAUTH2_ALLOWED_USERS` | string | — | Comma-separated allowed email addresses |
| `AUTH_OAUTH2_ALLOWED_DOMAINS` | string | — | Comma-separated allowed email domains |
| `AUTH_OAUTH2_SESSION_COOKIE_NAME` | string | `qaynaq_session` | Session cookie name |

See [Authentication](/docs/getting-started/authentication) for setup instructions.

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
