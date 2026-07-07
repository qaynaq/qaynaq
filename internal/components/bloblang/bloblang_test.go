package bloblang

import (
	"strings"
	"testing"
	"time"

	"github.com/warpstreamlabs/bento/public/bloblang"

	"github.com/qaynaq/qaynaq/internal/connauth"
	"github.com/qaynaq/qaynaq/internal/vault"
)

type stubVault struct {
	token string
}

func (s *stubVault) GetSecret(string) (string, error)          { return "", nil }
func (s *stubVault) GetConnectionToken(string) (string, error) { return "", nil }
func (s *stubVault) GetAccessToken(string) (vault.AccessToken, error) {
	return vault.AccessToken{AccessToken: s.token, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (s *stubVault) ForceRefreshAccessToken(string) (vault.AccessToken, error) {
	return vault.AccessToken{AccessToken: s.token, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func TestConnectionTokenFunction(t *testing.T) {
	exec, err := bloblang.Parse(`root = qaynaq_connection_token("conn1")`)
	if err != nil {
		t.Fatalf("failed to parse mapping: %v", err)
	}

	connauth.SetVaultProvider(&stubVault{token: "tok-abc"})
	defer connauth.SetVaultProvider(nil)

	res, err := exec.Query(nil)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if res != "tok-abc" {
		t.Errorf("expected tok-abc, got %v", res)
	}
}

func TestConnectionTokenFunctionWithoutProvider(t *testing.T) {
	exec, err := bloblang.Parse(`root = qaynaq_connection_token("conn1")`)
	if err != nil {
		t.Fatalf("failed to parse mapping: %v", err)
	}

	connauth.SetVaultProvider(nil)

	_, err = exec.Query(nil)
	if err == nil || !strings.Contains(err.Error(), "not available") {
		t.Errorf("expected not-available error, got %v", err)
	}
}
