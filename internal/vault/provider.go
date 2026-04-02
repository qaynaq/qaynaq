package vault

type VaultProvider interface {
	GetSecret(key string) (string, error)
	GetConnectionToken(name string) (string, error)
}
