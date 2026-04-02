package auth

import (
	"encoding/json"
	"net/http"

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

func (m *Manager) IsEnabled() bool {
	return m.authType != config.AuthTypeNone
}

func (m *Manager) GetAuthType() config.AuthType {
	return m.authType
}
