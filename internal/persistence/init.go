package persistence

import (
	"database/sql"

	configstruct "github.com/qaynaq/qaynaq/internal/config"
	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/qaynaq/qaynaq/internal/persistence/migrations"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewGormDB(config *configstruct.DatabaseConfig) *gorm.DB {
	var db *gorm.DB
	var err error

	customLogger := logger.NewGormLogger(log.Logger)

	gormConfig := &gorm.Config{
		TranslateError: true,
		Logger:         customLogger,
	}

	var dialect string

	switch config.Driver {
	case configstruct.DatabaseTypeSQLite:
		sqlDB, sqlErr := sql.Open("sqlite", config.URI)
		if sqlErr != nil {
			log.Fatal().Err(sqlErr).Msg("Failed to open SQLite connection")
			return nil
		}
		sqlDB.SetMaxOpenConns(1)
		db, err = gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), gormConfig)
		if err == nil {
			db.Exec("PRAGMA journal_mode=WAL")
			db.Exec("PRAGMA busy_timeout=5000")
		}
		dialect = "sqlite"
	case configstruct.DatabaseTypePostgres:
		db, err = gorm.Open(postgres.Open(config.URI), gormConfig)
		dialect = "postgres"
	default:
		log.Fatal().Msg("Unsupported database driver")
		return nil
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get underlying sql.DB")
		return nil
	}

	if err := migrations.Run(sqlDB, dialect); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	return db
}
