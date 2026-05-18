package config

import "fmt"

type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeOAuth2 AuthType = "oauth2"
)

type AuthConfig struct {
	Type AuthType

	BasicUsername string
	BasicPassword string

	OAuth2ClientID          string
	OAuth2ClientSecret      string
	OAuth2IssuerURL         string
	OAuth2Scopes            []string
	OAuth2AllowedUsers      []string
	OAuth2AllowedDomains    []string
	OAuth2SessionCookieName string

	OAuth2RoleAttributePath string
	OAuth2AdminUsers        []string
	OAuth2MCPUsers          []string
}

func (c *AuthConfig) Validate() error {
	switch c.Type {
	case AuthTypeNone:
		return nil
	case AuthTypeBasic:
		if c.BasicUsername == "" || c.BasicPassword == "" {
			return fmt.Errorf("basic auth requires auth.basic-username and auth.basic-password")
		}
		return nil
	case AuthTypeOAuth2:
		if c.OAuth2ClientID == "" {
			return fmt.Errorf("oauth2 requires auth.oauth2-client-id")
		}
		if c.OAuth2ClientSecret == "" {
			return fmt.Errorf("oauth2 requires auth.oauth2-client-secret")
		}
		if c.OAuth2IssuerURL == "" {
			return fmt.Errorf("oauth2 requires auth.oauth2-issuer-url (the OIDC issuer; we discover the rest via .well-known/openid-configuration)")
		}
		return nil
	default:
		return fmt.Errorf("invalid auth type: %s (must be none, basic, or oauth2)", c.Type)
	}
}
