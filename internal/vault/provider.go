package vault

import "time"

// AccessToken is the thin shape returned by GetAccessToken. The refresh token
// is intentionally absent - it never leaves coordinator.
type AccessToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

type VaultProvider interface {
	GetSecret(key string) (string, error)
	GetConnectionToken(name string) (string, error)
	// GetAccessToken returns a cached access token if still valid, otherwise
	// fetches a fresh one from coordinator.
	GetAccessToken(name string) (AccessToken, error)
	// ForceRefreshAccessToken drops the worker's cached entry and asks
	// coordinator to skip its own cache and perform a refresh exchange.
	// Use this after a 401 from the upstream API: the cached token may still
	// look valid by the clock but has been revoked or rotated upstream.
	ForceRefreshAccessToken(name string) (AccessToken, error)
}
