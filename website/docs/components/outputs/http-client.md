# HTTP Client

Sends messages to an HTTP endpoint.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | The URL to send requests to |
| Verb | string | `POST` | HTTP method |
| OAuth Connection | connection | | Authenticate requests with a connection from Settings > Connections |
| Headers | map | — | HTTP headers |
| Timeout | string | `5s` | Request timeout |
| Retries | integer | `3` | Number of retries |
| Max In Flight | integer | `64` | Maximum parallel requests |
| Rate Limit | string | — | Rate limit resource name |
| Batching | object | — | Batching policy |

The recommended way to authenticate is with an **OAuth Connection**: pick a connection created in [Settings > Connections](/docs/guides/connections) and Qaynaq attaches a fresh access token to every request, refreshing it automatically before it expires. The connection takes precedence over a manually set `Authorization` header.

Manual authentication is also supported: **Basic Auth**, **OAuth**, **OAuth2**, and **JWT**.
