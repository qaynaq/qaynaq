package migrations

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSqliteMigrationsApplyCleanly(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := Run(db, "sqlite"); err != nil {
		t.Fatalf("migrations failed on a fresh database: %v", err)
	}

	// Re-running must be a no-op, not an error.
	if err := Run(db, "sqlite"); err != nil {
		t.Fatalf("migrations are not idempotent: %v", err)
	}

	for _, table := range []string{"groups", "user_groups", "mcp_call_logs"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('flows') WHERE name='allowed_groups'").Scan(&count); err != nil || count != 1 {
		t.Errorf("flows.allowed_groups column missing (count=%d, err=%v)", count, err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('mcp_servers') WHERE name='allowed_groups'").Scan(&count); err != nil || count != 1 {
		t.Errorf("mcp_servers.allowed_groups column missing (count=%d, err=%v)", count, err)
	}
}
