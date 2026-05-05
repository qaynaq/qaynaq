package mcp

import (
	"context"
	"fmt"
	"net/http"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// ConnectMCPClient opens a streamable-HTTP MCP session. If httpClient is
// non-nil it is used as the underlying transport (e.g. one that injects
// auth headers and retries on 401); otherwise headers (if any) are attached
// statically to every request via mcp-go's default client.
//
// The resulting client always reads JSON responses through a bounded reader
// (see boundedJSONTransport) to stop a malicious upstream from OOMing the
// process with a multi-GB response. SSE streams are not bounded here.
func ConnectMCPClient(ctx context.Context, url string, headers map[string]string, httpClient *http.Client) (*mcpclient.Client, error) {
	if httpClient != nil {
		httpClient = withBoundedJSON(httpClient)
	} else {
		httpClient = withBoundedJSON(nil)
	}

	opts := []transport.StreamableHTTPCOption{transport.WithHTTPBasicClient(httpClient)}
	if len(headers) > 0 {
		opts = append(opts, transport.WithHTTPHeaders(headers))
	}

	c, err := mcpclient.NewStreamableHttpClient(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client for %s: %w", url, err)
	}

	if _, err = c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP session for %s: %w", url, err)
	}

	return c, nil
}

// BuildAuthHeaders maps (header, value) into an HTTP header map. Empty
// header defaults to "Authorization: Bearer <value>".
func BuildAuthHeaders(authHeader, authValue string) map[string]string {
	if authValue == "" {
		return nil
	}
	headers := make(map[string]string)
	if authHeader == "" {
		headers["Authorization"] = "Bearer " + authValue
	} else {
		headers[authHeader] = authValue
	}
	return headers
}
