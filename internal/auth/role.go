package auth

import (
	"path/filepath"
	"strings"

	"github.com/jmespath/go-jmespath"
	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/rs/zerolog/log"
)

const (
	RoleAdmin = "admin"
	RoleMCP   = "mcp"
)

// EvaluateRole returns "admin", "mcp", or "". JMESPath wins when it returns a
// recognized role; otherwise the lists are consulted. When all three fields are
// empty, every user is admin (legacy single-tier behavior).
func EvaluateRole(cfg *config.AuthConfig, claims map[string]any, email string) string {
	if cfg == nil {
		return RoleAdmin
	}

	if cfg.OAuth2RoleAttributePath != "" {
		result, err := jmespath.Search(cfg.OAuth2RoleAttributePath, claims)
		if err != nil {
			log.Warn().Err(err).Str("expr", cfg.OAuth2RoleAttributePath).Msg("role_attribute_path evaluation failed")
		} else if s, ok := result.(string); ok {
			switch s {
			case "Admin":
				return RoleAdmin
			case "MCP":
				return RoleMCP
			}
		}
	}

	emailLower := strings.ToLower(email)

	for _, pattern := range cfg.OAuth2AdminUsers {
		if matchEmailPattern(pattern, emailLower) {
			return RoleAdmin
		}
	}
	for _, pattern := range cfg.OAuth2MCPUsers {
		if matchEmailPattern(pattern, emailLower) {
			return RoleMCP
		}
	}

	if cfg.OAuth2RoleAttributePath == "" && len(cfg.OAuth2AdminUsers) == 0 && len(cfg.OAuth2MCPUsers) == 0 {
		return RoleAdmin
	}

	return ""
}

func matchEmailPattern(pattern, emailLower string) bool {
	p := strings.ToLower(strings.TrimSpace(pattern))
	if p == "" {
		return false
	}
	if !strings.ContainsAny(p, "*?[") {
		return p == emailLower
	}
	matched, err := filepath.Match(p, emailLower)
	if err != nil {
		return false
	}
	return matched
}
