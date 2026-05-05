package vault

import "time"

// AccessToken intentionally omits the refresh token - it never leaves
// coordinator.
type AccessToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

type VaultProvider interface {
	GetSecret(key string) (string, error)
	GetConnectionToken(name string) (string, error)
	GetAccessToken(name string) (AccessToken, error)
	// ForceRefreshAccessToken bypasses both worker and coordinator caches.
	// Call after a 401 - the cached token may still look valid by the clock
	// but has been revoked or rotated upstream.
	ForceRefreshAccessToken(name string) (AccessToken, error)
}
