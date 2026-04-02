package mcp

import (
	"net/http"
	"strings"
)

type TokenValidator interface {
	IsMCPProtected() bool
	ValidateMCPToken(rawToken string) bool
}

func AuthMiddleware(validator TokenValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validator.IsMCPProtected() {
			next.ServeHTTP(w, r)
			return
		}

		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"authentication required, provide token via Authorization header or ?token= query parameter"}`, http.StatusUnauthorized)
			return
		}

		if !validator.ValidateMCPToken(token) {
			http.Error(w, `{"error":"invalid or unauthorized token"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
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
