package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	authType      config.AuthType
	authConfig    *config.AuthConfig
	basicHandler  *BasicAuthHandler
	oauth2Handler *OAuth2Handler
}

func NewManager(cfg *config.AuthConfig, secretKey string) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	manager := &Manager{
		authType:   cfg.Type,
		authConfig: cfg,
	}

	switch cfg.Type {
	case config.AuthTypeNone:
		log.Info().Msg("Authentication disabled")
	case config.AuthTypeBasic:
		manager.basicHandler = NewBasicAuthHandler(cfg, secretKey)
		log.Info().Msg("Basic authentication enabled")
	case config.AuthTypeOAuth2:
		oauth2Handler, err := NewOAuth2Handler(cfg, secretKey)
		if err != nil {
			return nil, err
		}
		manager.oauth2Handler = oauth2Handler
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
	_ = json.NewEncoder(w).Encode(response)
}

func (m *Manager) HandleSessionCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{"authenticated": false, "role": "", "email": ""}

	if m.authType == config.AuthTypeNone {
		response["authenticated"] = true
		response["role"] = RoleAdmin
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	switch m.authType {
	case config.AuthTypeBasic:
		cookie, err := r.Cookie("qaynaq_session")
		if err == nil {
			claims, vErr := m.basicHandler.jwtManager.ValidateToken(cookie.Value)
			if vErr == nil && claims.AuthType == "basic" {
				response["authenticated"] = true
				response["role"] = RoleAdmin
				response["email"] = claims.UserID
			}
		}
	case config.AuthTypeOAuth2:
		cookie, err := r.Cookie(m.oauth2Handler.cookieName)
		if err == nil {
			claims, vErr := m.oauth2Handler.jwtManager.ValidateToken(cookie.Value)
			if vErr == nil && claims.AuthType == "oauth2" {
				response["authenticated"] = true
				response["email"] = claims.Email
				if claims.Claims == nil {
					response["role"] = RoleAdmin
				} else {
					response["role"] = EvaluateRole(m.authConfig, claims.Claims, claims.Email)
				}
			}
		}
	}

	_ = json.NewEncoder(w).Encode(response)
}

func (m *Manager) RequireRole(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m.authType == config.AuthTypeNone {
				next.ServeHTTP(w, r)
				return
			}
			if m.authType == config.AuthTypeBasic {
				if required == RoleAdmin {
					next.ServeHTTP(w, r)
					return
				}
				m.denyRole(w, r)
				return
			}

			if m.authType != config.AuthTypeOAuth2 || m.oauth2Handler == nil {
				m.denyRole(w, r)
				return
			}

			token := extractBearer(r.Header.Get("Authorization"))
			if token == "" {
				if cookie, err := r.Cookie(m.oauth2Handler.cookieName); err == nil {
					token = cookie.Value
				}
			}
			if token == "" {
				m.denyRole(w, r)
				return
			}
			claims, err := m.oauth2Handler.jwtManager.ValidateToken(token)
			if err != nil || claims.AuthType != "oauth2" {
				m.denyRole(w, r)
				return
			}

			role := ""
			if claims.Claims == nil {
				role = RoleAdmin
			} else {
				role = EvaluateRole(m.authConfig, claims.Claims, claims.Email)
			}

			if role != required {
				m.denyRole(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Manager) denyRole(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "application/json") || strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden: role required"}`))
		return
	}
	http.Redirect(w, r, "/no-access", http.StatusTemporaryRedirect)
}

func extractBearer(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
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

	http.SetCookie(w, &http.Cookie{ //nolint:gosec // HttpOnly, Secure, SameSite all set below
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
