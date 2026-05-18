package auth

import (
	"testing"

	"github.com/qaynaq/qaynaq/internal/config"
)

func TestEvaluateRole(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *config.AuthConfig
		claims map[string]any
		email  string
		want   string
	}{
		{
			name:  "legacy default (no config) → admin",
			cfg:   &config.AuthConfig{},
			email: "alice@x.com",
			want:  RoleAdmin,
		},
		{
			name: "jmespath returns Admin",
			cfg: &config.AuthConfig{
				OAuth2RoleAttributePath: "contains(groups[*], 'qaynaq-admins') && 'Admin' || 'None'",
			},
			claims: map[string]any{"groups": []any{"qaynaq-admins"}},
			email:  "alice@x.com",
			want:   RoleAdmin,
		},
		{
			name: "jmespath returns MCP",
			cfg: &config.AuthConfig{
				OAuth2RoleAttributePath: "contains(groups[*], 'qaynaq-mcp') && 'MCP' || 'None'",
			},
			claims: map[string]any{"groups": []any{"qaynaq-mcp"}},
			email:  "bob@x.com",
			want:   RoleMCP,
		},
		{
			name: "jmespath returns unrecognized → falls through to lists",
			cfg: &config.AuthConfig{
				OAuth2RoleAttributePath: "groups[0]",
				OAuth2AdminUsers:        []string{"alice@x.com"},
			},
			claims: map[string]any{"groups": []any{"some-other-group"}},
			email:  "alice@x.com",
			want:   RoleAdmin,
		},
		{
			name: "list match - admin exact",
			cfg: &config.AuthConfig{
				OAuth2AdminUsers: []string{"alice@x.com"},
				OAuth2MCPUsers:   []string{"bob@x.com"},
			},
			email: "alice@x.com",
			want:  RoleAdmin,
		},
		{
			name: "list match - case insensitive",
			cfg: &config.AuthConfig{
				OAuth2AdminUsers: []string{"Alice@X.com"},
			},
			email: "alice@x.com",
			want:  RoleAdmin,
		},
		{
			name: "list match - domain glob",
			cfg: &config.AuthConfig{
				OAuth2AdminUsers: []string{"*@admin.x.com"},
				OAuth2MCPUsers:   []string{"*@x.com"},
			},
			email: "bob@x.com",
			want:  RoleMCP,
		},
		{
			name: "list match - prefix glob",
			cfg: &config.AuthConfig{
				OAuth2MCPUsers: []string{"sales-*@x.com"},
			},
			email: "sales-1@x.com",
			want:  RoleMCP,
		},
		{
			name: "list match - admin wins over mcp when both contain the email",
			cfg: &config.AuthConfig{
				OAuth2AdminUsers: []string{"alice@x.com"},
				OAuth2MCPUsers:   []string{"*@x.com"},
			},
			email: "alice@x.com",
			want:  RoleAdmin,
		},
		{
			name: "no match with config set → empty",
			cfg: &config.AuthConfig{
				OAuth2AdminUsers: []string{"alice@x.com"},
			},
			email: "stranger@y.com",
			want:  "",
		},
		{
			name: "nil cfg → admin",
			cfg:  nil,
			want: RoleAdmin,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateRole(tc.cfg, tc.claims, tc.email)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
