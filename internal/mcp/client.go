package mcp

import (
	"context"
	"fmt"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// ConnectMCPClient opens a streamable-HTTP MCP session. Headers (if any) are
// attached to every request.
func ConnectMCPClient(ctx context.Context, url string, headers map[string]string) (*mcpclient.Client, error) {
	var opts []transport.StreamableHTTPCOption
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
