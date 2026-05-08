package mcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/vault"
)

// stubSecretRepo is a tiny in-memory secret repository for tests.
type stubSecretRepo struct {
	secrets map[string]string
}

func (s *stubSecretRepo) List() ([]persistence.Secret, error) { return nil, nil }
func (s *stubSecretRepo) GetByKey(key string) (*persistence.Secret, error) {
	v, ok := s.secrets[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return &persistence.Secret{Key: key, EncryptedValue: v}, nil
}
func (s *stubSecretRepo) Create(_ *persistence.Secret) (bool, error) { return false, nil }
func (s *stubSecretRepo) Delete(_ string) error                      { return nil }

func newTestVault(t *testing.T) *vault.AESGCM {
	t.Helper()
	v, err := vault.NewAESGCM([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("vault: %v", err)
	}
	return v
}

func TestEnvResolver_Literal(t *testing.T) {
	v := newTestVault(t)
	r := NewEnvResolver(&stubSecretRepo{}, v)
	out, err := r.Resolve(map[string]string{"DEBUG": "1", "URL": "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("want 2 entries, got %d", len(out))
	}
	got := strings.Join(out, ",")
	if !strings.Contains(got, "DEBUG=1") || !strings.Contains(got, "URL=https://example.com") {
		t.Fatalf("missing literals: %v", out)
	}
}

func TestEnvResolver_SecretRef(t *testing.T) {
	v := newTestVault(t)
	enc, _ := v.Encrypt("xoxb-real-token")
	repo := &stubSecretRepo{secrets: map[string]string{"SLACK_BOT": enc}}
	r := NewEnvResolver(repo, v)

	out, err := r.Resolve(map[string]string{"SLACK_BOT_TOKEN": "${SLACK_BOT}"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(out) != 1 || out[0] != "SLACK_BOT_TOKEN=xoxb-real-token" {
		t.Fatalf("expected resolved secret, got %v", out)
	}
}

func TestEnvResolver_MissingSecret(t *testing.T) {
	v := newTestVault(t)
	r := NewEnvResolver(&stubSecretRepo{}, v)
	_, err := r.Resolve(map[string]string{"X": "${NOPE}"})
	if !errors.Is(err, ErrMissingSecret) {
		t.Fatalf("expected ErrMissingSecret, got %v", err)
	}
}

func TestEnvResolver_Escape(t *testing.T) {
	v := newTestVault(t)
	r := NewEnvResolver(&stubSecretRepo{}, v)
	out, err := r.Resolve(map[string]string{"X": "$${NOT_A_REF}"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if out[0] != "X=${NOT_A_REF}" {
		t.Fatalf("expected escape to literal, got %s", out[0])
	}
}

func TestEnvResolver_PartialMatchSubstitutes(t *testing.T) {
	v := newTestVault(t)
	enc, _ := v.Encrypt("xoxb-real")
	r := NewEnvResolver(&stubSecretRepo{secrets: map[string]string{"X": enc}}, v)
	out, err := r.Resolve(map[string]string{"Y": "prefix-${X}-suffix"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if out[0] != "Y=prefix-xoxb-real-suffix" {
		t.Fatalf("expected inline substitution, got %s", out[0])
	}
}

func TestEnvResolver_MultipleRefsInOneValue(t *testing.T) {
	v := newTestVault(t)
	encUser, _ := v.Encrypt("alice")
	encPass, _ := v.Encrypt("s3cret")
	r := NewEnvResolver(&stubSecretRepo{secrets: map[string]string{
		"DB_USER": encUser,
		"DB_PASS": encPass,
	}}, v)
	out, err := r.Resolve(map[string]string{
		"DSN": "postgres://${DB_USER}:${DB_PASS}@db.example.com/qaynaq",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if out[0] != "DSN=postgres://alice:s3cret@db.example.com/qaynaq" {
		t.Fatalf("expected both refs resolved, got %s", out[0])
	}
}

func TestEnvResolver_MissingSecretInComposition(t *testing.T) {
	v := newTestVault(t)
	r := NewEnvResolver(&stubSecretRepo{}, v)
	_, err := r.Resolve(map[string]string{"Y": "prefix-${NOPE}-suffix"})
	if !errors.Is(err, ErrMissingSecret) {
		t.Fatalf("expected ErrMissingSecret in composed string, got %v", err)
	}
}

func TestSubstituteArgs_HappyPath(t *testing.T) {
	out, err := SubstituteArgs([]string{"-y", "@x/server-fs", "${ALLOWED_DIR}"}, map[string]string{"ALLOWED_DIR": "/tmp/sandbox"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if out[2] != "/tmp/sandbox" {
		t.Fatalf("substitution failed: %v", out)
	}
}

func TestSubstituteArgs_MissingPlaceholder(t *testing.T) {
	_, err := SubstituteArgs([]string{"--repo", "${REPO_PATH}"}, map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "REPO_PATH") {
		t.Fatalf("expected REPO_PATH error, got %v", err)
	}
}

func TestSubstituteArgs_Escape(t *testing.T) {
	out, err := SubstituteArgs([]string{"$${LITERAL}"}, map[string]string{"LITERAL": "should-not-be-used"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if out[0] != "${LITERAL}" {
		t.Fatalf("expected escape, got %s", out[0])
	}
}
