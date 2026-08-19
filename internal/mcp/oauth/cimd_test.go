package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProtectedResourceMetadataPathSuffix(t *testing.T) {
	srv, _, _ := newTestServer(t, "user@example.com")

	cases := []struct {
		path     string
		status   int
		resource string
	}{
		{"/.well-known/oauth-protected-resource", http.StatusOK, "http://example.com/mcp"},
		{"/.well-known/oauth-protected-resource/mcp", http.StatusOK, "http://example.com/mcp"},
		{"/.well-known/oauth-protected-resource/mcp/", http.StatusOK, "http://example.com/mcp/"},
		{"/.well-known/oauth-protected-resource/api/flows", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		rr := httptest.NewRecorder()
		srv.HandleProtectedResourceMetadata(rr, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rr.Code != tc.status {
			t.Fatalf("%s: status %d, want %d", tc.path, rr.Code, tc.status)
		}
		if tc.resource == "" {
			continue
		}
		var resp map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("%s: parse response: %v", tc.path, err)
		}
		if resp["resource"] != tc.resource {
			t.Fatalf("%s: resource %q, want %q", tc.path, resp["resource"], tc.resource)
		}
	}
}

func TestRegisterPublicClientGetsNoSecret(t *testing.T) {
	srv, _, _ := newTestServer(t, "user@example.com")

	body := strings.NewReader(`{"redirect_uris":["http://localhost:33418/cb"],"client_name":"pub","token_endpoint_auth_method":"none"}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/oauth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.HandleRegister(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d, body %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if _, ok := resp["client_secret"]; ok {
		t.Fatalf("public client got a client_secret: %s", rr.Body.String())
	}
	if resp["token_endpoint_auth_method"] != "none" {
		t.Fatalf("token_endpoint_auth_method %q, want none", resp["token_endpoint_auth_method"])
	}
}

func TestAuthorizationServerMetadataAdvertisesCIMD(t *testing.T) {
	srv, _, _ := newTestServer(t, "user@example.com")

	rr := httptest.NewRecorder()
	srv.HandleAuthorizationServerMetadata(rr, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil))
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["client_id_metadata_document_supported"] != true {
		t.Fatalf("client_id_metadata_document_supported not advertised: %s", rr.Body.String())
	}
}

func TestCIMDAuthorizationCodeFlow(t *testing.T) {
	var docURL string
	docServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":     docURL,
			"client_name":   "Claude Code",
			"redirect_uris": []string{"http://localhost/callback"},
		})
	}))
	defer docServer.Close()
	docURL = docServer.URL + "/client-metadata.json"

	origClient := cimdHTTPClient
	cimdHTTPClient = docServer.Client()
	defer func() { cimdHTTPClient = origClient }()

	srv, clients, _ := newTestServer(t, "user@example.com")

	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {docURL},
		"redirect_uri":          {"http://localhost:39321/callback"},
		"response_type":         {"code"},
		"state":                 {"xyz"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}.Encode()
	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, httptest.NewRequest(http.MethodGet, authURL, nil))
	if rr.Code != http.StatusFound {
		t.Fatalf("authorize: status %d, location %s", rr.Code, rr.Header().Get("Location"))
	}
	loc, err := url.Parse(rr.Header().Get("Location"))
	if err != nil || loc.Query().Get("code") == "" {
		t.Fatalf("authorize redirect missing code: %s", rr.Header().Get("Location"))
	}

	if stored, err := clients.FindByID(docURL); err != nil || stored.Name != "Claude Code" {
		t.Fatalf("CIMD client not persisted: %+v, %v", stored, err)
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {loc.Query().Get("code")},
		"redirect_uri":  {"http://localhost:39321/callback"},
		"code_verifier": {verifier},
		"client_id":     {docURL},
	}
	req := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	trr := httptest.NewRecorder()
	srv.HandleToken(trr, req)
	if trr.Code != http.StatusOK {
		t.Fatalf("token: status %d, body %s", trr.Code, trr.Body.String())
	}
	var tokens map[string]any
	if err := json.Unmarshal(trr.Body.Bytes(), &tokens); err != nil || tokens["access_token"] == "" {
		t.Fatalf("token response invalid: %s", trr.Body.String())
	}
}

func TestCIMDRejectsMismatchedClientID(t *testing.T) {
	docServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":     "https://evil.example.com/other.json",
			"redirect_uris": []string{"http://localhost/callback"},
		})
	}))
	defer docServer.Close()

	origClient := cimdHTTPClient
	cimdHTTPClient = docServer.Client()
	defer func() { cimdHTTPClient = origClient }()

	srv, _, _ := newTestServer(t, "user@example.com")
	if _, err := srv.resolveCIMDClient(docServer.URL + "/client-metadata.json"); err == nil {
		t.Fatal("expected mismatched client_id to be rejected")
	}
}

func TestLoopbackRedirectIgnoresPort(t *testing.T) {
	allowed := []string{"http://localhost/callback", "http://127.0.0.1/callback"}
	for _, ok := range []string{"http://localhost:3118/callback", "http://127.0.0.1:49152/callback"} {
		if !redirectURIAllowed(allowed, ok) {
			t.Fatalf("%s should be allowed", ok)
		}
	}
	for _, bad := range []string{"http://localhost:3118/other", "https://attacker.example.com/callback", "http://192.168.1.5:3118/callback"} {
		if redirectURIAllowed(allowed, bad) {
			t.Fatalf("%s should be rejected", bad)
		}
	}
}
