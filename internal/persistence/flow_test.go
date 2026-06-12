package persistence

import (
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func setupFlowTestDB(t *testing.T) (*gorm.DB, FlowRepository) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	if err := db.AutoMigrate(&Flow{}, &FlowProcessor{}, &FlowCache{}, &Buffer{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db, NewFlowRepository(db)
}

func seedFlow(t *testing.T, repo FlowRepository) *Flow {
	t.Helper()
	flow := &Flow{
		Name:            "f1",
		InputComponent:  "http_server",
		InputConfig:     []byte("{}"),
		OutputComponent: "stdout",
		OutputConfig:    []byte("{}"),
		Status:          FlowStatusActive,
	}
	if err := repo.Create(flow); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return flow
}

func TestRecordFailureSetsErrorAndStatus(t *testing.T) {
	_, repo := setupFlowTestDB(t)
	flow := seedFlow(t, repo)

	if err := repo.RecordFailure(flow.ID, "pipeline crashed"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	got, err := repo.FindByID(flow.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != FlowStatusFailed {
		t.Errorf("status = %q, want %q", got.Status, FlowStatusFailed)
	}
	if got.LastError != "pipeline crashed" {
		t.Errorf("last_error = %q, want %q", got.LastError, "pipeline crashed")
	}
	if got.LastErrorAt == nil {
		t.Error("last_error_at = nil, want set")
	}
}

func TestUpdateStatusClearsErrorWhenLeavingFailed(t *testing.T) {
	_, repo := setupFlowTestDB(t)
	flow := seedFlow(t, repo)

	if err := repo.RecordFailure(flow.ID, "pipeline crashed"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	if err := repo.UpdateStatus(flow.ID, FlowStatusActive); err != nil {
		t.Fatalf("UpdateStatus active: %v", err)
	}

	got, err := repo.FindByID(flow.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != FlowStatusActive {
		t.Errorf("status = %q, want %q", got.Status, FlowStatusActive)
	}
	if got.LastError != "" {
		t.Errorf("last_error = %q, want empty after recovery", got.LastError)
	}
	if got.LastErrorAt != nil {
		t.Errorf("last_error_at = %v, want nil after recovery", got.LastErrorAt)
	}
}

func TestUpdateStatusToFailedPreservesError(t *testing.T) {
	_, repo := setupFlowTestDB(t)
	flow := seedFlow(t, repo)

	if err := repo.RecordFailure(flow.ID, "pipeline crashed"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	// A redundant UpdateStatus(failed) (e.g. duplicate worker report) must not
	// wipe the reason we already captured via RecordFailure.
	if err := repo.UpdateStatus(flow.ID, FlowStatusFailed); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	got, err := repo.FindByID(flow.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.LastError != "pipeline crashed" {
		t.Errorf("last_error = %q, want %q (preserved)", got.LastError, "pipeline crashed")
	}
	if got.LastErrorAt == nil {
		t.Error("last_error_at = nil, want preserved")
	}
}

func TestListAllByManagedBy(t *testing.T) {
	_, repo := setupFlowTestDB(t)

	managed := "shopify"
	for _, name := range []string{"tool_a", "tool_b"} {
		flow := &Flow{
			Name:            name,
			InputComponent:  "mcp_tool",
			InputConfig:     []byte("{}"),
			OutputComponent: "sync_response",
			OutputConfig:    []byte("{}"),
			Status:          FlowStatusActive,
			IsCurrent:       true,
			ManagedBy:       &managed,
		}
		if err := repo.Create(flow); err != nil {
			t.Fatalf("seed managed flow: %v", err)
		}
	}
	if _, err := repo.FindByID(seedFlow(t, repo).ID); err != nil {
		t.Fatalf("seed unmanaged flow: %v", err)
	}

	flows, err := repo.ListAllByManagedBy(managed)
	if err != nil {
		t.Fatalf("ListAllByManagedBy: %v", err)
	}
	if len(flows) != 2 {
		t.Errorf("flows = %d, want 2", len(flows))
	}

	flows, err = repo.ListAllByManagedBy("missing_pack")
	if err != nil {
		t.Fatalf("ListAllByManagedBy empty: %v", err)
	}
	if len(flows) != 0 {
		t.Errorf("flows = %d, want 0 for unknown pack", len(flows))
	}
}
