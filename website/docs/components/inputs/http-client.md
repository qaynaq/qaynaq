# HTTP Client

Pulls data from an HTTP endpoint by making requests at a configured interval.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | The URL to send requests to |
| Verb | string | `GET` | HTTP method (GET, POST, PUT, DELETE) |
| OAuth Connection | connection | | Authenticate requests with a connection from Settings > Connections |
| Headers | map | — | HTTP headers to include |
| Timeout | string | `5s` | Request timeout |
| Retry Period | string | `1s` | Delay between retries |
| Retries | integer | `3` | Number of retries on failure |
| Rate Limit | string | — | Rate limit resource name |

The recommended way to authenticate is with an **OAuth Connection**: pick a connection created in [Settings > Connections](/docs/guides/connections) and Qaynaq attaches a fresh access token to every request, refreshing it automatically before it expires. The connection takes precedence over a manually set `Authorization` header.

Manual authentication is also supported: **Basic Auth**, **OAuth**, **OAuth2**, and **JWT**.
