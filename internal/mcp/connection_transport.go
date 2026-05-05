package mcp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/qaynaq/qaynaq/internal/connection"
)

// newConnectionHTTPClient returns an HTTP client that signs every upstream
// MCP request with a fresh access token from connection.Manager. On 401 it
// force-refreshes the token once and retries.
func newConnectionHTTPClient(mgr *connection.Manager, connectionName string) *http.Client {
	return &http.Client{Transport: &connectionTransport{
		mgr:            mgr,
		connectionName: connectionName,
		base:           http.DefaultTransport,
	}}
}

type connectionTransport struct {
	mgr            *connection.Manager
	connectionName string
	base           http.RoundTripper
}

func (t *connectionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tok, err := t.mgr.GetAccessToken(req.Context(), t.connectionName, false)
	if err != nil {
		return nil, fmt.Errorf("connection %q: %w", t.connectionName, err)
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	bodyBytes, err := snapshotBody(req)
	if err != nil {
		return nil, err
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	// 401 with a token coordinator considered fresh - force a refresh.
	freshTok, ferr := t.mgr.GetAccessToken(req.Context(), t.connectionName, true)
	if ferr != nil {
		return resp, fmt.Errorf("connection %q: 401 from upstream and force refresh failed: %w", t.connectionName, ferr)
	}

	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	retryReq := req.Clone(req.Context())
	retryReq.Header.Set("Authorization", "Bearer "+freshTok.AccessToken)
	if bodyBytes != nil {
		retryReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	} else if req.Body != nil && req.GetBody == nil {
		return resp, errors.New("connection: 401 from upstream and request body is not replayable")
	} else if req.GetBody != nil {
		body, gerr := req.GetBody()
		if gerr != nil {
			return nil, fmt.Errorf("connection: failed to rewind body for retry: %w", gerr)
		}
		retryReq.Body = body
	}

	return t.base.RoundTrip(retryReq)
}

// snapshotBody buffers the request body once so the retry can replay it.
// Returns nil for nil-body requests or when the caller already provided a
// replayable body via req.GetBody.
func snapshotBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.GetBody != nil {
		return nil, nil
	}
	buf, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("connection: failed to buffer request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(buf))
	return buf, nil
}
