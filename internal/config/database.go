package config

const (
	DatabaseTypeSQLite   = "sqlite"
	DatabaseTypePostgres = "postgres"
)

type DatabaseConfig struct {
	Driver string
	URI    string
}
