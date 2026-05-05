package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/vault"
)

// providerEndpoints holds OAuth auth/token URLs per provider. A URL containing
// {shop} (Shopify) or {cloud_id} (placeholder pattern, not currently used) is
// templated at authorize/callback time using values stored in pendingAuth. See
// substituteEndpointTemplate.
var providerEndpoints = map[string]oauth2.Endpoint{
	"google": google.Endpoint,
	"slack": {
		AuthURL:  "https://slack.com/oauth/v2/authorize",
		TokenURL: "https://slack.com/api/oauth.v2.access",
	},
	// slack_mcp connects to Slack's official remote MCP server at
	// https://mcp.slack.com/mcp. Unlike regular slack, this flow uses
	// dedicated user-token endpoints that return tokens at the top level
	// (no authed_user nesting) - so it does NOT need exchangeSlackUserToken.
	// Endpoints sourced from
	// https://mcp.slack.com/.well-known/oauth-authorization-server.
	"slack_mcp": {
		AuthURL:  "https://slack.com/oauth/v2_user/authorize",
		TokenURL: "https://slack.com/api/oauth.v2.user.access",
	},
	"asana": {
		AuthURL:  "https://app.asana.com/-/oauth_authorize",
		TokenURL: "https://app.asana.com/-/oauth_token",
	},
	// asana_mcp connects Qaynaq's MCP proxy to Asana's MCP server
	// (https://mcp.asana.com/v2). Tokens issued through this flow only work
	// against the MCP endpoint - they aren't generic Asana API tokens.
	"asana_mcp": {
		AuthURL:  "https://app.asana.com/-/oauth_authorize",
		TokenURL: "https://app.asana.com/-/oauth_token",
	},
	"atlassian": {
		AuthURL:  "https://auth.atlassian.com/authorize",
		TokenURL: "https://auth.atlassian.com/oauth/token",
	},
	"github": {
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
	"github_mcp": {
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
	"hubspot": {
		AuthURL:  "https://app.hubspot.com/oauth/authorize",
		TokenURL: "https://api.hubapi.com/oauth/2026-03/token",
	},
	"linear": {
		AuthURL:  "https://linear.app/oauth/authorize",
		TokenURL: "https://api.linear.app/oauth/token",
	},
	"notion": {
		AuthURL:  "https://api.notion.com/v1/oauth/authorize",
		TokenURL: "https://api.notion.com/v1/oauth/token",
	},
	"sentry": {
		AuthURL:  "https://sentry.io/oauth/authorize/",
		TokenURL: "https://sentry.io/oauth/token/",
	},
	// shopify uses per-shop URLs - the {shop} placeholder is replaced with the
	// user-supplied shop subdomain (e.g., "my-store") at authorize/callback
	// time. See substituteEndpointTemplate and pendingAuth.Shop.
	"shopify": {
		AuthURL:  "https://{shop}.myshopify.com/admin/oauth/authorize",
		TokenURL: "https://{shop}.myshopify.com/admin/oauth/access_token",
	},
}

type ProviderScope struct {
	Scope       string `json:"scope"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type ProviderSetup struct {
	URL   string `json:"setup_url"`
	Label string `json:"setup_label"`
}

var ProviderSetups = map[string]ProviderSetup{
	"google": {
		URL:   "https://console.cloud.google.com/apis/credentials",
		Label: "Create an OAuth Client ID (Web application type) in GCP Console.",
	},
	"slack": {
		URL:   "https://api.slack.com/apps",
		Label: "Create a Slack App and get Client ID / Client Secret from Basic Information.",
	},
	"slack_mcp": {
		URL:   "https://api.slack.com/apps",
		Label: "Create a Slack App for MCP. Approve the requested user scopes from this picker. PKCE is required.",
	},
	"asana": {
		URL:   "https://app.asana.com/0/my-apps",
		Label: "Create an Asana app and copy the Client ID / Client Secret from the app's settings.",
	},
	"asana_mcp": {
		URL:   "https://app.asana.com/0/my-apps",
		Label: "Create an Asana app, mark it as MCP-enabled, and configure workspace distribution. Use the Client ID / Client Secret from the app's settings.",
	},
	"atlassian": {
		URL:   "https://developer.atlassian.com/console/myapps/",
		Label: "Create an OAuth 2.0 (3LO) integration. Add Jira and/or Confluence APIs. You will be asked for your Atlassian Cloud ID separately.",
	},
	"github": {
		URL:   "https://github.com/settings/developers",
		Label: "Create an OAuth App (not a GitHub App). Set the callback URL to Qaynaq's redirect URI below.",
	},
	"github_mcp": {
		URL:   "https://github.com/settings/developers",
		Label: "Create an OAuth App for GitHub's remote MCP server. Set the callback URL to Qaynaq's redirect URI. PKCE is required.",
	},
	"hubspot": {
		URL:   "https://developers.hubspot.com/",
		Label: "Create a HubSpot public app. Use the OAuth Client ID / Client Secret from the Auth tab.",
	},
	"linear": {
		URL:   "https://linear.app/settings/api/applications/new",
		Label: "Create a Linear OAuth application. Use authorization-code flow with actor=user.",
	},
	"notion": {
		URL:   "https://www.notion.so/my-integrations",
		Label: "Create a public Notion integration with OAuth enabled. Copy the OAuth Client ID / Client Secret.",
	},
	"sentry": {
		URL:   "https://sentry.io/settings/account/api/applications/",
		Label: "Create a Sentry OAuth application. PKCE is recommended.",
	},
	"shopify": {
		URL:   "https://partners.shopify.com/",
		Label: "Create a Shopify app in your Partner dashboard. You will be asked for the shop subdomain separately.",
	},
}

// ProviderDisplayNames is the human-readable label rendered in the UI provider
// dropdown and error messages. Without this map, the UI title-cases the slug,
// which produces ugly labels like "Asana_mcp" or "Github_mcp". A custom
// display name fixes that without changing the stable provider id.
var ProviderDisplayNames = map[string]string{
	"google":     "Google",
	"slack":      "Slack",
	"slack_mcp":  "Slack MCP",
	"asana":      "Asana",
	"asana_mcp":  "Asana MCP",
	"atlassian":  "Atlassian (Jira & Confluence)",
	"github":     "GitHub",
	"github_mcp": "GitHub MCP",
	"hubspot":    "HubSpot",
	"linear":     "Linear",
	"notion":     "Notion",
	"sentry":     "Sentry",
	"shopify":    "Shopify",
}

var ProviderScopes = map[string][]ProviderScope{
	"slack": {
		{Scope: "search:read.public", Label: "Search (Public)", Description: "Search messages in public channels"},
		{Scope: "search:read.private", Label: "Search (Private)", Description: "Search messages in private channels"},
		{Scope: "search:read.files", Label: "Search (Files)", Description: "Search files"},
		{Scope: "channels:history", Label: "Channel History", Description: "View messages in public channels"},
		{Scope: "groups:history", Label: "Group History", Description: "View messages in private channels"},
		{Scope: "chat:write", Label: "Chat (Write)", Description: "Send messages"},
		{Scope: "users:read", Label: "Users (Read)", Description: "View users and their info"},
		{Scope: "users:read.email", Label: "Users (Email)", Description: "View user email addresses"},
		{Scope: "channels:read", Label: "Channels (Read)", Description: "View channels and their info"},
		{Scope: "reactions:read", Label: "Reactions (Read)", Description: "View emoji reactions"},
		{Scope: "reactions:write", Label: "Reactions (Write)", Description: "Add and remove emoji reactions"},
		{Scope: "files:read", Label: "Files (Read)", Description: "View files shared in channels"},
	},
	// slack_mcp scopes are exactly the set published by Slack's
	// .well-known/oauth-authorization-server endpoint. Different from regular
	// `slack` - includes canvases:* and search:read.users.
	"slack_mcp": {
		{Scope: "search:read.public", Label: "Search (Public)", Description: "Search messages in public channels"},
		{Scope: "search:read.private", Label: "Search (Private)", Description: "Search messages in private channels"},
		{Scope: "search:read.mpim", Label: "Search (Group DMs)", Description: "Search messages in multi-party DMs"},
		{Scope: "search:read.im", Label: "Search (DMs)", Description: "Search messages in direct messages"},
		{Scope: "search:read.files", Label: "Search (Files)", Description: "Search files"},
		{Scope: "search:read.users", Label: "Search (Users)", Description: "Search users"},
		{Scope: "chat:write", Label: "Chat (Write)", Description: "Send messages"},
		{Scope: "channels:history", Label: "Channel History", Description: "View messages in public channels"},
		{Scope: "groups:history", Label: "Group History", Description: "View messages in private channels"},
		{Scope: "mpim:history", Label: "Group DM History", Description: "View messages in multi-party DMs"},
		{Scope: "im:history", Label: "DM History", Description: "View messages in direct messages"},
		{Scope: "canvases:read", Label: "Canvases (Read)", Description: "Read Slack canvases"},
		{Scope: "canvases:write", Label: "Canvases (Write)", Description: "Create and edit Slack canvases"},
		{Scope: "users:read", Label: "Users (Read)", Description: "View users and their info"},
		{Scope: "users:read.email", Label: "Users (Email)", Description: "View user email addresses"},
		{Scope: "reactions:write", Label: "Reactions (Write)", Description: "Add and remove emoji reactions"},
	},
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
	"asana": {
		{Scope: "default", Label: "Default (full access)", Description: "Full access to the user's Asana resources"},
	},
	"atlassian": {
		// offline_access is required to receive a refresh token from Atlassian.
		{Scope: "offline_access", Label: "Refresh Tokens", Description: "Required for refresh tokens (offline_access)"},
		{Scope: "read:jira-user", Label: "Jira Users (Read)", Description: "Read Jira users"},
		{Scope: "read:jira-work", Label: "Jira Work (Read)", Description: "Read Jira issues, boards, projects"},
		{Scope: "write:jira-work", Label: "Jira Work (Write)", Description: "Create and update Jira issues"},
		{Scope: "read:confluence-content.all", Label: "Confluence Content (Read)", Description: "Read Confluence pages and spaces"},
		{Scope: "write:confluence-content", Label: "Confluence Content (Write)", Description: "Create and update Confluence pages"},
	},
	"github": {
		{Scope: "repo", Label: "Repositories", Description: "Full access to repositories"},
		{Scope: "read:org", Label: "Organization (Read)", Description: "Read organization membership"},
		{Scope: "read:user", Label: "User (Read)", Description: "Read user profile"},
		{Scope: "user:email", Label: "User Email", Description: "Read user email addresses"},
		{Scope: "workflow", Label: "Workflows", Description: "Update GitHub Actions workflow files"},
	},
	"github_mcp": {
		{Scope: "repo", Label: "Repositories", Description: "Full access to repositories"},
		{Scope: "read:org", Label: "Organization (Read)", Description: "Read organization membership"},
		{Scope: "read:user", Label: "User (Read)", Description: "Read user profile"},
	},
	"hubspot": {
		{Scope: "crm.objects.contacts.read", Label: "Contacts (Read)", Description: "Read CRM contacts"},
		{Scope: "crm.objects.contacts.write", Label: "Contacts (Write)", Description: "Create and update CRM contacts"},
		{Scope: "crm.objects.companies.read", Label: "Companies (Read)", Description: "Read CRM companies"},
		{Scope: "crm.objects.deals.read", Label: "Deals (Read)", Description: "Read CRM deals"},
		{Scope: "crm.objects.deals.write", Label: "Deals (Write)", Description: "Create and update CRM deals"},
		{Scope: "tickets", Label: "Tickets", Description: "Read and write support tickets"},
	},
	"linear": {
		{Scope: "read", Label: "Read", Description: "Read access to user's workspace"},
		{Scope: "write", Label: "Write", Description: "Create and update issues, projects, comments"},
		{Scope: "issues:create", Label: "Issues (Create)", Description: "Create issues"},
		{Scope: "comments:create", Label: "Comments (Create)", Description: "Add comments"},
	},
	"notion": {
		// Notion does not use OAuth scopes the way other providers do - the
		// integration's permissions are configured in the Notion UI and the
		// scope param is informational. We send a single placeholder so the
		// scope picker has something to show.
		{Scope: "read_content", Label: "Read Content", Description: "Read pages and databases the integration has access to"},
		{Scope: "update_content", Label: "Update Content", Description: "Update pages and database entries"},
		{Scope: "insert_content", Label: "Insert Content", Description: "Create new pages and database entries"},
	},
	"sentry": {
		{Scope: "org:read", Label: "Organization (Read)", Description: "Read organization details"},
		{Scope: "project:read", Label: "Projects (Read)", Description: "Read projects"},
		{Scope: "project:write", Label: "Projects (Write)", Description: "Update projects"},
		{Scope: "team:read", Label: "Teams (Read)", Description: "Read teams"},
		{Scope: "event:read", Label: "Events (Read)", Description: "Read events and issues"},
		{Scope: "member:read", Label: "Members (Read)", Description: "Read organization members"},
	},
	"shopify": {
		{Scope: "read_products", Label: "Products (Read)", Description: "Read product catalog"},
		{Scope: "write_products", Label: "Products (Write)", Description: "Create and update products"},
		{Scope: "read_orders", Label: "Orders (Read)", Description: "Read orders"},
		{Scope: "write_orders", Label: "Orders (Write)", Description: "Create and update orders"},
		{Scope: "read_customers", Label: "Customers (Read)", Description: "Read customer data"},
		{Scope: "read_inventory", Label: "Inventory (Read)", Description: "Read inventory levels"},
	},
}

// scopelessProviders bypass the "at least one scope" check at authorize time.
// Asana MCP is the canonical case: it rejects any scope= parameter on the
// authorize URL.
var scopelessProviders = map[string]bool{
	"asana_mcp": true,
}

// pkceProviders force PKCE (S256) on authorize. GitHub MCP requires it; Sentry
// recommends it; Shopify uses it for token security per Shopify 2026 changes.
var pkceProviders = map[string]bool{
	"github_mcp": true,
	"sentry":     true,
	"shopify":    true,
	"slack_mcp":  true,
}

// noRefreshProviders never refresh - the access token doesn't expire (e.g.,
// GitHub OAuth App tokens), or refresh isn't supported by the provider.
// loadAndMaybeRefreshLocked early-returns the stored access token instead of
// calling the refresh endpoint.
var noRefreshProviders = map[string]bool{
	"github": true,
}

// shopTemplateProviders use a per-instance subdomain in their auth/token URLs
// (e.g., Shopify's {shop}.myshopify.com). The OAuth handler reads a "shop"
// query param at authorize time, validates it, stores it in pendingAuth, and
// substitutes {shop} into the URLs at authorize and callback time.
var shopTemplateProviders = map[string]bool{
	"shopify": true,
}

// cloudIDProviders require the user to supply an instance identifier alongside
// their OAuth credentials (e.g., Atlassian Cloud ID). The handler reads a
// "cloud_id" query param at authorize time and stores it in pendingAuth and
// the connection's encrypted Config so consumers can construct API URLs like
// https://api.atlassian.com/ex/jira/{cloudid}/...
var cloudIDProviders = map[string]bool{
	"atlassian": true,
}

func GetProviders() map[string][]ProviderScope {
	out := make(map[string][]ProviderScope, len(ProviderScopes)+len(scopelessProviders))
	for k, v := range ProviderScopes {
		out[k] = v
	}
	for k := range scopelessProviders {
		if _, ok := out[k]; !ok {
			out[k] = nil
		}
	}
	return out
}

// ProviderAcceptsNoScopes reports whether a provider can be authorized without
// any scopes selected. Most providers require at least one scope; scopeless
// providers (Asana MCP) reject any scope= parameter outright.
func ProviderAcceptsNoScopes(provider string) bool {
	return scopelessProviders[provider]
}

// ProviderRequiresPKCE reports whether the provider mandates PKCE on the
// authorize+exchange round-trip. PKCE-required providers fail the flow if the
// code_challenge / code_verifier dance is missing.
func ProviderRequiresPKCE(provider string) bool {
	return pkceProviders[provider]
}

// ProviderUsesShopTemplate reports whether the provider's URLs contain a
// {shop} placeholder that must be substituted with a user-supplied subdomain.
func ProviderUsesShopTemplate(provider string) bool {
	return shopTemplateProviders[provider]
}

// ProviderRequiresCloudID reports whether the provider needs a per-instance
// identifier (e.g., Atlassian Cloud ID) stored alongside the connection.
func ProviderRequiresCloudID(provider string) bool {
	return cloudIDProviders[provider]
}

// GetDisplayName returns a human-readable label for the provider. Falls back
// to the slug if no display name is registered.
func GetDisplayName(provider string) string {
	if name, ok := ProviderDisplayNames[provider]; ok {
		return name
	}
	return provider
}

// substituteEndpointTemplate replaces {shop} in auth/token URLs with the
// user-supplied shop subdomain. Returns the original endpoint unchanged if
// the provider doesn't use templating.
func substituteEndpointTemplate(ep oauth2.Endpoint, shop string) oauth2.Endpoint {
	if shop == "" {
		return ep
	}
	return oauth2.Endpoint{
		AuthURL:   strings.ReplaceAll(ep.AuthURL, "{shop}", shop),
		TokenURL:  strings.ReplaceAll(ep.TokenURL, "{shop}", shop),
		AuthStyle: ep.AuthStyle,
	}
}

// shopSubdomainPattern restricts shop names to lowercase letters, digits, and
// hyphens (Shopify's convention). Without this, a malicious value like
// "evil.com/x?" could redirect authorize to an attacker URL.
var shopSubdomainPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// ValidateShopSubdomain returns nil for valid Shopify shop names, an error
// otherwise. The check is strict enough to prevent URL injection through the
// {shop} template substitution.
func ValidateShopSubdomain(shop string) error {
	if shop == "" {
		return errors.New("shop is required")
	}
	if len(shop) > 63 {
		return errors.New("shop name too long")
	}
	if !shopSubdomainPattern.MatchString(shop) {
		return errors.New("shop must be a lowercase alphanumeric subdomain (e.g., my-store)")
	}
	return nil
}

// cloudIDPattern matches Atlassian Cloud IDs (UUID v4-style). Atlassian's
// cloudid is a UUID returned by /oauth/token/accessible-resources.
var cloudIDPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

// ValidateCloudID returns nil for well-formed UUIDs, an error otherwise.
func ValidateCloudID(cloudID string) error {
	if cloudID == "" {
		return errors.New("cloud_id is required")
	}
	if !cloudIDPattern.MatchString(cloudID) {
		return errors.New("cloud_id must be a UUID (find it at https://YOUR-SITE.atlassian.net/_edge/tenant_info)")
	}
	return nil
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
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	// Shop is the Shopify shop subdomain (e.g., "my-store" for
	// my-store.myshopify.com). Empty for non-Shopify providers. Old encrypted
	// blobs without this field deserialize to "" via json's zero-value rule -
	// no migration needed.
	Shop string `json:"shop,omitempty"`
	// CloudID is the Atlassian Cloud ID (UUID) needed to construct API URLs
	// like https://api.atlassian.com/ex/jira/{cloudid}/... Empty for
	// non-Atlassian providers.
	CloudID string `json:"cloud_id,omitempty"`
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
	Shop             string    `json:"shop,omitempty"`
	CloudID          string    `json:"cloud_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

const refreshSafetyMargin = 2 * time.Minute

type cachedAccessToken struct {
	accessToken string
	expiresAt   time.Time
}

type Manager struct {
	connRepo persistence.ConnectionRepository
	aesgcm   *vault.AESGCM

	// Serializes refresh exchanges per connection so concurrent callers don't
	// burn refresh-token rotations against each other.
	refreshMu sync.Map // map[string]*sync.Mutex

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
					info.Shop = cfg.Shop
					info.CloudID = cfg.CloudID
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

// AccessToken is the thin shape returned to callers. Refresh tokens never
// leave the coordinator.
type AccessToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

var ErrConnectionNotFound = errors.New("connection not found")

// GetAccessToken returns a usable access token for the named connection,
// refreshing against the provider if the stored one is within
// refreshSafetyMargin of expiry. Concurrent callers for the same name are
// serialized by a per-connection mutex.
//
// forceRefresh skips the cache and stored-token validity checks - use it
// when the caller has just received a 401 using a previously-returned token.
func (m *Manager) GetAccessToken(ctx context.Context, name string, forceRefresh bool) (AccessToken, error) {
	if !forceRefresh {
		if v, ok := m.cacheGet(name); ok && cachedTokenStillValid(v) {
			return AccessToken{AccessToken: v.accessToken, ExpiresAt: v.expiresAt}, nil
		}
	}

	mu := m.connMutex(name)
	mu.Lock()
	defer mu.Unlock()

	if forceRefresh {
		m.cacheInvalidate(name)
	} else {
		// Another goroutine may have refreshed while we waited on the mutex.
		if v, ok := m.cacheGet(name); ok && cachedTokenStillValid(v) {
			return AccessToken{AccessToken: v.accessToken, ExpiresAt: v.expiresAt}, nil
		}
	}

	return m.loadAndMaybeRefreshLocked(ctx, name, forceRefresh)
}

// cachedTokenStillValid: zero ExpiresAt means a non-expiring token (Slack
// user tokens without rotation), always valid.
func cachedTokenStillValid(v cachedAccessToken) bool {
	if v.expiresAt.IsZero() {
		return true
	}
	return time.Until(v.expiresAt) > refreshSafetyMargin
}

func (m *Manager) InvalidateAccessToken(name string) {
	m.cacheInvalidate(name)
}

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

	// Non-refreshing providers (e.g., GitHub OAuth App): tokens semantically
	// don't expire upstream, even if a stored Expiry timestamp suggests
	// otherwise. Treat the cached access token as valid until the user
	// revokes it manually. Calling refresh would just fail.
	if noRefreshProviders[conn.Provider] {
		m.cachePut(name, cachedAccessToken{accessToken: token.AccessToken, expiresAt: time.Time{}})
		return AccessToken{AccessToken: token.AccessToken}, nil
	}

	// Non-expiring token (Slack without rotation): no expiry, no refresh
	// token, valid until the user revokes it upstream.
	if token.Expiry.IsZero() && token.RefreshToken == "" {
		m.cachePut(name, cachedAccessToken{accessToken: token.AccessToken, expiresAt: time.Time{}})
		return AccessToken{AccessToken: token.AccessToken}, nil
	}

	if token.RefreshToken == "" {
		return AccessToken{}, fmt.Errorf("connection %q access token expired and no refresh token is available; re-authorize required", name)
	}

	endpoint, err := GetEndpoint(conn.Provider)
	if err != nil {
		return AccessToken{}, fmt.Errorf("connection %q: %w", name, err)
	}

	// Shop-templated providers (Shopify) substitute the shop subdomain into
	// the token URL before refresh. Without this, the refresh request goes
	// to literal "{shop}.myshopify.com" and fails with DNS error.
	if shopTemplateProviders[conn.Provider] {
		endpoint = substituteEndpointTemplate(endpoint, cfg.Shop)
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
		// Google omits refresh_token on subsequent refreshes; keep the original.
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
