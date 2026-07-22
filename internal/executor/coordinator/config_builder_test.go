package coordinator

import (
	"testing"

	_ "github.com/qaynaq/qaynaq/internal/components/bloblang"
	"github.com/qaynaq/qaynaq/internal/persistence"
	_ "github.com/warpstreamlabs/bento/public/components/io"
)

func buildProcessorMap(t *testing.T, component, config string) map[string]any {
	t.Helper()
	b := &configBuilder{}
	procMap, err := b.buildProcessorConfig(persistence.FlowProcessor{
		Label:     "test",
		Component: component,
		Config:    []byte(config),
	})
	if err != nil {
		t.Fatalf("buildProcessorConfig failed: %v", err)
	}
	inner, ok := procMap[component].(map[string]any)
	if !ok {
		t.Fatalf("expected map config for %s, got %T", component, procMap[component])
	}
	return inner
}

func TestInjectConnectionAuthStripsFieldAndAddsHeader(t *testing.T) {
	cfg := buildProcessorMap(t, "http", "url: https://api.example.com\noauth_connection: ${QAYNAQ_CONN_slack}\n")

	if _, exists := cfg[oauthConnectionField]; exists {
		t.Error("oauth_connection should be stripped from the built config")
	}
	headers, ok := cfg["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected injected headers map, got %T", cfg["headers"])
	}
	want := `Bearer ${! qaynaq_connection_token("slack") }`
	if headers["Authorization"] != want {
		t.Errorf("expected %q, got %q", want, headers["Authorization"])
	}
}

func TestInjectConnectionAuthAcceptsBareName(t *testing.T) {
	cfg := buildProcessorMap(t, "http", "url: https://api.example.com\noauth_connection: slack\n")

	headers := cfg["headers"].(map[string]any)
	want := `Bearer ${! qaynaq_connection_token("slack") }`
	if headers["Authorization"] != want {
		t.Errorf("expected %q, got %q", want, headers["Authorization"])
	}
}

func TestInjectConnectionAuthMergesHeadersAndWins(t *testing.T) {
	cfg := buildProcessorMap(t, "http", `
url: https://api.example.com
oauth_connection: ${QAYNAQ_CONN_slack}
headers:
  Content-Type: application/json
  Authorization: Bearer manual
`)

	headers := cfg["headers"].(map[string]any)
	if headers["Content-Type"] != "application/json" {
		t.Errorf("existing headers should be preserved, got %v", headers)
	}
	want := `Bearer ${! qaynaq_connection_token("slack") }`
	if headers["Authorization"] != want {
		t.Errorf("connection should override a manual Authorization header, got %q", headers["Authorization"])
	}
}

func TestInjectConnectionAuthEmptyValueOnlyStrips(t *testing.T) {
	cfg := buildProcessorMap(t, "http", "url: https://api.example.com\noauth_connection: \"\"\n")

	if _, exists := cfg[oauthConnectionField]; exists {
		t.Error("empty oauth_connection should still be stripped")
	}
	if _, exists := cfg["headers"]; exists {
		t.Error("no headers should be injected for an empty connection")
	}
}

func TestInjectConnectionAuthIgnoresOtherComponents(t *testing.T) {
	cfg := buildProcessorMap(t, "google_calendar", "oauth_connection: ${QAYNAQ_CONN_g}\naction: list_events\n")

	if cfg[oauthConnectionField] != "${QAYNAQ_CONN_g}" {
		t.Errorf("components with a native oauth_connection field must keep it, got %v", cfg[oauthConnectionField])
	}
}

func TestValidateFlowHTTPComponentsWithConnection(t *testing.T) {
	flow := persistence.Flow{
		InputLabel:      "in",
		InputComponent:  "http_client",
		InputConfig:     []byte("url: https://api.example.com/items\noauth_connection: ${QAYNAQ_CONN_github}\n"),
		OutputLabel:     "out",
		OutputComponent: "http_client",
		OutputConfig:    []byte("url: https://api.example.com/sink\noauth_connection: ${QAYNAQ_CONN_github}\n"),
		Processors: []persistence.FlowProcessor{{
			Label:     "enrich",
			Component: "http",
			Config:    []byte("url: https://api.example.com/enrich\noauth_connection: ${QAYNAQ_CONN_github}\n"),
		}},
	}

	if err := ValidateFlow(flow); err != nil {
		t.Fatalf("flow with oauth_connection on HTTP components should lint cleanly, got: %v", err)
	}
}
