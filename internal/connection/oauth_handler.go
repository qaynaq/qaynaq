package connection

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type pendingAuth struct {
	Name         string
	Provider     string
	Config       Config
	OAuth2Config *oauth2.Config
	// CodeVerifier holds the PKCE verifier for this flow. Populated only for
	// providers that require PKCE; sent to the token endpoint on exchange.
	// Lives in memory only - dies with the 10-min state TTL. Never persisted.
	CodeVerifier string
	ExpiresAt    time.Time
}

type OAuthHandler struct {
	manager       *Manager
	stateStore    map[string]*pendingAuth
	stateStoreMux sync.RWMutex
}

func NewOAuthHandler(manager *Manager) *OAuthHandler {
	h := &OAuthHandler{
		manager:    manager,
		stateStore: make(map[string]*pendingAuth),
	}
	go h.cleanupExpiredStates()
	return h
}

func (h *OAuthHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	name := r.URL.Query().Get("name")
	clientID := r.URL.Query().Get("client_id")
	clientSecret := r.URL.Query().Get("client_secret")

	if provider == "" || name == "" || clientID == "" {
		http.Error(w, "provider, name, and client_id are required", http.StatusBadRequest)
		return
	}

	if clientSecret == "" {
		existing, err := h.manager.GetStoredConfig(name)
		if err != nil || existing.ClientSecret == "" {
			http.Error(w, "client_secret is required for new connections", http.StatusBadRequest)
			return
		}
		clientSecret = existing.ClientSecret
	}

	endpoint, err := GetEndpoint(provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Shopify uses per-shop URLs: substitute {shop} into the auth/token URLs
	// before building the oauth2.Config so the redirect goes to the correct
	// store. Validate the shop name strictly to prevent URL injection - a
	// value like "evil.com/x" would otherwise turn into a phishing redirect.
	var shop string
	if ProviderUsesShopTemplate(provider) {
		shop = strings.TrimSpace(r.URL.Query().Get("shop"))
		if err := ValidateShopSubdomain(shop); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		endpoint = substituteEndpointTemplate(endpoint, shop)
	}

	// Atlassian needs a Cloud ID to construct API URLs. We collect it now
	// and store it on the connection so consumers don't have to look it up
	// later. Atlassian's accessible-resources endpoint can return multiple
	// sites; manual entry lets the user pick which one this connection
	// addresses.
	var cloudID string
	if ProviderRequiresCloudID(provider) {
		cloudID = strings.TrimSpace(r.URL.Query().Get("cloud_id"))
		if err := ValidateCloudID(cloudID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	scopesParam := r.URL.Query().Get("scopes")
	var scopes []string
	if scopesParam != "" {
		for _, s := range strings.Split(scopesParam, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				scopes = append(scopes, trimmed)
			}
		}
	}
	if len(scopes) == 0 {
		scopes = GetDefaultScopes(provider)
	}
	if len(scopes) == 0 && !ProviderAcceptsNoScopes(provider) {
		http.Error(w, "at least one scope is required", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwdProto := r.Header.Get("X-Forwarded-Proto"); fwdProto != "" {
		scheme = fwdProto
	}
	redirectURL := fmt.Sprintf("%s://%s/connections/oauth/callback", scheme, r.Host)

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     endpoint,
		Scopes:       scopes,
		RedirectURL:  redirectURL,
	}

	authOpts, codeVerifier := buildAuthorizeOpts(provider, oauth2Config)

	state := generateState()
	h.stateStoreMux.Lock()
	h.stateStore[state] = &pendingAuth{
		Name:     name,
		Provider: provider,
		Config: Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Shop:         shop,
			CloudID:      cloudID,
		},
		OAuth2Config: oauth2Config,
		CodeVerifier: codeVerifier,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}
	h.stateStoreMux.Unlock()

	url := oauth2Config.AuthCodeURL(state, authOpts...)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// buildAuthorizeOpts returns the AuthCodeOptions for a provider's authorize
// URL, plus the PKCE code_verifier (if PKCE is required) so the caller can
// store it for the callback. It also mutates oauth2Config.Scopes when a
// provider requires the scopes to live in a different URL parameter (Slack
// puts user scopes in "user_scope=", Asana MCP rejects any "scope=" at all).
//
// This replaces the old `switch provider` branch in HandleAuthorize. New
// providers add an entry here; the rest of the OAuth flow stays generic.
func buildAuthorizeOpts(provider string, oauth2Config *oauth2.Config) ([]oauth2.AuthCodeOption, string) {
	var opts []oauth2.AuthCodeOption

	switch provider {
	case "slack":
		// Slack v2: "scope" requests bot scopes; user scopes must go in
		// "user_scope" or the auth flow tries to install a bot.
		userScopes := strings.Join(oauth2Config.Scopes, ",")
		oauth2Config.Scopes = nil
		opts = append(opts, oauth2.SetAuthURLParam("user_scope", userScopes))

	case "asana":
		// Asana rotates refresh tokens on every refresh. prompt=consent
		// forces the user-consent screen so refresh_token is reliably issued.
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "consent"))

	case "asana_mcp":
		// Asana MCP rejects any scope= parameter. The resource= param binds
		// the token to the MCP server endpoint instead of generic API access.
		oauth2Config.Scopes = nil
		opts = append(opts, oauth2.SetAuthURLParam("resource", "https://mcp.asana.com/v2"))

	case "atlassian":
		// audience= is required by Atlassian's OAuth 2.0 (3LO) flow.
		// prompt=consent ensures refresh tokens are issued (combined with
		// the offline_access scope which the user must include).
		opts = append(opts, oauth2.SetAuthURLParam("audience", "api.atlassian.com"))
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "consent"))

	case "linear":
		// actor=user issues a user-bound token (vs. an app actor token from
		// the client_credentials grant, which we don't use here).
		opts = append(opts, oauth2.SetAuthURLParam("actor", "user"))

	case "notion":
		// Notion uses owner=user to bind the integration to the authorizing
		// user's workspace.
		opts = append(opts, oauth2.SetAuthURLParam("owner", "user"))

	case "github", "github_mcp", "hubspot", "sentry", "shopify", "slack_mcp":
		// Standard OAuth 2.0/2.1: no provider-specific authorize params
		// beyond PKCE (handled below). slack_mcp uses Slack's MCP-dedicated
		// endpoints (oauth/v2_user/authorize + oauth.v2.user.access) which
		// return a clean top-level token response - unlike regular slack,
		// no user_scope handling or authed_user nesting.

	default:
		// Google and any unrecognized provider: access_type=offline +
		// prompt=consent forces issuance of a refresh token even on re-consent.
		opts = append(opts, oauth2.AccessTypeOffline)
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "consent"))
	}

	// PKCE plumbing: required for github_mcp/sentry/shopify per their
	// current docs (May 2026). Compute SHA-256 challenge from a random
	// verifier; verifier is sent on token exchange.
	var codeVerifier string
	if ProviderRequiresPKCE(provider) {
		codeVerifier = generateCodeVerifier()
		challenge := computeCodeChallenge(codeVerifier)
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge", challenge))
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	}

	return opts, codeVerifier
}

// generateCodeVerifier returns a random 32-byte URL-safe string for PKCE.
// RFC 7636 requires 43-128 characters of [A-Z][a-z][0-9]-._~. Base64URL of
// 32 random bytes lands at 43 chars after stripping padding.
func generateCodeVerifier() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// computeCodeChallenge returns the S256 PKCE challenge for a given verifier.
// SHA-256 the verifier, base64url-encode (no padding) per RFC 7636.
func computeCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")

	h.stateStoreMux.Lock()
	pending, exists := h.stateStore[state]
	if exists {
		delete(h.stateStore, state)
	}
	h.stateStoreMux.Unlock()

	if !exists || time.Now().After(pending.ExpiresAt) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, callbackHTML("error", "Invalid or expired authorization state. Please try again."))
		return
	}

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Error().Str("error", errParam).Str("description", errDesc).Msg("OAuth authorization denied")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, callbackHTML("error", fmt.Sprintf("Authorization denied: %s", errDesc)))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, callbackHTML("error", "Missing authorization code."))
		return
	}

	var token *oauth2.Token
	if pending.Provider == "slack" {
		// Slack returns the user token at authed_user.access_token; the
		// top-level access_token is the bot token, which we don't want.
		var err error
		token, err = exchangeSlackUserToken(pending.OAuth2Config, code)
		if err != nil {
			log.Error().Err(err).Msg("Failed to exchange Slack OAuth2 code for token")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, callbackHTML("error", "Failed to exchange authorization code for token."))
			return
		}
	} else {
		var exchangeOpts []oauth2.AuthCodeOption
		if pending.CodeVerifier != "" {
			exchangeOpts = append(exchangeOpts, oauth2.SetAuthURLParam("code_verifier", pending.CodeVerifier))
		}
		var err error
		token, err = pending.OAuth2Config.Exchange(context.Background(), code, exchangeOpts...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to exchange OAuth2 code for token")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, callbackHTML("error", "Failed to exchange authorization code for token."))
			return
		}
	}

	err := h.manager.StoreConnection(pending.Name, pending.Provider, pending.Config, token)
	if err != nil {
		err = h.manager.ReauthorizeConnection(pending.Name, pending.Config, token)
	}
	if err != nil {
		log.Error().Err(err).Str("name", pending.Name).Msg("Failed to store connection")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, callbackHTML("error", "Failed to store connection."))
		return
	}

	log.Info().Str("name", pending.Name).Str("provider", pending.Provider).Msg("OAuth connection saved")
	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprint(w, callbackHTML("success", pending.Name))
}

func (h *OAuthHandler) cleanupExpiredStates() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.stateStoreMux.Lock()
		now := time.Now()
		for state, pending := range h.stateStore {
			if now.After(pending.ExpiresAt) {
				delete(h.stateStore, state)
			}
		}
		h.stateStoreMux.Unlock()
	}
}

// exchangeSlackUserToken handles Slack's non-standard OAuth v2 token response.
// Slack returns user tokens inside authed_user.access_token, not the top-level access_token.
func exchangeSlackUserToken(cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	resp, err := http.PostForm(cfg.Endpoint.TokenURL, map[string][]string{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURL},
	})
	if err != nil {
		return nil, fmt.Errorf("slack token exchange request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		OK         bool   `json:"ok"`
		Error      string `json:"error"`
		AuthedUser struct {
			ID           string `json:"id"`
			AccessToken  string `json:"access_token"`
			TokenType    string `json:"token_type"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
		} `json:"authed_user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode slack token response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("slack token exchange failed: %s", result.Error)
	}

	if result.AuthedUser.AccessToken == "" {
		return nil, fmt.Errorf("slack returned no user access token (did you request user_scope?)")
	}

	token := &oauth2.Token{
		AccessToken:  result.AuthedUser.AccessToken,
		TokenType:    result.AuthedUser.TokenType,
		RefreshToken: result.AuthedUser.RefreshToken,
	}
	if result.AuthedUser.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(result.AuthedUser.ExpiresIn) * time.Second)
	}

	return token, nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func callbackHTML(status, message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>OAuth Connection</title></head>
<body>
<script>
if (window.opener) {
	window.opener.postMessage({type: "oauth_callback", status: "%s", message: "%s"}, "*");
}
window.close();
</script>
<p>%s - You can close this window.</p>
</body></html>`, status, message, message)
}
