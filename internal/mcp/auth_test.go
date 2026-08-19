package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeValidator struct{}

func (fakeValidator) IsMCPProtected() bool         { return true }
func (fakeValidator) ValidateMCPToken(string) bool { return false }

func TestUnauthorizedChallengePointsAtPathSuffixedMetadata(t *testing.T) {
	h := AuthMiddleware(fakeValidator{}, nil, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	for _, path := range []string{"/mcp", "/mcp/"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Host = "qaynaq.example.com"
		req.Header.Set("X-Forwarded-Proto", "https")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s: status %d, want 401", path, rr.Code)
		}
		want := `Bearer resource_metadata="https://qaynaq.example.com/.well-known/oauth-protected-resource` + path + `"`
		if got := rr.Header().Get("WWW-Authenticate"); got != want {
			t.Fatalf("%s: WWW-Authenticate %q, want %q", path, got, want)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("%s: Content-Type %q, want application/json", path, ct)
		}
	}
}
