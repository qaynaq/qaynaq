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
	// fetches a fresh one from coordinator. Callers should call
	// InvalidateAccessToken on a 401 from the upstream API.
	GetAccessToken(name string) (AccessToken, error)
	// InvalidateAccessToken drops the cached entry, forcing the next
	// GetAccessToken call to round-trip to coordinator.
	InvalidateAccessToken(name string)
}
