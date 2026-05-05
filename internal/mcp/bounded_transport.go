package mcp

import (
	"net/http"
	"strings"
)

// maxUpstreamJSONResponse caps the size of a unary JSON-RPC response body
// from an upstream MCP server. Tool definitions and most call results are
// well below this; a multi-GB response is either a bug or an attack. SSE
// streams are NOT bounded here - they're long-lived by design and the
// per-event JSON inside them is bounded separately by the JSON decoder's
// internal buffer.
const maxUpstreamJSONResponse = 10 * 1024 * 1024 // 10 MiB

// boundedJSONTransport wraps a base RoundTripper and limits the response
// body size when the upstream returns Content-Type: application/json. SSE
// (`text/event-stream`) and other content types pass through unchanged.
type boundedJSONTransport struct {
	base http.RoundTripper
}

func (b *boundedJSONTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := b.base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		resp.Body = http.MaxBytesReader(nil, resp.Body, maxUpstreamJSONResponse)
	}
	return resp, nil
}

// withBoundedJSON wraps the given http.Client (or http.DefaultClient) so
// that JSON responses larger than maxUpstreamJSONResponse are truncated and
// produce a read error. The original client is not mutated.
func withBoundedJSON(c *http.Client) *http.Client {
	if c == nil {
		c = http.DefaultClient
	}
	base := c.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone := *c
	clone.Transport = &boundedJSONTransport{base: base}
	return &clone
}
