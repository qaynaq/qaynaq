// Package connauth provides shared HTTP plumbing for components that
// authenticate with a Qaynaq OAuth connection. The package wraps
// vault.GetAccessToken with:
//
//   - an oauth2.TokenSource shim so existing Google API clients work
//     unchanged (they call NewClient(ctx, ts)), but the underlying token is
//     fetched from coordinator instead of refreshed locally
//   - a RoundTripper that transparently invalidates the cached token on a
//     401 response and retries once with a fresh token
//
// Refresh tokens never reach this layer; coordinator owns them. Workers
// only ever see access tokens with their expiry timestamps.
package connauth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"golang.org/x/oauth2"

	"github.com/qaynaq/qaynaq/internal/vault"
)

// Worker scope: there is exactly one VaultProvider per process and component
// constructors don't receive it directly through Bento's service.Resources.
// SetVaultProvider is called at worker boot to register the singleton, and
// components use Provider() to fetch it lazily inside their constructors.
var (
	providerMu sync.RWMutex
	provider   vault.VaultProvider
)

func SetVaultProvider(vp vault.VaultProvider) {
	providerMu.Lock()
	provider = vp
	providerMu.Unlock()
}

func Provider() vault.VaultProvider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return provider
}

// TokenSource returns an oauth2.TokenSource that fetches access tokens for
// the given Qaynaq connection name via the vault provider. Token() never
// includes a refresh token - those stay on coordinator.
func TokenSource(vp vault.VaultProvider, name string) oauth2.TokenSource {
	return &vaultTokenSource{vp: vp, name: name}
}

type vaultTokenSource struct {
	vp   vault.VaultProvider
	name string
}

func (s *vaultTokenSource) Token() (*oauth2.Token, error) {
	tok, err := s.vp.GetAccessToken(s.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token for %q: %w", s.name, err)
	}
	return &oauth2.Token{
		AccessToken: tok.AccessToken,
		TokenType:   "Bearer",
		Expiry:      tok.ExpiresAt,
	}, nil
}

// NewHTTPClient returns an *http.Client that:
//   - sets Authorization: Bearer <token> using vault.GetAccessToken
//   - on a 401 response, invalidates the cached token and retries the
//     request once with a fresh token
//
// The retry loop only fires for 401, only once. Any other status (including
// 403) is returned as-is. Non-replayable bodies (streams that can't be
// reset) are not retried; the original 401 is returned.
func NewHTTPClient(_ context.Context, vp vault.VaultProvider, name string) *http.Client {
	// Use the raw TokenSource (not ReuseTokenSource) so every request hits
	// vault.GetAccessToken, which has its own cache. After
	// InvalidateAccessToken the next Token() call returns a fresh value.
	// ReuseTokenSource would shadow the invalidation with its own cache.
	base := &oauth2.Transport{
		Source: TokenSource(vp, name),
		Base:   http.DefaultTransport,
	}
	return &http.Client{Transport: &retryOn401{vp: vp, name: name, base: base}}
}

type retryOn401 struct {
	vp   vault.VaultProvider
	name string
	base http.RoundTripper
}

func (r *retryOn401) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.base.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	// Cached token was rejected; force coordinator round-trip on the retry.
	r.vp.InvalidateAccessToken(r.name)

	// Replay only if the body is replayable. Streaming bodies (e.g. large
	// file uploads) can't be replayed, so return the 401 as-is and let the
	// caller re-issue. The cache invalidation ensures the next call gets a
	// fresh token regardless.
	if req.Body != nil && req.GetBody == nil {
		return resp, nil
	}

	// Drain and close the 401 body before retrying.
	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	retryReq := req.Clone(req.Context())
	if req.GetBody != nil {
		body, gerr := req.GetBody()
		if gerr != nil {
			return nil, fmt.Errorf("connauth: failed to rewind body for retry: %w", gerr)
		}
		retryReq.Body = body
	}

	return r.base.RoundTrip(retryReq)
}
