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

	// http.RoundTripper contract: do not modify the input request. Clone
	// before setting Authorization, and snapshot the body via GetBody (or by
	// buffering once) so we can replay on retry without touching req.Body.
	bodyBytes, err := snapshotBody(req)
	if err != nil {
		return nil, err
	}

	firstReq := cloneRequestForSend(req, bodyBytes)
	firstReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := t.base.RoundTrip(firstReq)
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

	if bodyBytes == nil && req.Body != nil && req.GetBody == nil {
		return resp, errors.New("connection: 401 from upstream and request body is not replayable")
	}

	retryReq := cloneRequestForSend(req, bodyBytes)
	retryReq.Header.Set("Authorization", "Bearer "+freshTok.AccessToken)

	return t.base.RoundTrip(retryReq)
}

// snapshotBody buffers req.Body once for replay, returning the bytes. Returns
// nil for nil-body requests or when the caller already provided a replayable
// body via req.GetBody. Does NOT mutate the input request - the buffered
// bytes are returned and used to construct fresh body readers per attempt.
func snapshotBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.GetBody != nil {
		return nil, nil
	}
	buf, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("connection: failed to buffer request body: %w", err)
	}
	return buf, nil
}

// cloneRequestForSend returns a clone of req with a fresh Body reader. If
// bodyBytes is non-nil it's used; otherwise GetBody is invoked; otherwise
// the body is left as-is (nil-body case).
func cloneRequestForSend(req *http.Request, bodyBytes []byte) *http.Request {
	r := req.Clone(req.Context())
	switch {
	case bodyBytes != nil:
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	case req.GetBody != nil:
		if body, err := req.GetBody(); err == nil {
			r.Body = body
		}
	}
	return r
}
