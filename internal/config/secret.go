package config

const (
	SecretProviderLocal = "local"
)

type SecretConfig struct {
	Provider string
	Key      string
}
