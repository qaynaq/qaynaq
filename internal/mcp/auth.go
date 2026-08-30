package mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

type TokenValidator interface {
	IsMCPProtected() bool
	ValidateMCPToken(rawToken string) (name string, ok bool)
}

// OAuthValidator is optional: when nil, only static API tokens are accepted.
type OAuthValidator interface {
	ValidateAccessToken(raw string) (email string, ok bool)
}

// GroupResolver returns the user's IdP groups snapshot captured at login.
type GroupResolver interface {
	GroupsForUser(email string) []string
}

// Principal identifies the MCP caller for access checks and audit logging.
// API-token and anonymous callers are unrestricted; group-based tool access
// applies to OAuth (OIDC-backed) callers only.
type Principal struct {
	Kind   string
	Name   string
	Groups []string
}

type principalKey struct{}

func withPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

func PrincipalFromContext(ctx context.Context) *Principal {
	p, _ := ctx.Value(principalKey{}).(*Principal)
	return p
}

// CanAccess reports whether the principal may call a tool restricted to
// allowedGroups. Empty allowedGroups means unrestricted.
func (p *Principal) CanAccess(allowedGroups []string) bool {
	if len(allowedGroups) == 0 {
		return true
	}
	if p == nil || p.Kind != persistence.MCPActorOAuth {
		return true
	}
	for _, allowed := range allowedGroups {
		for _, g := range p.Groups {
			if g == allowed {
				return true
			}
		}
	}
	return false
}

func AuthMiddleware(validator TokenValidator, oauth OAuthValidator, groups GroupResolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validator.IsMCPProtected() {
			p := &Principal{Kind: persistence.MCPActorAnonymous, Name: "anonymous"}
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
			return
		}

		token := extractToken(r)
		if token == "" {
			writeUnauthorized(w, r, "authentication required, provide token via Authorization header or ?token= query parameter")
			return
		}

		if oauth != nil {
			if email, ok := oauth.ValidateAccessToken(token); ok {
				p := &Principal{Kind: persistence.MCPActorOAuth, Name: email}
				if groups != nil {
					p.Groups = groups.GroupsForUser(email)
				}
				next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
				return
			}
		}

		if name, ok := validator.ValidateMCPToken(token); ok {
			p := &Principal{Kind: persistence.MCPActorAPIToken, Name: name}
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
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
