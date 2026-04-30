package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/qaynaq/qaynaq/internal/persistence"
)

// In-memory fakes keep tests free of a SQLite dependency and of any GORM
// behavior that is not relevant to the OAuth flow itself.

type fakeClientRepo struct {
	mu      sync.Mutex
	clients map[string]*persistence.OAuthClient
}

func newFakeClientRepo() *fakeClientRepo {
	return &fakeClientRepo{clients: map[string]*persistence.OAuthClient{}}
}

func (r *fakeClientRepo) List() ([]persistence.OAuthClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]persistence.OAuthClient, 0, len(r.clients))
	for _, c := range r.clients {
		out = append(out, *c)
	}
	return out, nil
}
func (r *fakeClientRepo) Create(c *persistence.OAuthClient) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c.ID] = c
	return nil
}
func (r *fakeClientRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, id)
	return nil
}
func (r *fakeClientRepo) FindByID(id string) (*persistence.OAuthClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clients[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	cc := *c
	return &cc, nil
}
func (r *fakeClientRepo) UpdateLastUsedAt(id string, t time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.clients[id]; ok {
		c.LastUsedAt = &t
	}
	return nil
}

type fakeRefreshRepo struct {
	mu     sync.Mutex
	tokens map[string]*persistence.OAuthRefreshToken
	next   int64
}

func newFakeRefreshRepo() *fakeRefreshRepo {
	return &fakeRefreshRepo{tokens: map[string]*persistence.OAuthRefreshToken{}}
}

func (r *fakeRefreshRepo) Create(t *persistence.OAuthRefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.next++
	t.ID = r.next
	r.tokens[t.TokenHash] = t
	return nil
}
func (r *fakeRefreshRepo) FindByHash(hash string) (*persistence.OAuthRefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	tt := *t
	return &tt, nil
}
func (r *fakeRefreshRepo) Revoke(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for _, t := range r.tokens {
		if t.ID == id {
			t.RevokedAt = &now
			return nil
		}
	}
	return nil
}
func (r *fakeRefreshRepo) DeleteByClient(clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, t := range r.tokens {
		if t.ClientID == clientID {
			delete(r.tokens, hash)
		}
	}
	return nil
}
func (r *fakeRefreshRepo) DeleteExpired(before time.Time) error { return nil }
func (r *fakeRefreshRepo) FindByID(id int64) (*persistence.OAuthRefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.ID == id {
			tt := *t
			return &tt, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (r *fakeRefreshRepo) ListActiveSessions() ([]persistence.OAuthSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]persistence.OAuthSession, 0, len(r.tokens))
	for _, t := range r.tokens {
		if t.RevokedAt != nil || time.Now().After(t.ExpiresAt) {
			continue
		}
		out = append(out, persistence.OAuthSession{
			ID:        t.ID,
			ClientID:  t.ClientID,
			UserEmail: t.UserEmail,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
		})
	}
	return out, nil
}

type fakeSession struct {
	email string
}

func (s fakeSession) ResolveUser(_ *http.Request) (string, bool) {
	if s.email == "" {
		return "", false
	}
	return s.email, true
}
func (s fakeSession) AuthType() config.AuthType { return config.AuthTypeOAuth2 }
func (s fakeSession) LoginRedirectPath() string { return "/auth/login" }

type fakeConsentRepo struct {
	mu       sync.Mutex
	approved map[string]bool // userEmail|clientID|scope -> true
}

func newFakeConsentRepo() *fakeConsentRepo {
	return &fakeConsentRepo{approved: map[string]bool{}}
}
func (r *fakeConsentRepo) key(userEmail, clientID, scope string) string {
	return userEmail + "|" + clientID + "|" + scope
}
func (r *fakeConsentRepo) Has(userEmail, clientID, scope string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.approved[r.key(userEmail, clientID, scope)], nil
}
func (r *fakeConsentRepo) Upsert(c *persistence.OAuthConsent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.approved[r.key(c.UserEmail, c.ClientID, c.Scope)] = true
	return nil
}
func (r *fakeConsentRepo) ApprovedClientIDs(userEmail string) (map[string]bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]bool{}
	for k := range r.approved {
		parts := strings.SplitN(k, "|", 3)
		if len(parts) >= 2 && parts[0] == userEmail {
			out[parts[1]] = true
		}
	}
	return out, nil
}
func (r *fakeConsentRepo) ClientIDsWithAnyConsent() (map[string]bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]bool{}
	for k := range r.approved {
		parts := strings.SplitN(k, "|", 3)
		if len(parts) >= 2 {
			out[parts[1]] = true
		}
	}
	return out, nil
}
func (r *fakeConsentRepo) DeleteUserClient(userEmail, clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	prefix := userEmail + "|" + clientID + "|"
	for k := range r.approved {
		if strings.HasPrefix(k, prefix) {
			delete(r.approved, k)
		}
	}
	return nil
}
func (r *fakeConsentRepo) DeleteByClient(clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k := range r.approved {
		if strings.Contains(k, "|"+clientID+"|") {
			delete(r.approved, k)
		}
	}
	return nil
}

// newTestServer builds a server with consent pre-approved, so existing
// auth-code tests skip the consent page. Consent-flow tests build their own.
func newTestServer(t *testing.T, email string) (*Server, *fakeClientRepo, *fakeRefreshRepo) {
	t.Helper()
	clients := newFakeClientRepo()
	refreshes := newFakeRefreshRepo()
	srv := NewServer(clients, refreshes, alwaysApproved{}, fakeSession{email: email}, "test-secret-key-1234567890abcdef")
	return srv, clients, refreshes
}

type alwaysApproved struct{}

func (alwaysApproved) Has(string, string, string) (bool, error)              { return true, nil }
func (alwaysApproved) Upsert(*persistence.OAuthConsent) error                { return nil }
func (alwaysApproved) ApprovedClientIDs(string) (map[string]bool, error)     { return nil, nil }
func (alwaysApproved) ClientIDsWithAnyConsent() (map[string]bool, error)     { return nil, nil }
func (alwaysApproved) DeleteUserClient(string, string) error                 { return nil }
func (alwaysApproved) DeleteByClient(string) error                           { return nil }

func registerClient(t *testing.T, srv *Server, redirect string) (id, secret string) {
	t.Helper()
	body := strings.NewReader(`{"redirect_uris":["` + redirect + `"],"client_name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp/oauth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.HandleRegister(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register: status %d, body %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse register response: %v", err)
	}
	return resp["client_id"].(string), resp["client_secret"].(string)
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestAuthorizationCodeFlow(t *testing.T) {
	srv, _, refreshes := newTestServer(t, "user@example.com")
	clientID, clientSecret := registerClient(t, srv, "http://localhost:33418/cb")

	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	challenge := pkceChallenge(verifier)

	// /authorize
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:33418/cb"},
		"response_type":         {"code"},
		"state":                 {"xyz"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, authURL, nil)
	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("authorize: expected 302, got %d (%s)", rr.Code, rr.Body.String())
	}
	loc, err := url.Parse(rr.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse Location: %v", err)
	}
	code := loc.Query().Get("code")
	if code == "" {
		t.Fatalf("authorize: missing code in redirect, got %s", loc.String())
	}
	if loc.Query().Get("state") != "xyz" {
		t.Fatalf("state not echoed back: %q", loc.Query().Get("state"))
	}

	// /token (authorization_code)
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"http://localhost:33418/cb"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code_verifier": {verifier},
	}
	tokReq := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(form.Encode()))
	tokReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokRec := httptest.NewRecorder()
	srv.HandleToken(tokRec, tokReq)
	if tokRec.Code != http.StatusOK {
		t.Fatalf("token: expected 200, got %d (%s)", tokRec.Code, tokRec.Body.String())
	}
	var tokResp map[string]any
	if err := json.Unmarshal(tokRec.Body.Bytes(), &tokResp); err != nil {
		t.Fatalf("parse token response: %v", err)
	}
	access, _ := tokResp["access_token"].(string)
	refresh, _ := tokResp["refresh_token"].(string)
	if access == "" || refresh == "" {
		t.Fatalf("missing tokens: %v", tokResp)
	}

	// Validate access token
	email, ok := srv.ValidateAccessToken(access)
	if !ok || email != "user@example.com" {
		t.Fatalf("ValidateAccessToken: ok=%v email=%q", ok, email)
	}

	// /token (refresh_token) - should rotate
	refreshForm := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refresh},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	rfReq := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(refreshForm.Encode()))
	rfReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rfRec := httptest.NewRecorder()
	srv.HandleToken(rfRec, rfReq)
	if rfRec.Code != http.StatusOK {
		t.Fatalf("refresh: expected 200, got %d (%s)", rfRec.Code, rfRec.Body.String())
	}
	var rfResp map[string]any
	_ = json.Unmarshal(rfRec.Body.Bytes(), &rfResp)
	newRefresh, _ := rfResp["refresh_token"].(string)
	if newRefresh == "" || newRefresh == refresh {
		t.Fatalf("refresh did not rotate: old=%q new=%q", refresh, newRefresh)
	}

	// Old refresh token is revoked. Per RFC 6749 §5.2 we return 400
	// invalid_grant. The MCP SDK reacts by invalidating only its tokens
	// (not its client_info) and runs a fresh auth flow against the same
	// client_id - no re-registration spam.
	rfReq2 := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(refreshForm.Encode()))
	rfReq2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rfRec2 := httptest.NewRecorder()
	srv.HandleToken(rfRec2, rfReq2)
	if rfRec2.Code != http.StatusBadRequest {
		t.Fatalf("expected revoked refresh token to be rejected with 400 invalid_grant, got %d", rfRec2.Code)
	}
	if !strings.Contains(rfRec2.Body.String(), "invalid_grant") {
		t.Fatalf("expected invalid_grant error, got body: %s", rfRec2.Body.String())
	}

	// Sanity: the refresh repo holds two records, one revoked.
	if got := len(refreshes.tokens); got != 2 {
		t.Fatalf("expected 2 refresh records, got %d", got)
	}
}

func TestAuthorizationRequiresPKCE(t *testing.T) {
	srv, _, _ := newTestServer(t, "user@example.com")
	clientID, _ := registerClient(t, srv, "http://localhost:33418/cb")

	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {"http://localhost:33418/cb"},
		"response_type": {"code"},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, authURL, nil)
	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected redirect with error, got %d", rr.Code)
	}
	loc, _ := url.Parse(rr.Header().Get("Location"))
	if loc.Query().Get("error") != "invalid_request" {
		t.Fatalf("expected invalid_request error, got %q", loc.Query().Get("error"))
	}
}

func TestAuthorizeWithoutSessionRedirectsToLogin(t *testing.T) {
	srv, _, _ := newTestServer(t, "")
	clientID, _ := registerClient(t, srv, "http://localhost:33418/cb")

	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:33418/cb"},
		"response_type":         {"code"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, authURL, nil)
	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "/auth/login?") {
		t.Fatalf("expected login redirect, got %q", loc)
	}
}

func TestPKCEVerificationFails(t *testing.T) {
	srv, _, _ := newTestServer(t, "user@example.com")
	clientID, clientSecret := registerClient(t, srv, "http://localhost:33418/cb")

	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:33418/cb"},
		"response_type":         {"code"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, authURL, nil)
	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, req)
	loc, _ := url.Parse(rr.Header().Get("Location"))
	code := loc.Query().Get("code")

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"http://localhost:33418/cb"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code_verifier": {"wrong-verifier"},
	}
	tokReq := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(form.Encode()))
	tokReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokRec := httptest.NewRecorder()
	srv.HandleToken(tokRec, tokReq)
	if tokRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad PKCE, got %d (%s)", tokRec.Code, tokRec.Body.String())
	}
}

func TestAccessTokenRejectedAfterClientDeleted(t *testing.T) {
	srv, clients, _ := newTestServer(t, "user@example.com")
	clientID, clientSecret := registerClient(t, srv, "http://localhost:33418/cb")

	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:33418/cb"},
		"response_type":         {"code"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, authURL, nil)
	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, req)
	loc, _ := url.Parse(rr.Header().Get("Location"))
	code := loc.Query().Get("code")

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"http://localhost:33418/cb"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code_verifier": {verifier},
	}
	tokReq := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(form.Encode()))
	tokReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokRec := httptest.NewRecorder()
	srv.HandleToken(tokRec, tokReq)
	if tokRec.Code != http.StatusOK {
		t.Fatalf("token: %d (%s)", tokRec.Code, tokRec.Body.String())
	}
	var tokResp map[string]any
	_ = json.Unmarshal(tokRec.Body.Bytes(), &tokResp)
	access := tokResp["access_token"].(string)

	// Pre-deletion: token works.
	if _, ok := srv.ValidateAccessToken(access); !ok {
		t.Fatalf("access token should validate before client is deleted")
	}

	// Delete the client.
	if err := clients.Delete(clientID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Post-deletion: same JWT must be rejected, even though signature
	// and expiry are still valid - the FindByID lookup fails.
	if _, ok := srv.ValidateAccessToken(access); ok {
		t.Fatalf("access token should be rejected after client is deleted")
	}
}

func TestAuthorizeDedupesRapidRetries(t *testing.T) {
	// MCP SDK's auth() routinely fires several /authorize calls in
	// parallel after a 401, each with a fresh PKCE verifier. The client
	// only persists the LAST verifier. Without dedupe, the codes from
	// earlier calls have stored challenges that no verifier can satisfy
	// (they got overwritten on disk), so /token always rejects them and
	// the only flow that succeeds is the one whose code matches the
	// latest saved verifier. We dedupe per (client, user) so the latest
	// verifier always validates.
	srv, _, _ := newTestServer(t, "user@example.com")
	clientID, clientSecret := registerClient(t, srv, "http://localhost:33418/cb")

	v1 := "verifier-1234567890abcdef-1234567890abcdef"
	v2 := "verifier-abcdef1234567890-abcdef1234567890"
	v3 := "verifier-fedcba0987654321-fedcba0987654321"

	build := func(v string) *http.Request {
		return httptest.NewRequest(http.MethodGet, "/mcp/oauth/authorize?"+url.Values{
			"client_id":             {clientID},
			"redirect_uri":          {"http://localhost:33418/cb"},
			"response_type":         {"code"},
			"code_challenge":        {pkceChallenge(v)},
			"code_challenge_method": {"S256"},
		}.Encode(), nil)
	}

	codeOf := func(rr *httptest.ResponseRecorder) string {
		t.Helper()
		if rr.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d (%s)", rr.Code, rr.Body.String())
		}
		loc, err := url.Parse(rr.Header().Get("Location"))
		if err != nil {
			t.Fatalf("parse location: %v", err)
		}
		return loc.Query().Get("code")
	}

	rr1 := httptest.NewRecorder()
	srv.HandleAuthorize(rr1, build(v1))
	c1 := codeOf(rr1)

	rr2 := httptest.NewRecorder()
	srv.HandleAuthorize(rr2, build(v2))
	c2 := codeOf(rr2)

	rr3 := httptest.NewRecorder()
	srv.HandleAuthorize(rr3, build(v3))
	c3 := codeOf(rr3)

	if c1 != c2 || c2 != c3 {
		t.Fatalf("expected same code across rapid retries, got %q %q %q", c1, c2, c3)
	}

	// /token with the LATEST verifier must succeed even though three
	// /authorize calls happened, because we updated the stored challenge
	// to the most recent one each time.
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {c3},
		"redirect_uri":  {"http://localhost:33418/cb"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code_verifier": {v3},
	}
	tokReq := httptest.NewRequest(http.MethodPost, "/mcp/oauth/token", strings.NewReader(form.Encode()))
	tokReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokRec := httptest.NewRecorder()
	srv.HandleToken(tokRec, tokReq)
	if tokRec.Code != http.StatusOK {
		t.Fatalf("expected /token to accept latest verifier, got %d (%s)", tokRec.Code, tokRec.Body.String())
	}
}

func TestMetadataDocuments(t *testing.T) {
	srv, _, _ := newTestServer(t, "")
	req := httptest.NewRequest(http.MethodGet, "https://qaynaq.example.com/.well-known/oauth-authorization-server", nil)
	req.Host = "qaynaq.example.com"
	rr := httptest.NewRecorder()
	srv.HandleAuthorizationServerMetadata(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("metadata: %d", rr.Code)
	}
	var meta map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &meta)
	if got := meta["authorization_endpoint"]; got == nil || !strings.HasSuffix(got.(string), "/mcp/oauth/authorize") {
		t.Fatalf("authorization_endpoint missing or wrong: %v", got)
	}
	if got := meta["code_challenge_methods_supported"]; got == nil {
		t.Fatalf("code_challenge_methods_supported missing")
	}
}

func TestConsentFlow(t *testing.T) {
	clients := newFakeClientRepo()
	refreshes := newFakeRefreshRepo()
	consents := newFakeConsentRepo()
	srv := NewServer(clients, refreshes, consents, fakeSession{email: "user@example.com"}, "test-secret-key-1234567890abcdef")

	clientID, _ := registerClient(t, srv, "http://localhost:33418/cb")
	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:33418/cb"},
		"response_type":         {"code"},
		"state":                 {"st1"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}.Encode()

	// First /authorize: no prior consent -> 302 to /oauth/consent (SPA).
	rr1 := httptest.NewRecorder()
	srv.HandleAuthorize(rr1, httptest.NewRequest(http.MethodGet, authURL, nil))
	if rr1.Code != http.StatusFound {
		t.Fatalf("expected redirect to consent, got %d", rr1.Code)
	}
	loc, _ := url.Parse(rr1.Header().Get("Location"))
	if loc.Path != "/oauth/consent" {
		t.Fatalf("expected /oauth/consent, got %q", loc.Path)
	}
	requestID := loc.Query().Get("request_id")
	if requestID == "" {
		t.Fatalf("missing request_id in consent URL: %s", loc.String())
	}

	// SPA fetches consent metadata via JSON endpoint.
	getReq := httptest.NewRequest(http.MethodGet, "/v0/mcp/oauth/consent-request?request_id="+requestID, nil)
	getRr := httptest.NewRecorder()
	srv.HandleConsentRequest(getRr, getReq)
	if getRr.Code != http.StatusOK {
		t.Fatalf("consent GET: expected 200, got %d (%s)", getRr.Code, getRr.Body.String())
	}
	var meta map[string]any
	if err := json.Unmarshal(getRr.Body.Bytes(), &meta); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if meta["client_id"] != clientID {
		t.Fatalf("metadata missing client_id, got %v", meta)
	}

	// SPA POSTs allow decision.
	body := `{"request_id":"` + requestID + `","decision":"allow"}`
	postReq := httptest.NewRequest(http.MethodPost, "/v0/mcp/oauth/consent-request", strings.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	postRr := httptest.NewRecorder()
	srv.HandleConsentRequest(postRr, postReq)
	if postRr.Code != http.StatusOK {
		t.Fatalf("consent POST allow: expected 200, got %d (%s)", postRr.Code, postRr.Body.String())
	}
	var allowResp map[string]string
	_ = json.Unmarshal(postRr.Body.Bytes(), &allowResp)
	cb, err := url.Parse(allowResp["redirect_url"])
	if err != nil || cb.Host != "localhost:33418" || cb.Path != "/cb" {
		t.Fatalf("expected redirect_url to client callback, got %v", allowResp)
	}
	if cb.Query().Get("code") == "" {
		t.Fatalf("expected code in callback URL, got %s", cb.String())
	}
	if cb.Query().Get("state") != "st1" {
		t.Fatalf("state must round-trip, got %q", cb.Query().Get("state"))
	}

	// Consent persisted: a second /authorize should skip the page.
	rr2 := httptest.NewRecorder()
	srv.HandleAuthorize(rr2, httptest.NewRequest(http.MethodGet, authURL, nil))
	if rr2.Code != http.StatusFound {
		t.Fatalf("second authorize: expected 302, got %d", rr2.Code)
	}
	loc2, _ := url.Parse(rr2.Header().Get("Location"))
	if loc2.Host != "localhost:33418" {
		t.Fatalf("second authorize should skip consent and redirect to callback, got %s", loc2.String())
	}
}

func TestConsentDeny(t *testing.T) {
	clients := newFakeClientRepo()
	refreshes := newFakeRefreshRepo()
	consents := newFakeConsentRepo()
	srv := NewServer(clients, refreshes, consents, fakeSession{email: "user@example.com"}, "test-secret-key-1234567890abcdef")

	clientID, _ := registerClient(t, srv, "http://localhost:33418/cb")
	verifier := "verifier-1234567890abcdef-1234567890abcdef"
	authURL := "/mcp/oauth/authorize?" + url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:33418/cb"},
		"response_type":         {"code"},
		"state":                 {"st2"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}.Encode()

	rr := httptest.NewRecorder()
	srv.HandleAuthorize(rr, httptest.NewRequest(http.MethodGet, authURL, nil))
	loc, _ := url.Parse(rr.Header().Get("Location"))
	requestID := loc.Query().Get("request_id")

	body := `{"request_id":"` + requestID + `","decision":"deny"}`
	req := httptest.NewRequest(http.MethodPost, "/v0/mcp/oauth/consent-request", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	denyRr := httptest.NewRecorder()
	srv.HandleConsentRequest(denyRr, req)
	if denyRr.Code != http.StatusOK {
		t.Fatalf("expected 200 from deny, got %d", denyRr.Code)
	}
	var denyResp map[string]string
	_ = json.Unmarshal(denyRr.Body.Bytes(), &denyResp)
	cb, err := url.Parse(denyResp["redirect_url"])
	if err != nil {
		t.Fatalf("parse redirect_url: %v", err)
	}
	if cb.Query().Get("error") != "access_denied" {
		t.Fatalf("expected error=access_denied, got %q", cb.Query().Get("error"))
	}
	if cb.Query().Get("code") != "" {
		t.Fatalf("must not include code on deny, got %s", cb.String())
	}
	if has, _ := consents.Has("user@example.com", clientID, ""); has {
		t.Fatalf("deny must not record consent")
	}
}
