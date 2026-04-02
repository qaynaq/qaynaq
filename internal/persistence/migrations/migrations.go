package migrations

import (
	"crypto/md5"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed sqlite/*.sql
var sqliteFS embed.FS

//go:embed postgres/*.sql
var postgresFS embed.FS

func Run(db *sql.DB, dialect string) error {
	var (
		fsys        fs.FS
		err         error
		placeholder [2]string
	)

	switch dialect {
	case "sqlite":
		fsys, err = fs.Sub(sqliteFS, "sqlite")
		placeholder = [2]string{"?", "?"}
	case "postgres":
		fsys, err = fs.Sub(postgresFS, "postgres")
		placeholder = [2]string{"$1", "$2"}
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}
	if err != nil {
		return fmt.Errorf("failed to get migration filesystem: %w", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version text PRIMARY KEY,
		checksum text NOT NULL,
		applied_at timestamp DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	selectQuery := fmt.Sprintf("SELECT checksum FROM schema_migrations WHERE version = %s", placeholder[0])
	insertQuery := fmt.Sprintf("INSERT INTO schema_migrations (version, checksum) VALUES (%s, %s)", placeholder[0], placeholder[1])

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		checksum := fmt.Sprintf("%x", md5.Sum(content))

		var storedChecksum string
		err = db.QueryRow(selectQuery, version).Scan(&storedChecksum)
		if err == nil {
			if storedChecksum != checksum {
				return fmt.Errorf("migration %s has been modified after being applied (expected checksum %s, got %s)", version, storedChecksum, checksum)
			}
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check migration %s: %w", version, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", name, err)
		}

		if _, err := db.Exec(insertQuery, version, checksum); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}
	}

	return nil
}
