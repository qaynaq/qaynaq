// Package oauth implements the OAuth 2.1 Authorization Server endpoints
// required by the MCP spec (2025-06-18) so that MCP clients such as Claude
// Desktop, Cursor, or Continue.dev can authenticate against Qaynaq.
//
// The user-facing /authorize step delegates to the host application's
// existing auth (cookie-based session for OAuth2 mode, or open access for
// AuthTypeNone). Tokens issued to MCP clients are minted by Qaynaq itself,
// so they have a Qaynaq audience and can be revoked independently of the
// upstream IdP.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/qaynaq/qaynaq/internal/persistence"
)

const (
	accessTokenTTL  = 1 * time.Hour
	refreshTokenTTL = 30 * 24 * time.Hour
	authCodeTTL     = 60 * time.Second

	tokenAudience = "qaynaq-mcp"
	tokenIssuer   = "qaynaq"
)

// SessionResolver returns the authenticated user (email) associated with the
// incoming HTTP request, sourced from the host application's session cookie.
// When the host runs without auth, it returns an anonymous identity.
type SessionResolver interface {
	ResolveUser(r *http.Request) (email string, ok bool)
	AuthType() config.AuthType
	LoginRedirectPath() string
}

type authCode struct {
	clientID            string
	userEmail           string
	redirectURI         string
	codeChallenge       string
	codeChallengeMethod string
	scope               string
	expiresAt           time.Time
}

// pendingRequest is an authorize request awaiting user consent. Stored
// in-memory for the lifetime of the consent page (~10 minutes).
type pendingRequest struct {
	clientID            string
	clientName          string
	redirectURI         string
	state               string
	codeChallenge       string
	codeChallengeMethod string
	scope               string
	expiresAt           time.Time
}

const consentRequestTTL = 10 * time.Minute

type Server struct {
	clientRepo  persistence.OAuthClientRepository
	refreshRepo persistence.OAuthRefreshTokenRepository
	consentRepo persistence.OAuthConsentRepository
	session     SessionResolver
	jwtSecret   []byte

	codes        map[string]authCode
	codesByOwner map[string]string
	codesMu      sync.Mutex

	pending   map[string]pendingRequest
	pendingMu sync.Mutex

	stopCleanup chan struct{}
	cleanupDone chan struct{}
}

func NewServer(
	clientRepo persistence.OAuthClientRepository,
	refreshRepo persistence.OAuthRefreshTokenRepository,
	consentRepo persistence.OAuthConsentRepository,
	session SessionResolver,
	jwtSecret string,
) *Server {
	s := &Server{
		clientRepo:   clientRepo,
		refreshRepo:  refreshRepo,
		consentRepo:  consentRepo,
		session:      session,
		jwtSecret:    []byte(jwtSecret),
		codes:        make(map[string]authCode),
		codesByOwner: make(map[string]string),
		pending:      make(map[string]pendingRequest),
		stopCleanup:  make(chan struct{}),
		cleanupDone:  make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

// Shutdown stops background goroutines. Safe to call once. Blocks until the
// cleanup loop has returned.
func (s *Server) Shutdown(ctx context.Context) {
	select {
	case <-s.stopCleanup:
		return
	default:
	}
	close(s.stopCleanup)
	select {
	case <-s.cleanupDone:
	case <-ctx.Done():
	}
}

// MountRoutes registers the public OAuth AS endpoints (no Qaynaq auth middleware).
func (s *Server) MountRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/.well-known/oauth-authorization-server", s.HandleAuthorizationServerMetadata)
	mux.HandleFunc("/.well-known/oauth-protected-resource", s.HandleProtectedResourceMetadata)
	mux.HandleFunc("/mcp/oauth/register", s.HandleRegister)
	mux.HandleFunc("/mcp/oauth/authorize", s.HandleAuthorize)
	mux.HandleFunc("/mcp/oauth/token", s.HandleToken)
	mux.HandleFunc("/mcp/oauth/revoke", s.HandleRevoke)
}

// MountAPIRoutes registers the SPA-facing JSON endpoints. These run behind
// the regular Qaynaq auth middleware so the user must be signed in.
func (s *Server) MountAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v0/mcp/oauth/consent-request", s.HandleConsentRequest)
}

// ValidateAccessToken returns the user email bound to a JWT issued by /token.
// Rejects tokens whose client has been deleted, so deletion takes effect on
// the next MCP request.
func (s *Server) ValidateAccessToken(raw string) (email string, ok bool) {
	tok, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return "", false
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	if aud, _ := claims["aud"].(string); aud != tokenAudience {
		return "", false
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", false
	}
	clientID, _ := claims["client_id"].(string)
	if clientID == "" {
		return "", false
	}
	if _, err := s.clientRepo.FindByID(clientID); err != nil {
		return "", false
	}
	return sub, true
}

func (s *Server) HandleAuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	issuer := s.issuerURL(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/mcp/oauth/authorize",
		"token_endpoint":                        issuer + "/mcp/oauth/token",
		"registration_endpoint":                 issuer + "/mcp/oauth/register",
		"revocation_endpoint":                   issuer + "/mcp/oauth/revoke",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic", "none"},
		"scopes_supported":                      []string{"mcp"},
	})
}

func (s *Server) HandleProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	issuer := s.issuerURL(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"resource":              issuer + "/mcp",
		"authorization_servers": []string{issuer},
		"scopes_supported":      []string{"mcp"},
		"bearer_methods_supported": []string{"header"},
	})
}

type registerRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeOAuthError(w, http.StatusMethodNotAllowed, "invalid_request", "POST required")
		return
	}
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_client_metadata", "could not parse request body")
		return
	}
	if len(req.RedirectURIs) == 0 {
		writeOAuthError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris is required")
		return
	}
	for _, u := range req.RedirectURIs {
		if _, err := url.Parse(u); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_redirect_uri", "invalid redirect_uri: "+u)
			return
		}
	}

	clientID, err := randomString(16)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to generate client_id")
		return
	}
	clientID = "mcp_" + clientID
	clientSecret, err := randomString(32)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to generate client_secret")
		return
	}

	name := strings.TrimSpace(req.ClientName)
	if name == "" {
		name = "Unnamed MCP Client"
	}

	client := &persistence.OAuthClient{
		ID:           clientID,
		SecretHash:   hashSecret(clientSecret),
		Name:         name,
		RedirectURIs: req.RedirectURIs,
		CreatedAt:    time.Now(),
	}
	if err := s.clientRepo.Create(client); err != nil {
		log.Error().Err(err).Msg("failed to persist oauth client")
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to register client")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"client_id":                  client.ID,
		"client_secret":              clientSecret,
		"client_id_issued_at":        client.CreatedAt.Unix(),
		"client_secret_expires_at":   0,
		"redirect_uris":              client.RedirectURIs,
		"client_name":                client.Name,
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "client_secret_post",
	})
}

func (s *Server) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")
	scope := q.Get("scope")

	if clientID == "" || redirectURI == "" {
		redirectToErrorPage(w, r, "invalid_request", "Missing client_id or redirect_uri")
		return
	}
	// Per RFC 6749 §3.1.2.4 we must NOT redirect to the client when the
	// client_id or redirect_uri is invalid - the redirect URI can't be
	// trusted yet. Send the user to the SPA error page instead.
	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		redirectToErrorPage(w, r, "stale_client", clientID)
		return
	}
	if !redirectURIAllowed(client.RedirectURIs, redirectURI) {
		redirectToErrorPage(w, r, "invalid_redirect_uri", "redirect_uri does not match any registered URI")
		return
	}

	if responseType != "code" {
		redirectError(w, r, redirectURI, state, "unsupported_response_type", "only code is supported")
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		redirectError(w, r, redirectURI, state, "invalid_request", "PKCE with S256 is required")
		return
	}

	email, ok := s.session.ResolveUser(r)
	if !ok {
		s.redirectToLogin(w, r)
		return
	}

	approved, err := s.consentRepo.Has(email, client.ID, scope)
	if err != nil {
		log.Error().Err(err).Msg("failed to check oauth consent")
		redirectError(w, r, redirectURI, state, "server_error", "failed to check consent")
		return
	}
	if !approved {
		requestID, err := randomString(24)
		if err != nil {
			redirectError(w, r, redirectURI, state, "server_error", "failed to start consent flow")
			return
		}
		s.pendingMu.Lock()
		s.pending[requestID] = pendingRequest{
			clientID:            client.ID,
			clientName:          client.Name,
			redirectURI:         redirectURI,
			state:               state,
			codeChallenge:       codeChallenge,
			codeChallengeMethod: codeChallengeMethod,
			scope:               scope,
			expiresAt:           time.Now().Add(consentRequestTTL),
		}
		s.pendingMu.Unlock()
		http.Redirect(w, r, "/oauth/consent?request_id="+url.QueryEscape(requestID), http.StatusFound)
		return
	}

	s.completeAuthorize(w, r, client, email, redirectURI, state, codeChallenge, codeChallengeMethod, scope)
}

// completeAuthorize issues an auth code and redirects to the client's
// callback. Used by the silent re-consent path of /authorize.
func (s *Server) completeAuthorize(w http.ResponseWriter, r *http.Request, client *persistence.OAuthClient, email, redirectURI, state, codeChallenge, codeChallengeMethod, scope string) {
	dest, err := s.buildAuthorizationCodeRedirect(client, email, redirectURI, state, codeChallenge, codeChallengeMethod, scope)
	if err != nil {
		redirectError(w, r, redirectURI, state, "server_error", "failed to generate code")
		return
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

// buildAuthorizationCodeRedirect issues an auth code (deduping rapid retries
// from the same (client, user) so the latest PKCE verifier on the client
// side keeps validating) and returns the URL to redirect to.
func (s *Server) buildAuthorizationCodeRedirect(client *persistence.OAuthClient, email, redirectURI, state, codeChallenge, codeChallengeMethod, scope string) (string, error) {
	ownerKey := client.ID + "|" + email
	s.codesMu.Lock()
	code, reused := s.codesByOwner[ownerKey]
	if reused {
		if existing, ok := s.codes[code]; ok && time.Now().Before(existing.expiresAt) {
			existing.codeChallenge = codeChallenge
			existing.codeChallengeMethod = codeChallengeMethod
			existing.redirectURI = redirectURI
			existing.scope = scope
			s.codes[code] = existing
		} else {
			reused = false
		}
	}
	if !reused {
		var err error
		code, err = randomString(32)
		if err != nil {
			s.codesMu.Unlock()
			return "", err
		}
		s.codes[code] = authCode{
			clientID:            client.ID,
			userEmail:           email,
			redirectURI:         redirectURI,
			codeChallenge:       codeChallenge,
			codeChallengeMethod: codeChallengeMethod,
			scope:               scope,
			expiresAt:           time.Now().Add(authCodeTTL),
		}
		s.codesByOwner[ownerKey] = code
	}
	s.codesMu.Unlock()

	dest, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}
	values := dest.Query()
	values.Set("code", code)
	if state != "" {
		values.Set("state", state)
	}
	dest.RawQuery = values.Encode()
	return dest.String(), nil
}

// buildErrorRedirect returns the redirect URL for an OAuth error response,
// preserving state. Returns an empty string if the redirect URI is unparseable.
func buildErrorRedirect(redirectURI, state, code, description string) string {
	dest, err := url.Parse(redirectURI)
	if err != nil {
		return ""
	}
	values := dest.Query()
	values.Set("error", code)
	values.Set("error_description", description)
	if state != "" {
		values.Set("state", state)
	}
	dest.RawQuery = values.Encode()
	return dest.String()
}

// HandleConsentRequest serves the SPA-facing JSON endpoints behind
// /api/v0/mcp/oauth/consent-request. GET fetches metadata for the consent
// page, POST records the user's decision and returns a redirect URL.
func (s *Server) HandleConsentRequest(w http.ResponseWriter, r *http.Request) {
	email, ok := s.session.ResolveUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleConsentRequestGet(w, r, email)
	case http.MethodPost:
		s.handleConsentRequestPost(w, r, email)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleConsentRequestGet(w http.ResponseWriter, r *http.Request, email string) {
	requestID := r.URL.Query().Get("request_id")
	if requestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing request_id"})
		return
	}
	s.pendingMu.Lock()
	req, found := s.pending[requestID]
	s.pendingMu.Unlock()
	if !found || time.Now().After(req.expiresAt) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "authorization request not found or expired"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"request_id":   requestID,
		"client_id":    req.clientID,
		"client_name":  req.clientName,
		"redirect_uri": req.redirectURI,
		"scope":        req.scope,
		"user_email":   email,
	})
}

type consentDecisionRequest struct {
	RequestID string `json:"request_id"`
	Decision  string `json:"decision"`
}

func (s *Server) handleConsentRequestPost(w http.ResponseWriter, r *http.Request, email string) {
	var body consentDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if body.RequestID == "" || (body.Decision != "allow" && body.Decision != "deny") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request_id and decision (allow|deny) are required"})
		return
	}

	s.pendingMu.Lock()
	req, found := s.pending[body.RequestID]
	if found {
		delete(s.pending, body.RequestID)
	}
	s.pendingMu.Unlock()
	if !found || time.Now().After(req.expiresAt) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "authorization request not found or expired"})
		return
	}

	if body.Decision == "deny" {
		writeJSON(w, http.StatusOK, map[string]string{
			"redirect_url": buildErrorRedirect(req.redirectURI, req.state, "access_denied", "user denied the request"),
		})
		return
	}

	if err := s.consentRepo.Upsert(&persistence.OAuthConsent{
		UserEmail:  email,
		ClientID:   req.clientID,
		Scope:      req.scope,
		ApprovedAt: time.Now(),
	}); err != nil {
		log.Error().Err(err).Msg("failed to persist oauth consent")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to record consent"})
		return
	}

	client, err := s.clientRepo.FindByID(req.clientID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "client no longer exists"})
		return
	}
	redirectURL, err := s.buildAuthorizationCodeRedirect(client, email, req.redirectURI, req.state, req.codeChallenge, req.codeChallengeMethod, req.scope)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue code"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"redirect_url": redirectURL})
}

func (s *Server) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeOAuthError(w, http.StatusMethodNotAllowed, "invalid_request", "POST required")
		return
	}
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "could not parse form")
		return
	}
	grantType := r.Form.Get("grant_type")

	switch grantType {
	case "authorization_code":
		s.handleAuthCodeGrant(w, r)
	case "refresh_token":
		s.handleRefreshGrant(w, r)
	default:
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "unsupported grant_type")
	}
}

func (s *Server) handleAuthCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.Form.Get("code")
	redirectURI := r.Form.Get("redirect_uri")
	codeVerifier := r.Form.Get("code_verifier")
	clientID, clientSecret := extractClientCredentials(r)

	if code == "" || redirectURI == "" || codeVerifier == "" || clientID == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "missing required parameters")
		return
	}

	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "unknown client")
		return
	}
	if clientSecret != "" && !verifySecret(client.SecretHash, clientSecret) {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "invalid client_secret")
		return
	}

	s.codesMu.Lock()
	ac, ok := s.codes[code]
	if ok {
		delete(s.codes, code)
		delete(s.codesByOwner, ac.clientID+"|"+ac.userEmail)
	}
	s.codesMu.Unlock()
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "unknown or used code")
		return
	}
	if time.Now().After(ac.expiresAt) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}
	if ac.clientID != clientID {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "code/client mismatch")
		return
	}
	if ac.redirectURI != redirectURI {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}
	if !verifyPKCE(ac.codeChallenge, codeVerifier) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
		return
	}

	s.issueTokens(w, r, client, ac.userEmail)
}

func (s *Server) handleRefreshGrant(w http.ResponseWriter, r *http.Request) {
	refreshTokenRaw := r.Form.Get("refresh_token")
	clientID, clientSecret := extractClientCredentials(r)

	if refreshTokenRaw == "" || clientID == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "missing required parameters")
		return
	}
	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "unknown client")
		return
	}
	if clientSecret != "" && !verifySecret(client.SecretHash, clientSecret) {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "invalid client_secret")
		return
	}

	hash := hashSecret(refreshTokenRaw)
	stored, err := s.refreshRepo.FindByHash(hash)
	if err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "refresh_token not recognized")
		return
	}
	if stored.RevokedAt != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "refresh_token revoked")
		return
	}
	if time.Now().After(stored.ExpiresAt) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "refresh_token expired")
		return
	}
	if stored.ClientID != clientID {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "refresh_token does not belong to client")
		return
	}

	// Rotate: revoke the presented refresh token, then issue a new pair.
	if err := s.refreshRepo.Revoke(stored.ID); err != nil {
		log.Error().Err(err).Msg("failed to revoke refresh token during rotation")
	}
	s.issueTokens(w, r, client, stored.UserEmail)
}

func (s *Server) issueTokens(w http.ResponseWriter, r *http.Request, client *persistence.OAuthClient, userEmail string) {
	now := time.Now()
	accessToken, err := s.signAccessToken(userEmail, client.ID, now)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to sign token")
		return
	}
	refreshRaw, err := randomString(48)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to generate refresh token")
		return
	}
	refreshRecord := &persistence.OAuthRefreshToken{
		TokenHash: hashSecret(refreshRaw),
		ClientID:  client.ID,
		UserEmail: userEmail,
		ExpiresAt: now.Add(refreshTokenTTL),
		CreatedAt: now,
	}
	if err := s.refreshRepo.Create(refreshRecord); err != nil {
		log.Error().Err(err).Msg("failed to persist refresh token")
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to issue refresh token")
		return
	}
	if err := s.clientRepo.UpdateLastUsedAt(client.ID, now); err != nil {
		log.Debug().Err(err).Msg("failed to update oauth client last_used_at")
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(accessTokenTTL.Seconds()),
		"refresh_token": refreshRaw,
		"scope":         "mcp",
	})
}

func (s *Server) signAccessToken(email, clientID string, now time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss":       tokenIssuer,
		"aud":       tokenAudience,
		"sub":       email,
		"client_id": clientID,
		"iat":       now.Unix(),
		"exp":       now.Add(accessTokenTTL).Unix(),
		"scope":     "mcp",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

// ----- Revocation -----

func (s *Server) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "could not parse form")
		return
	}
	token := r.Form.Get("token")
	if token == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if stored, err := s.refreshRepo.FindByHash(hashSecret(token)); err == nil {
		if err := s.refreshRepo.Revoke(stored.ID); err != nil {
			log.Error().Err(err).Msg("failed to revoke refresh token")
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	target := r.URL.RequestURI()
	loginPath := s.session.LoginRedirectPath()
	if loginPath == "" {
		redirectToErrorPage(w, r, "no_login", "Authentication is required to authorize MCP clients, but no login flow is configured.")
		return
	}
	dest := loginPath + "?return_to=" + url.QueryEscape(target)
	http.Redirect(w, r, dest, http.StatusFound)
}

func (s *Server) issuerURL(r *http.Request) string {
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
	return fmt.Sprintf("%s://%s", scheme, host)
}

func (s *Server) cleanupLoop() {
	defer close(s.cleanupDone)
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCleanup:
			return
		case <-ticker.C:
			now := time.Now()
			s.codesMu.Lock()
			for k, v := range s.codes {
				if now.After(v.expiresAt) {
					delete(s.codes, k)
					delete(s.codesByOwner, v.clientID+"|"+v.userEmail)
				}
			}
			s.codesMu.Unlock()
			s.pendingMu.Lock()
			for k, v := range s.pending {
				if now.After(v.expiresAt) {
					delete(s.pending, k)
				}
			}
			s.pendingMu.Unlock()
			if err := s.refreshRepo.DeleteExpired(now.Add(-7 * 24 * time.Hour)); err != nil {
				log.Debug().Err(err).Msg("failed to delete expired refresh tokens")
			}
		}
	}
}

func extractClientCredentials(r *http.Request) (id, secret string) {
	if u, p, ok := r.BasicAuth(); ok {
		return u, p
	}
	return r.Form.Get("client_id"), r.Form.Get("client_secret")
}

func redirectURIAllowed(allowed []string, candidate string) bool {
	for _, a := range allowed {
		if a == candidate {
			return true
		}
	}
	return false
}

func verifyPKCE(challenge, verifier string) bool {
	sum := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(sum[:])
	return got == challenge
}

func verifySecret(hash, raw string) bool {
	return hash == hashSecret(raw)
}

func hashSecret(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error().Err(err).Msg("failed to write JSON response")
	}
}

func writeOAuthError(w http.ResponseWriter, status int, code, description string) {
	writeJSON(w, status, map[string]string{
		"error":             code,
		"error_description": description,
	})
}

// redirectToErrorPage 302s to the SPA error route. Used for OAuth flow
// failures where we cannot safely redirect to the client's redirect_uri
// (unknown client, redirect_uri mismatch, missing query params).
func redirectToErrorPage(w http.ResponseWriter, r *http.Request, code, message string) {
	q := url.Values{}
	q.Set("code", code)
	q.Set("message", message)
	http.Redirect(w, r, "/oauth/error?"+q.Encode(), http.StatusFound)
}

func redirectError(w http.ResponseWriter, r *http.Request, redirectURI, state, code, description string) {
	dest, err := url.Parse(redirectURI)
	if err != nil {
		redirectToErrorPage(w, r, "invalid_request", description)
		return
	}
	values := dest.Query()
	values.Set("error", code)
	values.Set("error_description", description)
	if state != "" {
		values.Set("state", state)
	}
	dest.RawQuery = values.Encode()
	http.Redirect(w, r, dest.String(), http.StatusFound)
}
