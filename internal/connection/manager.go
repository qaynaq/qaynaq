package connection

import (
	"encoding/json"
	"fmt"
	"time"

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

type Manager struct {
	connRepo persistence.ConnectionRepository
	aesgcm   *vault.AESGCM
}

func NewManager(connRepo persistence.ConnectionRepository, aesgcm *vault.AESGCM) *Manager {
	return &Manager{
		connRepo: connRepo,
		aesgcm:   aesgcm,
	}
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
	}

	_, err = m.connRepo.Create(conn)
	return err
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

	if err := m.connRepo.UpdateToken(name, encryptedToken); err != nil {
		return err
	}

	return m.connRepo.UpdateConfig(name, encryptedConfig)
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
