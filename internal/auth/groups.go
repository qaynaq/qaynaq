package auth

import (
	"strings"

	"github.com/jmespath/go-jmespath"
	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/rs/zerolog/log"
)

// EvaluateGroups extracts the caller's IdP group names from OIDC claims using
// the configured JMESPath. Accepts a string array or a single string result.
func EvaluateGroups(cfg *config.AuthConfig, claims map[string]any) []string {
	if cfg == nil || cfg.OAuth2GroupsAttributePath == "" || claims == nil {
		return nil
	}

	result, err := jmespath.Search(cfg.OAuth2GroupsAttributePath, claims)
	if err != nil {
		log.Warn().Err(err).Str("expr", cfg.OAuth2GroupsAttributePath).Msg("groups_attribute_path evaluation failed")
		return nil
	}

	var raw []any
	switch v := result.(type) {
	case []any:
		raw = v
	case string:
		raw = []any{v}
	default:
		return nil
	}

	seen := make(map[string]bool, len(raw))
	groups := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		groups = append(groups, s)
	}
	return groups
}
