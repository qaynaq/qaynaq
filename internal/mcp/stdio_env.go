package mcp

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/vault"
)

// ErrMissingSecret wraps the cause when an env value references ${NAME} but
// the secret is absent or undecryptable. The supervisor reads this sentinel
// to decide "stay failed, do not respawn."
var ErrMissingSecret = errors.New("missing secret")

// refPattern matches ${NAME} anywhere in a string. Used by both the env
// resolver (substitutes from secrets) and SubstituteArgs (substitutes from
// the env map). $${NAME} is the literal-escape form, handled before this
// pattern runs.
var refPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

type EnvResolver struct {
	secretRepo persistence.SecretRepository
	aesgcm     *vault.AESGCM
}

func NewEnvResolver(secretRepo persistence.SecretRepository, aesgcm *vault.AESGCM) *EnvResolver {
	return &EnvResolver{secretRepo: secretRepo, aesgcm: aesgcm}
}

func (r *EnvResolver) AESGCM() *vault.AESGCM {
	return r.aesgcm
}

// Resolve returns env entries in `KEY=value` form, ready for cmd.Env. Values
// matching ${NAME} are looked up in the secrets table and decrypted.
func (r *EnvResolver) Resolve(raw map[string]string) ([]string, error) {
	out := make([]string, 0, len(raw))
	for k, v := range raw {
		resolved, err := r.resolveOne(v)
		if err != nil {
			return nil, fmt.Errorf("env %s: %w", k, err)
		}
		out = append(out, k+"="+resolved)
	}
	return out, nil
}

// SubstituteArgs replaces ${NAME} placeholders in args with values from env.
// Unresolved placeholders are an error: the catalog asks for them, the env
// map must provide them. $${NAME} escapes to literal ${NAME}.
func SubstituteArgs(args []string, env map[string]string) ([]string, error) {
	out := make([]string, 0, len(args))
	for _, a := range args {
		a = strings.ReplaceAll(a, "$${", "\x00{")
		var unresolved []string
		replaced := refPattern.ReplaceAllStringFunc(a, func(match string) string {
			name := match[2 : len(match)-1]
			if val, ok := env[name]; ok {
				return val
			}
			unresolved = append(unresolved, name)
			return match
		})
		if len(unresolved) > 0 {
			return nil, fmt.Errorf("unresolved arg placeholder(s): %s", strings.Join(unresolved, ", "))
		}
		replaced = strings.ReplaceAll(replaced, "\x00{", "${")
		out = append(out, replaced)
	}
	return out, nil
}

// resolveOne expands every ${NAME} reference inside v by looking NAME up in
// the secrets table and substituting the decrypted value. Plain values pass
// through. Mixed strings like "prefix-${TOKEN}" are supported (matches the
// inline-interpolation behavior used elsewhere in Bento pipelines). Use
// $${NAME} to pass a literal ${NAME} through to the child.
func (r *EnvResolver) resolveOne(v string) (string, error) {
	if !strings.Contains(v, "${") && !strings.Contains(v, "$${") {
		return v, nil
	}
	v = strings.ReplaceAll(v, "$${", "\x00{")

	var resolveErr error
	replaced := refPattern.ReplaceAllStringFunc(v, func(match string) string {
		if resolveErr != nil {
			return match
		}
		name := match[2 : len(match)-1]
		secret, err := r.secretRepo.GetByKey(name)
		if err != nil {
			resolveErr = fmt.Errorf("%w: %s", ErrMissingSecret, name)
			return match
		}
		if r.aesgcm == nil {
			resolveErr = fmt.Errorf("vault not configured for secret %s", name)
			return match
		}
		plain, err := r.aesgcm.Decrypt(secret.EncryptedValue)
		if err != nil {
			resolveErr = fmt.Errorf("decrypt secret %s: %w", name, err)
			return match
		}
		return plain
	})
	if resolveErr != nil {
		return "", resolveErr
	}
	return strings.ReplaceAll(replaced, "\x00{", "${"), nil
}
