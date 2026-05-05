package connection

import (
	"strings"
	"testing"
)

func TestProviderAcceptsNoScopes(t *testing.T) {
	cases := map[string]bool{
		"asana_mcp": true,
		"asana":     false,
		"slack":     false,
		"google":    false,
		"unknown":   false,
	}
	for provider, want := range cases {
		if got := ProviderAcceptsNoScopes(provider); got != want {
			t.Errorf("ProviderAcceptsNoScopes(%q) = %v, want %v", provider, got, want)
		}
	}
}

func TestProviderRequiresPKCE(t *testing.T) {
	cases := map[string]bool{
		"github_mcp": true,
		"sentry":     true,
		"shopify":    true,
		"github":     false,
		"google":     false,
		"slack":      false,
	}
	for provider, want := range cases {
		if got := ProviderRequiresPKCE(provider); got != want {
			t.Errorf("ProviderRequiresPKCE(%q) = %v, want %v", provider, got, want)
		}
	}
}

func TestProviderUsesShopTemplate(t *testing.T) {
	if !ProviderUsesShopTemplate("shopify") {
		t.Errorf("expected shopify to use shop template")
	}
	if ProviderUsesShopTemplate("google") {
		t.Errorf("expected google not to use shop template")
	}
}

func TestProviderRequiresCloudID(t *testing.T) {
	if !ProviderRequiresCloudID("atlassian") {
		t.Errorf("expected atlassian to require cloud_id")
	}
	if ProviderRequiresCloudID("google") {
		t.Errorf("expected google not to require cloud_id")
	}
}

func TestGetDisplayName(t *testing.T) {
	cases := map[string]string{
		"asana_mcp":  "Asana MCP",
		"github_mcp": "GitHub MCP",
		"atlassian":  "Atlassian (Jira & Confluence)",
		// unregistered: falls back to slug
		"unknown_provider": "unknown_provider",
	}
	for provider, want := range cases {
		if got := GetDisplayName(provider); got != want {
			t.Errorf("GetDisplayName(%q) = %q, want %q", provider, got, want)
		}
	}
}

func TestValidateShopSubdomain(t *testing.T) {
	valid := []string{"my-store", "store123", "a", "abc-def-ghi"}
	invalid := []string{
		"",                                 // empty
		"My-Store",                         // uppercase
		"my_store",                         // underscore
		"my.store",                         // dot
		"-leading-dash",                    // leading dash
		"my-store/admin",                   // slash injection
		"evil.com",                         // dot
		strings.Repeat("a", 64),            // too long
		"my-store?query=injection",         // query string
	}
	for _, s := range valid {
		if err := ValidateShopSubdomain(s); err != nil {
			t.Errorf("ValidateShopSubdomain(%q) = %v, want nil", s, err)
		}
	}
	for _, s := range invalid {
		if err := ValidateShopSubdomain(s); err == nil {
			t.Errorf("ValidateShopSubdomain(%q) = nil, want error", s)
		}
	}
}

func TestValidateCloudID(t *testing.T) {
	valid := []string{
		"00000000-0000-0000-0000-000000000000",
		"1324a887-45db-1bf4-1e99-ef0ff456d987",
		"DEADBEEF-DEAD-BEEF-DEAD-BEEFDEADBEEF", // upper case OK
	}
	invalid := []string{
		"",
		"not-a-uuid",
		"1324a887-45db-1bf4-1e99",                          // too short
		"1324a887-45db-1bf4-1e99-ef0ff456d987-extra",       // too long
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",             // non-hex
		"1324a887_45db_1bf4_1e99_ef0ff456d987",             // wrong separator
	}
	for _, s := range valid {
		if err := ValidateCloudID(s); err != nil {
			t.Errorf("ValidateCloudID(%q) = %v, want nil", s, err)
		}
	}
	for _, s := range invalid {
		if err := ValidateCloudID(s); err == nil {
			t.Errorf("ValidateCloudID(%q) = nil, want error", s)
		}
	}
}

func TestSubstituteEndpointTemplate(t *testing.T) {
	ep, err := GetEndpoint("shopify")
	if err != nil {
		t.Fatalf("GetEndpoint(shopify): %v", err)
	}
	if !strings.Contains(ep.AuthURL, "{shop}") {
		t.Fatalf("expected shopify auth URL to contain {shop} placeholder, got %q", ep.AuthURL)
	}

	subbed := substituteEndpointTemplate(ep, "my-store")
	if strings.Contains(subbed.AuthURL, "{shop}") {
		t.Errorf("substitution didn't replace {shop} in AuthURL: %q", subbed.AuthURL)
	}
	if strings.Contains(subbed.TokenURL, "{shop}") {
		t.Errorf("substitution didn't replace {shop} in TokenURL: %q", subbed.TokenURL)
	}
	if !strings.Contains(subbed.AuthURL, "my-store.myshopify.com") {
		t.Errorf("expected my-store.myshopify.com in AuthURL, got %q", subbed.AuthURL)
	}
}

func TestGetProvidersIncludesScopelessProviders(t *testing.T) {
	providers := GetProviders()
	if _, ok := providers["asana_mcp"]; !ok {
		t.Errorf("asana_mcp missing from GetProviders() despite being scopeless")
	}
	if scopes := providers["asana_mcp"]; len(scopes) != 0 {
		t.Errorf("asana_mcp should have no scopes, got %d", len(scopes))
	}
}

func TestAllProvidersHaveDisplayNames(t *testing.T) {
	for slug := range providerEndpoints {
		if _, ok := ProviderDisplayNames[slug]; !ok {
			t.Errorf("provider %q has endpoint but no display name", slug)
		}
	}
}

func TestAllProvidersHaveSetups(t *testing.T) {
	for slug := range providerEndpoints {
		if _, ok := ProviderSetups[slug]; !ok {
			t.Errorf("provider %q has endpoint but no setup info", slug)
		}
	}
}

func TestPKCEVerifierLength(t *testing.T) {
	v := generateCodeVerifier()
	// RFC 7636: verifier must be 43-128 chars from [A-Z][a-z][0-9]-._~
	if len(v) < 43 || len(v) > 128 {
		t.Errorf("verifier length %d, want 43-128", len(v))
	}
}

func TestPKCEChallengeStable(t *testing.T) {
	// Known test vector from RFC 7636 Appendix B
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	got := computeCodeChallenge(verifier)
	if got != want {
		t.Errorf("computeCodeChallenge(%q) = %q, want %q", verifier, got, want)
	}
}

func TestPKCEVerifierIsRandom(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v := generateCodeVerifier()
		if seen[v] {
			t.Errorf("duplicate verifier generated: %q", v)
		}
		seen[v] = true
	}
}
