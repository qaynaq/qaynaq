# HTTP Client

Pulls data from an HTTP endpoint by making requests at a configured interval.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | The URL to send requests to |
| Verb | string | `GET` | HTTP method (GET, POST, PUT, DELETE) |
| Headers | map | — | HTTP headers to include |
| Timeout | string | `5s` | Request timeout |
| Retry Period | string | `1s` | Delay between retries |
| Retries | integer | `3` | Number of retries on failure |
| Rate Limit | string | — | Rate limit resource name |

Supports authentication: **Basic Auth**, **OAuth**, **OAuth2**, and **JWT**.
