# HTTP Server

Accepts incoming HTTP requests — ideal for webhooks and event-driven data.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Path | string | `/post` | Endpoint path to listen on |
| Allowed Verbs | array | `POST` | HTTP methods to accept |
| Timeout | string | `5s` | Request timeout |
| Sync Response | object | — | Customize synchronous response |

:::tip
When using HTTP Server input, pair it with the [Sync Response](/docs/components/outputs/sync-response) output to return custom responses to the caller.
:::
