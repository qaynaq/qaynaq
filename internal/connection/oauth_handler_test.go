package connection

import (
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// authorizeURLFor builds the authorize URL for a provider with the given
// scopes. Mirrors the relevant portion of HandleAuthorize so the per-provider
// quirks can be asserted without standing up an HTTP server.
func authorizeURLFor(t *testing.T, provider string, scopes []string, shop string) (*url.URL, string) {
	t.Helper()

	endpoint, err := GetEndpoint(provider)
	if err != nil {
		t.Fatalf("GetEndpoint(%q): %v", provider, err)
	}
	if ProviderUsesShopTemplate(provider) {
		endpoint = substituteEndpointTemplate(endpoint, shop)
	}

	cfg := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		Endpoint:     endpoint,
		Scopes:       scopes,
		RedirectURL:  "http://localhost:8080/connections/oauth/callback",
	}

	opts, verifier := buildAuthorizeOpts(provider, cfg)
	rawURL := cfg.AuthCodeURL("test-state", opts...)
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	return u, verifier
}

func TestAuthorize_Slack_UsesUserScope(t *testing.T) {
	u, verifier := authorizeURLFor(t, "slack", []string{"chat:write", "users:read"}, "")

	if u.Query().Get("scope") != "" {
		t.Errorf("slack authorize URL should not have scope=, got %q", u.Query().Get("scope"))
	}
	got := u.Query().Get("user_scope")
	if got != "chat:write,users:read" {
		t.Errorf("user_scope = %q, want %q", got, "chat:write,users:read")
	}
	if verifier != "" {
		t.Errorf("slack should not use PKCE")
	}
}

func TestAuthorize_AsanaMCP_NoScopeWithResource(t *testing.T) {
	u, _ := authorizeURLFor(t, "asana_mcp", nil, "")

	if u.Query().Get("scope") != "" {
		t.Errorf("asana_mcp must NOT include scope= (provider rejects it), got %q", u.Query().Get("scope"))
	}
	want := "https://mcp.asana.com/v2"
	if got := u.Query().Get("resource"); got != want {
		t.Errorf("resource = %q, want %q", got, want)
	}
}

func TestAuthorize_Asana_PromptConsent(t *testing.T) {
	u, _ := authorizeURLFor(t, "asana", []string{"default"}, "")
	if got := u.Query().Get("prompt"); got != "consent" {
		t.Errorf("asana prompt = %q, want consent", got)
	}
}

func TestAuthorize_Atlassian_AudienceAndConsent(t *testing.T) {
	u, _ := authorizeURLFor(t, "atlassian", []string{"offline_access", "read:jira-work"}, "")

	if got := u.Query().Get("audience"); got != "api.atlassian.com" {
		t.Errorf("audience = %q, want api.atlassian.com", got)
	}
	if got := u.Query().Get("prompt"); got != "consent" {
		t.Errorf("prompt = %q, want consent", got)
	}
	scope := u.Query().Get("scope")
	if !strings.Contains(scope, "offline_access") {
		t.Errorf("expected offline_access in scope, got %q", scope)
	}
}

func TestAuthorize_Linear_ActorUser(t *testing.T) {
	u, _ := authorizeURLFor(t, "linear", []string{"read"}, "")
	if got := u.Query().Get("actor"); got != "user" {
		t.Errorf("linear actor = %q, want user", got)
	}
}

func TestAuthorize_Notion_OwnerUser(t *testing.T) {
	u, _ := authorizeURLFor(t, "notion", []string{"read_content"}, "")
	if got := u.Query().Get("owner"); got != "user" {
		t.Errorf("notion owner = %q, want user", got)
	}
}

func TestAuthorize_GitHubMCP_PKCE(t *testing.T) {
	u, verifier := authorizeURLFor(t, "github_mcp", []string{"repo"}, "")

	if got := u.Query().Get("code_challenge_method"); got != "S256" {
		t.Errorf("code_challenge_method = %q, want S256", got)
	}
	challenge := u.Query().Get("code_challenge")
	if challenge == "" {
		t.Errorf("code_challenge missing in github_mcp authorize URL")
	}
	if verifier == "" {
		t.Errorf("expected non-empty PKCE verifier for github_mcp")
	}
	if got := computeCodeChallenge(verifier); got != challenge {
		t.Errorf("code_challenge in URL doesn't match SHA-256 of verifier")
	}
}

func TestAuthorize_SlackMCP_StandardWithPKCE(t *testing.T) {
	u, verifier := authorizeURLFor(t, "slack_mcp", []string{"chat:write", "search:read.public"}, "")

	// Critical: slack_mcp uses Slack's MCP-dedicated endpoint, not the
	// regular /oauth/v2/authorize - so it must NOT mangle scopes into
	// user_scope= the way regular `slack` does.
	if u.Query().Get("user_scope") != "" {
		t.Errorf("slack_mcp must NOT use user_scope= (that's regular slack's quirk)")
	}
	scope := u.Query().Get("scope")
	if !strings.Contains(scope, "chat:write") || !strings.Contains(scope, "search:read.public") {
		t.Errorf("slack_mcp scope = %q, expected standard scope= param", scope)
	}
	// Endpoint must hit the MCP-dedicated authorize URL.
	if u.Path != "/oauth/v2_user/authorize" {
		t.Errorf("slack_mcp authorize path = %q, want /oauth/v2_user/authorize", u.Path)
	}
	// PKCE required per the metadata at .well-known/oauth-authorization-server.
	if u.Query().Get("code_challenge") == "" {
		t.Errorf("slack_mcp should have code_challenge")
	}
	if u.Query().Get("code_challenge_method") != "S256" {
		t.Errorf("slack_mcp code_challenge_method should be S256")
	}
	if verifier == "" {
		t.Errorf("slack_mcp should have PKCE verifier")
	}
}

func TestAuthorize_Sentry_PKCE(t *testing.T) {
	u, verifier := authorizeURLFor(t, "sentry", []string{"org:read"}, "")
	if u.Query().Get("code_challenge") == "" {
		t.Errorf("sentry should have code_challenge")
	}
	if verifier == "" {
		t.Errorf("sentry should have PKCE verifier")
	}
}

func TestAuthorize_Shopify_ShopSubstitutionAndPKCE(t *testing.T) {
	u, verifier := authorizeURLFor(t, "shopify", []string{"read_products"}, "my-store")

	if !strings.Contains(u.Host, "my-store.myshopify.com") {
		t.Errorf("shopify host = %q, expected my-store.myshopify.com", u.Host)
	}
	if u.Query().Get("code_challenge") == "" {
		t.Errorf("shopify should have code_challenge (PKCE)")
	}
	if verifier == "" {
		t.Errorf("shopify should have PKCE verifier")
	}
}

func TestAuthorize_GitHub_NoPKCE(t *testing.T) {
	u, verifier := authorizeURLFor(t, "github", []string{"repo"}, "")
	if u.Query().Get("code_challenge") != "" {
		t.Errorf("github (OAuth App) should NOT have PKCE")
	}
	if verifier != "" {
		t.Errorf("github should not return a verifier")
	}
}

func TestAuthorize_Hubspot_Standard(t *testing.T) {
	u, _ := authorizeURLFor(t, "hubspot", []string{"crm.objects.contacts.read"}, "")
	scope := u.Query().Get("scope")
	if !strings.Contains(scope, "crm.objects.contacts.read") {
		t.Errorf("hubspot scope = %q", scope)
	}
}

func TestAuthorize_Google_Unchanged(t *testing.T) {
	// Regression: Google's existing access_type=offline + prompt=consent still
	// applies after the spec-driven refactor.
	u, _ := authorizeURLFor(t, "google", []string{"https://www.googleapis.com/auth/calendar"}, "")
	if got := u.Query().Get("access_type"); got != "offline" {
		t.Errorf("google access_type = %q, want offline", got)
	}
	if got := u.Query().Get("prompt"); got != "consent" {
		t.Errorf("google prompt = %q, want consent", got)
	}
}
