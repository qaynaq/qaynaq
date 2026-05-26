package persistence

import (
	"database/sql"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func setupConnectionTestDB(t *testing.T) (*gorm.DB, ConnectionRepository) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	if err := db.AutoMigrate(&Connection{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db, NewConnectionRepository(db)
}

func seedConnection(t *testing.T, repo ConnectionRepository, name string) {
	t.Helper()
	conn := &Connection{
		Name:            name,
		Provider:        "google",
		EncryptedConfig: "x",
		EncryptedToken:  "y",
	}
	if _, err := repo.Create(conn); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func TestRecordFailurePreservesFirstFailedAt(t *testing.T) {
	_, repo := setupConnectionTestDB(t)
	seedConnection(t, repo, "c1")

	if err := repo.RecordFailure("c1", "boom"); err != nil {
		t.Fatalf("first RecordFailure: %v", err)
	}
	first, err := repo.GetByName("c1")
	if err != nil {
		t.Fatalf("get after first failure: %v", err)
	}
	if first.FirstFailedAt == nil {
		t.Fatal("first_failed_at should be set on first failure")
	}
	if first.ConsecutiveFailures != 1 {
		t.Errorf("consecutive_failures = %d, want 1", first.ConsecutiveFailures)
	}
	firstFailedAt := *first.FirstFailedAt

	// sqlite's CURRENT_TIMESTAMP is second-resolution; pause so a second
	// failure can't end up at the exact same instant.
	time.Sleep(1100 * time.Millisecond)

	if err := repo.RecordFailure("c1", "boom again"); err != nil {
		t.Fatalf("second RecordFailure: %v", err)
	}
	second, err := repo.GetByName("c1")
	if err != nil {
		t.Fatalf("get after second failure: %v", err)
	}
	if second.FirstFailedAt == nil {
		t.Fatal("first_failed_at should still be set after second failure")
	}
	if !second.FirstFailedAt.Equal(firstFailedAt) {
		t.Errorf("first_failed_at moved: was %v, now %v", firstFailedAt, *second.FirstFailedAt)
	}
	if second.ConsecutiveFailures != 2 {
		t.Errorf("consecutive_failures = %d, want 2", second.ConsecutiveFailures)
	}
	if second.LastError != "boom again" {
		t.Errorf("last_error = %q, want %q", second.LastError, "boom again")
	}
	if second.LastErrorAt == nil || !second.LastErrorAt.After(firstFailedAt) {
		t.Errorf("last_error_at = %v, expected later than first_failed_at %v", second.LastErrorAt, firstFailedAt)
	}
}

func TestClearErrorResetsAllFields(t *testing.T) {
	_, repo := setupConnectionTestDB(t)
	seedConnection(t, repo, "c1")

	if err := repo.RecordFailure("c1", "boom"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	if err := repo.RecordFailure("c1", "boom"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	if err := repo.ClearError("c1"); err != nil {
		t.Fatalf("ClearError: %v", err)
	}

	got, err := repo.GetByName("c1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.LastError != "" {
		t.Errorf("last_error = %q, want empty", got.LastError)
	}
	if got.LastErrorAt != nil {
		t.Errorf("last_error_at = %v, want nil", got.LastErrorAt)
	}
	if got.FirstFailedAt != nil {
		t.Errorf("first_failed_at = %v, want nil", got.FirstFailedAt)
	}
	if got.ConsecutiveFailures != 0 {
		t.Errorf("consecutive_failures = %d, want 0", got.ConsecutiveFailures)
	}
}

func TestListFailingOnlyReturnsBrokenConnections(t *testing.T) {
	_, repo := setupConnectionTestDB(t)
	seedConnection(t, repo, "healthy")
	seedConnection(t, repo, "broken")

	if err := repo.RecordFailure("broken", "nope"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	failing, err := repo.ListFailing()
	if err != nil {
		t.Fatalf("ListFailing: %v", err)
	}
	if len(failing) != 1 {
		t.Fatalf("ListFailing returned %d, want 1", len(failing))
	}
	if failing[0].Name != "broken" {
		t.Errorf("ListFailing[0].Name = %q, want %q", failing[0].Name, "broken")
	}
}
