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
	OAuth2AuthorizationURL  string
	OAuth2TokenURL          string
	OAuth2RedirectURL       string
	OAuth2Scopes            []string
	OAuth2UserInfoURL       string
	OAuth2AllowedUsers      []string
	OAuth2AllowedDomains    []string
	OAuth2SessionCookieName string
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
		if c.OAuth2AuthorizationURL == "" {
			return fmt.Errorf("oauth2 requires auth.oauth2-authorization-url")
		}
		if c.OAuth2TokenURL == "" {
			return fmt.Errorf("oauth2 requires auth.oauth2-token-url")
		}
		if c.OAuth2RedirectURL == "" {
			return fmt.Errorf("oauth2 requires auth.oauth2-redirect-url")
		}
		return nil
	default:
		return fmt.Errorf("invalid auth type: %s (must be none, basic, or oauth2)", c.Type)
	}
}
