package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

func TestPrincipalCanAccess(t *testing.T) {
	oauth := &Principal{Kind: persistence.MCPActorOAuth, Name: "a@b.c", Groups: []string{"accounting"}}
	apiToken := &Principal{Kind: persistence.MCPActorAPIToken, Name: "ci-token"}

	tests := []struct {
		name    string
		p       *Principal
		allowed []string
		want    bool
	}{
		{"unrestricted tool", oauth, nil, true},
		{"member of allowed group", oauth, []string{"legal", "accounting"}, true},
		{"not a member", oauth, []string{"legal"}, false},
		{"api token bypasses restrictions", apiToken, []string{"legal"}, true},
		{"nil principal bypasses restrictions", nil, []string{"legal"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.CanAccess(tt.allowed); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeValidator struct {
	protected bool
	tokenName string
}

func (f fakeValidator) IsMCPProtected() bool { return f.protected }
func (f fakeValidator) ValidateMCPToken(raw string) (string, bool) {
	if raw == "valid" {
		return f.tokenName, true
	}
	return "", false
}

type fakeOAuth struct{ email string }

func (f fakeOAuth) ValidateAccessToken(raw string) (string, bool) {
	if raw == "oauth-token" {
		return f.email, true
	}
	return "", false
}

type fakeGroups struct{ groups []string }

func (f fakeGroups) GroupsForUser(string) []string { return f.groups }

func TestAuthMiddlewarePrincipals(t *testing.T) {
	t.Run("unprotected yields anonymous principal", func(t *testing.T) {
		var captured *Principal
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured = PrincipalFromContext(r.Context())
		})
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		rr := httptest.NewRecorder()
		AuthMiddleware(fakeValidator{protected: false}, nil, nil, inner).ServeHTTP(rr, req)
		if captured == nil || captured.Kind != persistence.MCPActorAnonymous {
			t.Fatalf("expected anonymous principal, got %+v", captured)
		}
	})

	t.Run("oauth token yields oauth principal with groups", func(t *testing.T) {
		var captured *Principal
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured = PrincipalFromContext(r.Context())
		})
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer oauth-token")
		rr := httptest.NewRecorder()
		AuthMiddleware(fakeValidator{protected: true}, fakeOAuth{email: "a@b.c"}, fakeGroups{groups: []string{"accounting"}}, inner).ServeHTTP(rr, req)
		if captured == nil || captured.Kind != persistence.MCPActorOAuth || captured.Name != "a@b.c" {
			t.Fatalf("expected oauth principal, got %+v", captured)
		}
		if len(captured.Groups) != 1 || captured.Groups[0] != "accounting" {
			t.Fatalf("expected groups [accounting], got %v", captured.Groups)
		}
	})

	t.Run("api token yields token principal named after token", func(t *testing.T) {
		var captured *Principal
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured = PrincipalFromContext(r.Context())
		})
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer valid")
		rr := httptest.NewRecorder()
		AuthMiddleware(fakeValidator{protected: true, tokenName: "ci-token"}, nil, nil, inner).ServeHTTP(rr, req)
		if captured == nil || captured.Kind != persistence.MCPActorAPIToken || captured.Name != "ci-token" {
			t.Fatalf("expected api_token principal, got %+v", captured)
		}
	})

	t.Run("invalid token is rejected", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer nope")
		rr := httptest.NewRecorder()
		AuthMiddleware(fakeValidator{protected: true}, nil, nil, inner).ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})
}
