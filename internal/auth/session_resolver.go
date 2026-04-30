package auth

import (
	"net/http"
	"strings"

	"github.com/qaynaq/qaynaq/internal/config"
)

// SessionResolver returns the authenticated Qaynaq user from an incoming
// request, used by the MCP OAuth Authorization Server to delegate end-user
// authentication to the host application's existing login flow.
type SessionResolver struct {
	manager *Manager
}

func NewSessionResolver(m *Manager) *SessionResolver {
	return &SessionResolver{manager: m}
}

func (s *SessionResolver) ResolveUser(r *http.Request) (string, bool) {
	switch s.manager.authType {
	case config.AuthTypeNone:
		return "anonymous@qaynaq.local", true
	case config.AuthTypeOAuth2:
		if s.manager.oauth2Handler == nil {
			return "", false
		}
		cookie, err := r.Cookie(s.manager.oauth2Handler.cookieName)
		if err != nil {
			return "", false
		}
		claims, err := s.manager.oauth2Handler.jwtManager.ValidateToken(cookie.Value)
		if err != nil || claims.AuthType != "oauth2" {
			return "", false
		}
		return claims.Email, true
	case config.AuthTypeBasic:
		if s.manager.basicHandler == nil {
			return "", false
		}
		// Prefer the cookie set by /auth/exchange; fall back to Bearer for
		// the same-origin SPA case.
		var raw string
		if cookie, err := r.Cookie("qaynaq_session"); err == nil {
			raw = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return "", false
			}
			raw = strings.TrimPrefix(authHeader, "Bearer ")
		}
		claims, err := s.manager.basicHandler.jwtManager.ValidateToken(raw)
		if err != nil || claims.AuthType != "basic" {
			return "", false
		}
		if claims.UserID == "" {
			return s.manager.basicHandler.username, true
		}
		return claims.UserID, true
	}
	return "", false
}

func (s *SessionResolver) AuthType() config.AuthType {
	return s.manager.authType
}

func (s *SessionResolver) LoginRedirectPath() string {
	switch s.manager.authType {
	case config.AuthTypeOAuth2:
		return "/auth/login"
	case config.AuthTypeBasic:
		return "/login"
	}
	return ""
}
