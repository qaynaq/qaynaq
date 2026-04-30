package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/qaynaq/qaynaq/internal/config"
)

type Manager struct {
	authType      config.AuthType
	basicHandler  *BasicAuthHandler
	oauth2Handler *OAuth2Handler
}

func NewManager(cfg *config.AuthConfig, secretKey string) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	manager := &Manager{
		authType: cfg.Type,
	}

	switch cfg.Type {
	case config.AuthTypeNone:
		log.Info().Msg("Authentication disabled")
	case config.AuthTypeBasic:
		manager.basicHandler = NewBasicAuthHandler(cfg, secretKey)
		log.Info().Msg("Basic authentication enabled")
	case config.AuthTypeOAuth2:
		manager.oauth2Handler = NewOAuth2Handler(cfg, secretKey)
		log.Info().Msg("OAuth2 authentication enabled")
	}

	return manager, nil
}

func (m *Manager) Middleware(next http.Handler) http.Handler {
	switch m.authType {
	case config.AuthTypeNone:
		return next
	case config.AuthTypeBasic:
		return m.basicHandler.Middleware(next)
	case config.AuthTypeOAuth2:
		return m.oauth2Handler.Middleware(next)
	default:
		return next
	}
}

func (m *Manager) APIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.authType == config.AuthTypeNone {
			next.ServeHTTP(w, r)
			return
		}

		switch m.authType {
		case config.AuthTypeBasic:
			m.basicHandler.Middleware(next).ServeHTTP(w, r)
		case config.AuthTypeOAuth2:
			m.oauth2Handler.Middleware(next).ServeHTTP(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func (m *Manager) SetupAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/info", m.HandleAuthInfo)
	mux.HandleFunc("/auth/session", m.HandleSessionCheck)
	mux.HandleFunc("/auth/exchange", m.HandleExchange)

	if m.authType == config.AuthTypeBasic {
		mux.HandleFunc("/auth/login", m.basicHandler.HandleLogin)
		mux.HandleFunc("/auth/logout", m.basicHandler.HandleLogout)
		log.Info().Msg("Basic auth routes registered")
	}

	if m.authType == config.AuthTypeOAuth2 {
		mux.HandleFunc("/auth/login", m.oauth2Handler.HandleLogin)
		mux.HandleFunc("/auth/callback", m.oauth2Handler.HandleCallback)
		mux.HandleFunc("/auth/logout", m.oauth2Handler.HandleLogout)
		log.Info().Msg("OAuth2 auth routes registered")
	}
}

func (m *Manager) HandleAuthInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"auth_type": string(m.authType),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *Manager) HandleSessionCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.authType == config.AuthTypeNone {
		json.NewEncoder(w).Encode(map[string]bool{"authenticated": true})
		return
	}

	authenticated := false

	switch m.authType {
	case config.AuthTypeBasic:
		cookie, err := r.Cookie("qaynaq_session")
		if err == nil {
			authenticated = m.basicHandler.isValidSession(cookie.Value)
		}
	case config.AuthTypeOAuth2:
		cookie, err := r.Cookie(m.oauth2Handler.cookieName)
		if err == nil {
			authenticated = m.oauth2Handler.isValidSession(cookie.Value)
		}
	}

	json.NewEncoder(w).Encode(map[string]bool{"authenticated": authenticated})
}

// HandleExchange writes the SPA's localStorage JWT (sent as a Bearer header)
// into an HttpOnly session cookie so server-side handlers (notably the MCP
// OAuth /authorize endpoint) can identify the user. Idempotent.
func (m *Manager) HandleExchange(w http.ResponseWriter, r *http.Request) {
	if m.authType == config.AuthTypeNone {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	cookieName := "qaynaq_session"
	switch m.authType {
	case config.AuthTypeBasic:
		if m.basicHandler == nil || !m.basicHandler.isValidSession(token) {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
	case config.AuthTypeOAuth2:
		if m.oauth2Handler == nil || !m.oauth2Handler.isValidSession(token) {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		cookieName = m.oauth2Handler.cookieName
	default:
		http.Error(w, `{"error":"unsupported auth type"}`, http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) IsEnabled() bool {
	return m.authType != config.AuthTypeNone
}

func (m *Manager) GetAuthType() config.AuthType {
	return m.authType
}
