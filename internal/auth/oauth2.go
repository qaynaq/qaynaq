package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/qaynaq/qaynaq/internal/config"
	"golang.org/x/oauth2"
)

type OAuth2Handler struct {
	config         *oauth2.Config
	userInfoURL    string
	allowedUsers   map[string]bool
	allowedDomains map[string]bool
	jwtManager     *JWTManager
	cookieName     string
	stateStore     map[string]stateEntry
	stateStoreMux  sync.RWMutex
}

type stateEntry struct {
	expiresAt time.Time
	returnTo  string
}

type UserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewOAuth2Handler(cfg *config.AuthConfig, secretKey string) *OAuth2Handler {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.OAuth2ClientID,
		ClientSecret: cfg.OAuth2ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.OAuth2AuthorizationURL,
			TokenURL: cfg.OAuth2TokenURL,
		},
		RedirectURL: cfg.OAuth2RedirectURL,
		Scopes:      cfg.OAuth2Scopes,
	}

	allowedUsers := make(map[string]bool)
	for _, user := range cfg.OAuth2AllowedUsers {
		allowedUsers[strings.ToLower(user)] = true
	}

	allowedDomains := make(map[string]bool)
	for _, domain := range cfg.OAuth2AllowedDomains {
		allowedDomains[strings.ToLower(domain)] = true
	}

	handler := &OAuth2Handler{
		config:         oauth2Config,
		userInfoURL:    cfg.OAuth2UserInfoURL,
		allowedUsers:   allowedUsers,
		allowedDomains: allowedDomains,
		jwtManager:     NewJWTManager(secretKey, 24*time.Hour),
		cookieName:     cfg.OAuth2SessionCookieName,
		stateStore:     make(map[string]stateEntry),
	}

	go handler.cleanupExpiredStates()

	return handler
}

func (h *OAuth2Handler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Authorization header first
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := authHeader[7:]
			claims, err := h.jwtManager.ValidateToken(token)
			if err == nil && claims.AuthType == "oauth2" {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Check cookie
		cookie, err := r.Cookie(h.cookieName)
		if err != nil {
			h.redirectToLogin(w, r)
			return
		}

		token := cookie.Value
		claims, err := h.jwtManager.ValidateToken(token)
		if err != nil || claims.AuthType != "oauth2" {
			h.redirectToLogin(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *OAuth2Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := h.generateState()
	returnTo := safeReturnTo(r.URL.Query().Get("return_to"))
	h.stateStoreMux.Lock()
	h.stateStore[state] = stateEntry{
		expiresAt: time.Now().Add(10 * time.Minute),
		returnTo:  returnTo,
	}
	h.stateStoreMux.Unlock()

	url := h.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuth2Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	entry, ok := h.consumeState(state)
	if !ok {
		log.Error().Msg("Invalid OAuth2 state parameter")
		http.Redirect(w, r, "/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Error().Msg("Missing OAuth2 authorization code")
		http.Redirect(w, r, "/login?error=missing_code", http.StatusTemporaryRedirect)
		return
	}

	token, err := h.config.Exchange(context.Background(), code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange OAuth2 code for token")
		http.Redirect(w, r, "/login?error=token_exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := h.fetchUserInfo(token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch user info")
		http.Redirect(w, r, "/login?error=user_info_failed", http.StatusTemporaryRedirect)
		return
	}

	if !h.isUserAllowed(userInfo.Email) {
		log.Warn().Str("email", userInfo.Email).Msg("User not allowed to access")
		http.Redirect(w, r, "/login?error=access_denied", http.StatusTemporaryRedirect)
		return
	}

	jwtToken, err := h.createJWTToken(userInfo.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create JWT token")
		http.Redirect(w, r, "/login?error=token_generation_failed", http.StatusTemporaryRedirect)
		return
	}

	// HttpOnly cookie so /mcp/oauth/authorize can identify the user.
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})

	if entry.returnTo != "" {
		http.Redirect(w, r, entry.returnTo, http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/?token=%s", jwtToken), http.StatusTemporaryRedirect)
}

func (h *OAuth2Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}

func (h *OAuth2Handler) fetchUserInfo(token *oauth2.Token) (*UserInfo, error) {
	log.Debug().
		Str("url", h.userInfoURL).
		Bool("token_valid", token.Valid()).
		Str("token_type", token.TokenType).
		Msg("Fetching user info from OAuth2 provider")

	client := h.config.Client(context.Background(), token)
	resp, err := client.Get(h.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Error().
			Int("status", resp.StatusCode).
			Str("url", h.userInfoURL).
			Str("response_body", string(bodyBytes)).
			Str("content_type", resp.Header.Get("Content-Type")).
			Bool("token_valid", token.Valid()).
			Msg("User info request failed - check OAuth2 provider configuration and token")
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

func (h *OAuth2Handler) isUserAllowed(email string) bool {
	email = strings.ToLower(email)

	if len(h.allowedUsers) == 0 && len(h.allowedDomains) == 0 {
		return true
	}

	if h.allowedUsers[email] {
		return true
	}

	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		domain := parts[1]
		if h.allowedDomains[domain] {
			return true
		}
	}

	return false
}

func (h *OAuth2Handler) createJWTToken(email string) (string, error) {
	token, err := h.jwtManager.GenerateToken(email, email, "oauth2")
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	log.Info().Str("email", email).Msg("Created JWT token for OAuth2 user")
	return token, nil
}

func (h *OAuth2Handler) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *OAuth2Handler) consumeState(state string) (stateEntry, bool) {
	h.stateStoreMux.Lock()
	defer h.stateStoreMux.Unlock()

	entry, exists := h.stateStore[state]
	if !exists {
		return stateEntry{}, false
	}
	delete(h.stateStore, state)
	if time.Now().After(entry.expiresAt) {
		return stateEntry{}, false
	}
	return entry, true
}

func (h *OAuth2Handler) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	dest := "/auth/login"
	if rt := r.URL.Query().Get("return_to"); rt != "" {
		if safe := safeReturnTo(rt); safe != "" {
			dest += "?return_to=" + url.QueryEscape(safe)
		}
	}
	http.Redirect(w, r, dest, http.StatusTemporaryRedirect)
}

func (h *OAuth2Handler) cleanupExpiredStates() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.stateStoreMux.Lock()
		now := time.Now()
		for state, entry := range h.stateStore {
			if now.After(entry.expiresAt) {
				delete(h.stateStore, state)
			}
		}
		h.stateStoreMux.Unlock()
	}
}

// safeReturnTo accepts only same-origin paths to prevent open-redirect attacks.
func safeReturnTo(raw string) string {
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "/") {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		return ""
	}
	return raw
}

func (h *OAuth2Handler) isValidSession(token string) bool {
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		return false
	}
	return claims.AuthType == "oauth2"
}
