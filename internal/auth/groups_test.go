package auth

import (
	"reflect"
	"testing"

	"github.com/qaynaq/qaynaq/internal/config"
)

func TestEvaluateGroups(t *testing.T) {
	cfg := &config.AuthConfig{OAuth2GroupsAttributePath: "groups"}

	tests := []struct {
		name   string
		cfg    *config.AuthConfig
		claims map[string]any
		want   []string
	}{
		{
			name:   "string array",
			cfg:    cfg,
			claims: map[string]any{"groups": []any{"accounting", "legal"}},
			want:   []string{"accounting", "legal"},
		},
		{
			name:   "single string",
			cfg:    cfg,
			claims: map[string]any{"groups": "accounting"},
			want:   []string{"accounting"},
		},
		{
			name:   "dedupe and trim",
			cfg:    cfg,
			claims: map[string]any{"groups": []any{" accounting ", "accounting", ""}},
			want:   []string{"accounting"},
		},
		{
			name:   "non-string entries skipped",
			cfg:    cfg,
			claims: map[string]any{"groups": []any{"accounting", 42, true}},
			want:   []string{"accounting"},
		},
		{
			name:   "missing claim",
			cfg:    cfg,
			claims: map[string]any{"email": "a@b.c"},
			want:   nil,
		},
		{
			name:   "no path configured",
			cfg:    &config.AuthConfig{},
			claims: map[string]any{"groups": []any{"accounting"}},
			want:   nil,
		},
		{
			name:   "nested path",
			cfg:    &config.AuthConfig{OAuth2GroupsAttributePath: "realm_access.roles"},
			claims: map[string]any{"realm_access": map[string]any{"roles": []any{"engineering"}}},
			want:   []string{"engineering"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateGroups(tt.cfg, tt.claims)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
