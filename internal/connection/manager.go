package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/vault"
)

var providerEndpoints = map[string]oauth2.Endpoint{
	"google": google.Endpoint,
}

type ProviderScope struct {
	Scope       string `json:"scope"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

var ProviderScopes = map[string][]ProviderScope{
	"google": {
		{Scope: "https://www.googleapis.com/auth/calendar", Label: "Google Calendar", Description: "Manage events and calendars"},
		{Scope: "https://www.googleapis.com/auth/contacts.readonly", Label: "Google Contacts", Description: "Read contacts (People API)"},
		{Scope: "https://www.googleapis.com/auth/documents", Label: "Google Docs", Description: "Read and write documents"},
		{Scope: "https://www.googleapis.com/auth/drive", Label: "Google Drive", Description: "Manage files and folders"},
		{Scope: "https://www.googleapis.com/auth/gmail.readonly", Label: "Gmail (Read Only)", Description: "Read emails and labels"},
		{Scope: "https://www.googleapis.com/auth/gmail.modify", Label: "Gmail (Full Access)", Description: "Read, send, and modify emails"},
		{Scope: "https://www.googleapis.com/auth/presentations", Label: "Google Slides", Description: "Read and write presentations"},
		{Scope: "https://www.googleapis.com/auth/spreadsheets", Label: "Google Sheets", Description: "Read and write spreadsheets"},
	},
}

func GetProviders() map[string][]ProviderScope {
	return ProviderScopes
}

func GetDefaultScopes(provider string) []string {
	scopes := ProviderScopes[provider]
	result := make([]string, len(scopes))
	for i, s := range scopes {
		result[i] = s.Scope
	}
	return result
}

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scopes       []string `json:"scopes"`
}

type ConnectionData struct {
	ClientID     string     `json:"client_id"`
	ClientSecret string     `json:"client_secret"`
	Provider     string     `json:"provider"`
	Token        TokenData  `json:"token"`
}

type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

type ConnectionInfo struct {
	Name             string    `json:"name"`
	Provider         string    `json:"provider"`
	Scopes           []string  `json:"scopes"`
	ClientID         string    `json:"client_id"`
	ClientSecretHint string    `json:"client_secret_hint"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// refreshSafetyMargin is how long before the stored expiry we proactively refresh.
// Anything that returns a token will only return one valid for at least this long.
const refreshSafetyMargin = 2 * time.Minute

type cachedAccessToken struct {
	accessToken string
	expiresAt   time.Time
}

type Manager struct {
	connRepo persistence.ConnectionRepository
	aesgcm   *vault.AESGCM

	// per-connection mutex serializes refresh exchanges so concurrent callers
	// don't burn refresh-token rotations against each other.
	refreshMu sync.Map // map[string]*sync.Mutex

	// in-memory access-token cache populated on read/refresh; avoids decrypting
	// the row on every GetAccessToken call.
	cacheMu sync.RWMutex
	cache   map[string]cachedAccessToken
}

func NewManager(connRepo persistence.ConnectionRepository, aesgcm *vault.AESGCM) *Manager {
	return &Manager{
		connRepo: connRepo,
		aesgcm:   aesgcm,
		cache:    make(map[string]cachedAccessToken),
	}
}

func (m *Manager) connMutex(name string) *sync.Mutex {
	if mu, ok := m.refreshMu.Load(name); ok {
		return mu.(*sync.Mutex)
	}
	mu, _ := m.refreshMu.LoadOrStore(name, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

func (m *Manager) cacheGet(name string) (cachedAccessToken, bool) {
	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()
	v, ok := m.cache[name]
	return v, ok
}

func (m *Manager) cachePut(name string, v cachedAccessToken) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	m.cache[name] = v
}

func (m *Manager) cacheInvalidate(name string) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	delete(m.cache, name)
}

func (m *Manager) GetConnectionData(name string) (string, error) {
	conn, err := m.connRepo.GetByName(name)
	if err != nil {
		return "", fmt.Errorf("connection %q not found: %w", name, err)
	}

	configJSON, err := m.aesgcm.Decrypt(conn.EncryptedConfig)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt connection config: %w", err)
	}

	tokenJSON, err := m.aesgcm.Decrypt(conn.EncryptedToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt connection token: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("failed to parse connection config: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return "", fmt.Errorf("failed to parse connection token: %w", err)
	}

	data := ConnectionData{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Provider:     conn.Provider,
		Token: TokenData{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenType:    token.TokenType,
			Expiry:       token.Expiry,
		},
	}

	result, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal connection data: %w", err)
	}

	return string(result), nil
}

func (m *Manager) StoreConnection(name, provider string, cfg Config, token *oauth2.Token) error {
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	encryptedConfig, err := m.aesgcm.Encrypt(string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	encryptedToken, err := m.aesgcm.Encrypt(string(tokenJSON))
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	conn := &persistence.Connection{
		Name:            name,
		Provider:        provider,
		EncryptedConfig: encryptedConfig,
		EncryptedToken:  encryptedToken,
		ExpiresAt:       expiryPtr(token.Expiry),
	}

	if _, err := m.connRepo.Create(conn); err != nil {
		return err
	}

	if token.AccessToken != "" {
		m.cachePut(name, cachedAccessToken{accessToken: token.AccessToken, expiresAt: token.Expiry})
	}
	return nil
}

func (m *Manager) ReauthorizeConnection(name string, cfg Config, token *oauth2.Token) error {
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	encryptedConfig, err := m.aesgcm.Encrypt(string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	encryptedToken, err := m.aesgcm.Encrypt(string(tokenJSON))
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	if err := m.connRepo.UpdateToken(name, encryptedToken, expiryPtr(token.Expiry)); err != nil {
		return err
	}

	if err := m.connRepo.UpdateConfig(name, encryptedConfig); err != nil {
		return err
	}

	if token.AccessToken != "" {
		m.cachePut(name, cachedAccessToken{accessToken: token.AccessToken, expiresAt: token.Expiry})
	} else {
		m.cacheInvalidate(name)
	}
	return nil
}

func (m *Manager) DeleteConnection(name string) error {
	return m.connRepo.Delete(name)
}

func (m *Manager) ListConnections() ([]ConnectionInfo, error) {
	conns, err := m.connRepo.List()
	if err != nil {
		return nil, err
	}

	result := make([]ConnectionInfo, 0, len(conns))
	for _, c := range conns {
		info := ConnectionInfo{
			Name:      c.Name,
			Provider:  c.Provider,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}
		if c.EncryptedConfig != "" {
			if configJSON, err := m.aesgcm.Decrypt(c.EncryptedConfig); err == nil {
				var cfg Config
				if json.Unmarshal([]byte(configJSON), &cfg) == nil {
					info.Scopes = cfg.Scopes
					info.ClientID = cfg.ClientID
					if len(cfg.ClientSecret) >= 3 {
						info.ClientSecretHint = "***" + cfg.ClientSecret[len(cfg.ClientSecret)-3:]
					}
				}
			}
		}
		result = append(result, info)
	}
	return result, nil
}

func (m *Manager) GetStoredConfig(name string) (*Config, error) {
	conn, err := m.connRepo.GetByName(name)
	if err != nil {
		return nil, err
	}

	configJSON, err := m.aesgcm.Decrypt(conn.EncryptedConfig)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GetEndpoint(provider string) (oauth2.Endpoint, error) {
	ep, ok := providerEndpoints[provider]
	if !ok {
		return oauth2.Endpoint{}, fmt.Errorf("unsupported provider: %s", provider)
	}
	return ep, nil
}

func expiryPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	v := t
	return &v
}

// AccessToken is the thin shape returned to clients that just need a usable
// bearer token. Refresh tokens never leave the coordinator.
type AccessToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

// ErrConnectionNotFound is returned when GetAccessToken is called for a name
// that has no stored connection.
var ErrConnectionNotFound = errors.New("connection not found")

// GetAccessToken returns a fresh access token for the named connection. If the
// cached/stored token is within refreshSafetyMargin of expiry, it refreshes
// against the provider before returning.
//
// Concurrent callers for the same name are serialized by a per-connection
// mutex; the second caller through reads the just-persisted token instead of
// triggering its own refresh.
//
// If forceRefresh is true, the cache and stored-token expiry checks are
// skipped and a refresh exchange is performed unconditionally. Use this when
// the caller has just received a 401 using a previously-returned token: the
// stored token may still look valid by the clock but has been revoked or
// rotated upstream.
func (m *Manager) GetAccessToken(ctx context.Context, name string, forceRefresh bool) (AccessToken, error) {
	if !forceRefresh {
		if v, ok := m.cacheGet(name); ok && time.Until(v.expiresAt) > refreshSafetyMargin {
			return AccessToken{AccessToken: v.accessToken, ExpiresAt: v.expiresAt}, nil
		}
	}

	mu := m.connMutex(name)
	mu.Lock()
	defer mu.Unlock()

	if forceRefresh {
		// Drop any cached entry so loadAndMaybeRefreshLocked doesn't re-cache
		// before deciding to refresh.
		m.cacheInvalidate(name)
	} else {
		// Re-check after acquiring the lock - another goroutine may have refreshed.
		if v, ok := m.cacheGet(name); ok && time.Until(v.expiresAt) > refreshSafetyMargin {
			return AccessToken{AccessToken: v.accessToken, ExpiresAt: v.expiresAt}, nil
		}
	}

	return m.loadAndMaybeRefreshLocked(ctx, name, forceRefresh)
}

// InvalidateAccessToken drops the cached entry for a connection, forcing the
// next GetAccessToken call to read the row (and refresh if needed). Called
// when a downstream API returns 401 with the cached token.
func (m *Manager) InvalidateAccessToken(name string) {
	m.cacheInvalidate(name)
}

// RefreshIfExpiring is the entry point for the background job. Same path as
// GetAccessToken's refresh branch but without returning the token.
func (m *Manager) RefreshIfExpiring(ctx context.Context, name string) error {
	mu := m.connMutex(name)
	mu.Lock()
	defer mu.Unlock()

	_, err := m.loadAndMaybeRefreshLocked(ctx, name, false)
	return err
}

func (m *Manager) loadAndMaybeRefreshLocked(ctx context.Context, name string, force bool) (AccessToken, error) {
	conn, err := m.connRepo.GetByName(name)
	if err != nil {
		return AccessToken{}, fmt.Errorf("%w: %q", ErrConnectionNotFound, name)
	}

	cfg, token, err := m.decryptRow(conn)
	if err != nil {
		return AccessToken{}, err
	}

	if !force && !token.Expiry.IsZero() && time.Until(token.Expiry) > refreshSafetyMargin {
		m.cachePut(name, cachedAccessToken{accessToken: token.AccessToken, expiresAt: token.Expiry})
		return AccessToken{AccessToken: token.AccessToken, ExpiresAt: token.Expiry}, nil
	}

	if token.RefreshToken == "" {
		return AccessToken{}, fmt.Errorf("connection %q has no refresh token; re-authorize required", name)
	}

	endpoint, err := GetEndpoint(conn.Provider)
	if err != nil {
		return AccessToken{}, fmt.Errorf("connection %q: %w", name, err)
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     endpoint,
	}

	src := oauth2Cfg.TokenSource(ctx, token)
	newToken, err := src.Token()
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to refresh token for %q: %w", name, err)
	}

	if newToken.RefreshToken == "" {
		// Some providers (Google after first refresh) only return access tokens
		// on subsequent refreshes; preserve the existing refresh token.
		newToken.RefreshToken = token.RefreshToken
	}

	tokenJSON, err := json.Marshal(newToken)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to marshal refreshed token: %w", err)
	}
	encryptedToken, err := m.aesgcm.Encrypt(string(tokenJSON))
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to encrypt refreshed token: %w", err)
	}
	if err := m.connRepo.UpdateToken(name, encryptedToken, expiryPtr(newToken.Expiry)); err != nil {
		return AccessToken{}, fmt.Errorf("failed to persist refreshed token: %w", err)
	}

	m.cachePut(name, cachedAccessToken{accessToken: newToken.AccessToken, expiresAt: newToken.Expiry})

	log.Debug().
		Str("connection", name).
		Time("expires_at", newToken.Expiry).
		Msg("Refreshed OAuth access token")

	return AccessToken{AccessToken: newToken.AccessToken, ExpiresAt: newToken.Expiry}, nil
}

func (m *Manager) decryptRow(conn *persistence.Connection) (*Config, *oauth2.Token, error) {
	configJSON, err := m.aesgcm.Decrypt(conn.EncryptedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt connection config: %w", err)
	}
	tokenJSON, err := m.aesgcm.Decrypt(conn.EncryptedToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt connection token: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to parse connection config: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, nil, fmt.Errorf("failed to parse connection token: %w", err)
	}

	return &cfg, &token, nil
}
