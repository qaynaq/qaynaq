package connection

import (
	"context"
	"crypto/rand"
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
	if len(scopes) == 0 {
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

	state := generateState()
	h.stateStoreMux.Lock()
	h.stateStore[state] = &pendingAuth{
		Name:         name,
		Provider:     provider,
		Config:       Config{ClientID: clientID, ClientSecret: clientSecret, Scopes: scopes},
		OAuth2Config: oauth2Config,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}
	h.stateStoreMux.Unlock()

	var authOpts []oauth2.AuthCodeOption

	switch provider {
	case "slack":
		// Slack v2: "scope" requests bot scopes; user scopes must go in
		// "user_scope" or the auth flow tries to install a bot.
		userScopes := strings.Join(oauth2Config.Scopes, ",")
		oauth2Config.Scopes = nil
		authOpts = append(authOpts, oauth2.SetAuthURLParam("user_scope", userScopes))
	default:
		// access_type=offline + prompt=consent forces Google to issue a refresh
		// token even on re-consent.
		authOpts = append(authOpts, oauth2.AccessTypeOffline)
		authOpts = append(authOpts, oauth2.SetAuthURLParam("prompt", "consent"))
	}

	url := oauth2Config.AuthCodeURL(state, authOpts...)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
		fmt.Fprint(w, callbackHTML("error", "Invalid or expired authorization state. Please try again."))
		return
	}

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Error().Str("error", errParam).Str("description", errDesc).Msg("OAuth authorization denied")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, callbackHTML("error", fmt.Sprintf("Authorization denied: %s", errDesc)))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, callbackHTML("error", "Missing authorization code."))
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
			fmt.Fprint(w, callbackHTML("error", "Failed to exchange authorization code for token."))
			return
		}
	} else {
		var err error
		token, err = pending.OAuth2Config.Exchange(context.Background(), code)
		if err != nil {
			log.Error().Err(err).Msg("Failed to exchange OAuth2 code for token")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, callbackHTML("error", "Failed to exchange authorization code for token."))
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
		fmt.Fprint(w, callbackHTML("error", "Failed to store connection."))
		return
	}

	log.Info().Str("name", pending.Name).Str("provider", pending.Provider).Msg("OAuth connection saved")
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, callbackHTML("success", pending.Name))
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
	defer resp.Body.Close()

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
