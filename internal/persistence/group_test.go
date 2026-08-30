package persistence

import (
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func setupGroupTestDB(t *testing.T) (GroupRepository, UserGroupsRepository) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	if err := db.AutoMigrate(&Group{}, &UserGroups{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewGroupRepository(db), NewUserGroupsRepository(db)
}

func TestGroupObservedWinsOverManual(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	imported, err := repo.ImportManual([]string{"accounting", "legal"})
	if err != nil {
		t.Fatalf("ImportManual: %v", err)
	}
	if imported != 2 {
		t.Fatalf("imported: got %d, want 2", imported)
	}

	// Re-import of an existing name is a no-op.
	imported, err = repo.ImportManual([]string{"accounting"})
	if err != nil {
		t.Fatalf("ImportManual: %v", err)
	}
	if imported != 0 {
		t.Fatalf("re-import: got %d, want 0", imported)
	}

	// A login observation upgrades the manual entry and stamps last_seen_at.
	if err := repo.UpsertObserved([]string{"accounting", "engineering"}); err != nil {
		t.Fatalf("UpsertObserved: %v", err)
	}

	groups, err := repo.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	bySource := map[string]string{}
	seen := map[string]bool{}
	for _, g := range groups {
		bySource[g.Name] = g.Source
		seen[g.Name] = g.LastSeenAt != nil
	}
	if bySource["accounting"] != GroupSourceObserved || !seen["accounting"] {
		t.Errorf("accounting should be observed with last_seen_at set")
	}
	if bySource["legal"] != GroupSourceManual || seen["legal"] {
		t.Errorf("legal should remain manual without last_seen_at")
	}
	if bySource["engineering"] != GroupSourceObserved {
		t.Errorf("engineering should be observed")
	}

	if err := repo.Delete("legal"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	groups, _ = repo.List()
	if len(groups) != 2 {
		t.Errorf("after delete: got %d groups, want 2", len(groups))
	}
}

func TestUserGroupsSnapshot(t *testing.T) {
	_, repo := setupGroupTestDB(t)

	if groups, err := repo.Get("a@b.c"); err != nil || groups != nil {
		t.Fatalf("unknown user: got %v, %v", groups, err)
	}

	if err := repo.Set("a@b.c", []string{"accounting"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	groups, err := repo.Get("a@b.c")
	if err != nil || len(groups) != 1 || groups[0] != "accounting" {
		t.Fatalf("got %v, %v", groups, err)
	}

	// Login with changed membership replaces the snapshot.
	if err := repo.Set("a@b.c", []string{"legal", "engineering"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	groups, _ = repo.Get("a@b.c")
	if len(groups) != 2 || groups[0] != "legal" {
		t.Fatalf("snapshot not replaced: %v", groups)
	}

	// Leaving all groups yields an empty (not stale) snapshot.
	if err := repo.Set("a@b.c", nil); err != nil {
		t.Fatalf("Set nil: %v", err)
	}
	groups, _ = repo.Get("a@b.c")
	if len(groups) != 0 {
		t.Fatalf("expected empty snapshot, got %v", groups)
	}
}
