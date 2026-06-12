package templates

import (
	"strings"
	"testing"
)

const validManifest = `
id: demo_pack
name: Demo Pack
description: A demo pack
version: 0.1.0
variables:
  - key: shop_name
    title: Store Name
    type: string
    required: true
  - key: api_token
    title: API Token
    type: secret
flows:
  - name: demo_list_orders
    kind: tool
    description: List orders
    input:
      component: mcp_tool
      config:
        name: demo_list_orders
        description: List orders
    processors:
      - label: shopify
        component: shopify
        config:
          shop_name: "{{ shop_name }}"
          api_access_token: "{{ api_token }}"
          action: list_orders
      - label: error_handler
        component: catch
        config:
          - mapping: root.error = error()
    output:
      component: sync_response
`

func TestParseManifestValid(t *testing.T) {
	m, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != "demo_pack" {
		t.Errorf("id = %q, want demo_pack", m.ID)
	}
	if len(m.Variables) != 2 {
		t.Errorf("variables = %d, want 2", len(m.Variables))
	}
	if len(m.Flows) != 1 {
		t.Errorf("flows = %d, want 1", len(m.Flows))
	}
	if m.Flows[0].Kind != FlowKindTool {
		t.Errorf("kind = %q, want tool", m.Flows[0].Kind)
	}
}

func TestParseManifestErrors(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{
			name:    "unknown field",
			mutate:  func(s string) string { return strings.Replace(s, "version: 0.1.0", "version: 0.1.0\nbogus: true", 1) },
			wantErr: "field bogus not found",
		},
		{
			name:    "file reference rejected",
			mutate:  func(s string) string { return strings.Replace(s, "kind: tool", "kind: tool\n    file: flows/x.yaml", 1) },
			wantErr: "not supported yet",
		},
		{
			name:    "unknown placeholder",
			mutate:  func(s string) string { return strings.Replace(s, "{{ shop_name }}", "{{ missing_var }}", 1) },
			wantErr: "placeholder {{ missing_var }}",
		},
		{
			name:    "bad variable type",
			mutate:  func(s string) string { return strings.Replace(s, "type: secret", "type: password", 1) },
			wantErr: "type must be one of",
		},
		{
			name:    "bad flow kind",
			mutate:  func(s string) string { return strings.Replace(s, "kind: tool", "kind: trigger", 1) },
			wantErr: "kind must be one of",
		},
		{
			name:    "missing version",
			mutate:  func(s string) string { return strings.Replace(s, "version: 0.1.0", "", 1) },
			wantErr: "version is required",
		},
		{
			name:    "missing input component",
			mutate:  func(s string) string { return strings.Replace(s, "      component: mcp_tool\n", "", 1) },
			wantErr: "input.component is required",
		},
		{
			name: "duplicate variable key",
			mutate: func(s string) string {
				return strings.Replace(s, "  - key: api_token", "  - key: shop_name\n    title: Dup\n    type: string\n  - key: api_token", 1)
			},
			wantErr: "duplicate variable key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(tc.mutate(validManifest)))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}
