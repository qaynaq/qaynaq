// Package connauth wires Bento components into Qaynaq's vault-backed OAuth
// flow: an oauth2.TokenSource that pulls from coordinator instead of
// refreshing locally, plus a RoundTripper that force-refreshes and retries
// once on a 401. Refresh tokens never reach this layer.
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

// Singleton: Bento's service.Resources doesn't surface our VaultProvider, and
// there's exactly one per worker. SetVaultProvider runs at worker boot;
// component constructors call Provider() lazily.
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

// NewHTTPClient returns a client that signs requests with a vault-issued
// access token and force-refreshes + retries once on 401. Non-401 responses
// pass through; non-replayable bodies are not retried.
func NewHTTPClient(_ context.Context, vp vault.VaultProvider, name string) *http.Client {
	// Raw TokenSource (not ReuseTokenSource): the vault has its own cache,
	// and ReuseTokenSource would shadow our InvalidateAccessToken.
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

	// Force coordinator to bypass its cache; otherwise it'd hand back the
	// same revoked-but-not-yet-expired token we just got rejected with.
	if _, ferr := r.vp.ForceRefreshAccessToken(r.name); ferr != nil {
		return resp, fmt.Errorf("connauth: 401 from upstream and force refresh failed: %w", ferr)
	}

	// Streaming bodies can't be rewound for retry; surface the original 401.
	if req.Body != nil && req.GetBody == nil {
		return resp, nil
	}

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
