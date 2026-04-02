# HTTP Client

Sends messages to an HTTP endpoint.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | The URL to send requests to |
| Verb | string | `POST` | HTTP method |
| Headers | map | — | HTTP headers |
| Timeout | string | `5s` | Request timeout |
| Retries | integer | `3` | Number of retries |
| Max In Flight | integer | `64` | Maximum parallel requests |
| Rate Limit | string | — | Rate limit resource name |
| Batching | object | — | Batching policy |

Supports authentication: **Basic Auth**, **OAuth**, **OAuth2**, and **JWT**.
