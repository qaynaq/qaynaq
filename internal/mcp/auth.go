package mcp

import (
	"net/http"
	"strings"
)

type TokenValidator interface {
	IsMCPProtected() bool
	ValidateMCPToken(rawToken string) bool
}

// OAuthValidator is optional: when nil, only static API tokens are accepted.
type OAuthValidator interface {
	ValidateAccessToken(raw string) (email string, ok bool)
}

func AuthMiddleware(validator TokenValidator, oauth OAuthValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validator.IsMCPProtected() {
			next.ServeHTTP(w, r)
			return
		}

		token := extractToken(r)
		if token == "" {
			writeUnauthorized(w, r, "authentication required, provide token via Authorization header or ?token= query parameter")
			return
		}

		if oauth != nil {
			if _, ok := oauth.ValidateAccessToken(token); ok {
				next.ServeHTTP(w, r)
				return
			}
		}

		if validator.ValidateMCPToken(token) {
			next.ServeHTTP(w, r)
			return
		}

		writeUnauthorized(w, r, "invalid or unauthorized token")
	})
}

func extractToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	return ""
}

// writeUnauthorized emits a 401 with WWW-Authenticate pointing to the
// protected-resource metadata, so MCP clients can discover the auth server.
func writeUnauthorized(w http.ResponseWriter, r *http.Request, msg string) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	resourceMetadata := scheme + "://" + host + "/.well-known/oauth-protected-resource"
	w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+resourceMetadata+`"`)
	http.Error(w, `{"error":"`+msg+`"}`, http.StatusUnauthorized)
}
