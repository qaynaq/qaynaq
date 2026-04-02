# HTTP Client

Performs an HTTP request using the message as the request body, and replaces the message with the body of the response.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | The URL to send requests to (required) |
| Verb | select | `POST` | HTTP method: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD` |
| Headers | map | — | HTTP headers to include in the request |
| Payload | bloblang | — | Optional Bloblang mapping to construct the request body. When empty, the raw message payload is sent |
| Timeout | string | `5s` | Request timeout |
| Retry Period | string | `1s` | Delay between retries |
| Max Retry Backoff | string | `300s` | Maximum backoff duration between retries |
| Retries | integer | `3` | Number of retries on failure |
| Backoff On | array | `[429]` | HTTP status codes that trigger a retry with backoff |
| Drop On | array | — | HTTP status codes that cause the message to be dropped without retrying |
| Successful On | array | — | HTTP status codes to treat as successful (in addition to 2xx) |
| Rate Limit | string | — | Name of a rate limit resource to apply to outgoing requests |
| Proxy URL | string | — | Optional HTTP proxy URL |
| Batch as Multipart | boolean | `false` | Send batched messages as a single multipart HTTP request |
| Parallel | boolean | `false` | When processing a batch, send requests in parallel rather than sequentially |
| Extract Headers — Include Prefixes | array | — | Forward response headers whose names start with any of these prefixes as message metadata |
| Extract Headers — Include Patterns | array | — | Forward response headers matching any of these regex patterns as message metadata |
| Metadata — Include Prefixes | array | — | Attach message metadata fields whose keys start with any of these prefixes as request headers |
| Metadata — Include Patterns | array | — | Attach message metadata fields matching any of these regex patterns as request headers |

## Authentication

The HTTP Client processor supports four authentication methods. Enable at most one per processor.

**Basic Auth** — Sends a username and password via the HTTP `Authorization` header.

**OAuth** — Signs requests with OAuth 1.0 using a consumer key/secret and access token/secret.

**OAuth2** — Obtains a bearer token from the configured Token URL using the client credentials flow, then attaches it to each request.

**JWT** — Signs requests with a JWT using a private key file and the specified signing method. Custom claims and headers can be added to the token.

## Payload Mapping

By default, the raw message payload is sent as the request body. Use the Payload field to transform it first with a Bloblang mapping — `root` is the outgoing body and `this` refers to the current message.

## Response Handling

The processor replaces the message content with the HTTP response body. If the response status code is not in the 2xx range (or in Successful On), the message is marked as failed and the retry logic applies.

Response headers can be extracted into message metadata using the Extract Headers fields, which is useful when downstream processors need values like `Content-Type` or custom headers returned by the API.

## Retries and Backoff

On failure, the processor waits for Retry Period before attempting again, up to the number of Retries. Backoff On status codes (default: `429`) trigger exponential backoff up to Max Retry Backoff. Drop On status codes discard the message immediately without retrying.

:::tip
Use the Payload field with a Bloblang mapping to reshape your message before sending — for example, to wrap the payload in a JSON envelope or extract only the fields the API expects.
:::
