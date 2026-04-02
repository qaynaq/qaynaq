package ratelimiter

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"database/sql"

	"github.com/qaynaq/qaynaq/internal/persistence"
	_ "modernc.org/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&persistence.RateLimit{}, &persistence.RateLimitState{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func createTestRateLimit(t *testing.T, db *gorm.DB, label string, count int64, interval string, burst int64) {
	repo := persistence.NewRateLimitRepository(db)

	config := map[string]any{
		"coordinator_address": "localhost:50000",
		"count":               count,
		"interval":            interval,
		"burst":               burst,
	}
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	rateLimit := &persistence.RateLimit{
		Label:     label,
		Component: "coordinator",
		Config:    configYaml,
		IsCurrent: true,
		CreatedAt: time.Now(),
	}
	err = repo.Create(rateLimit)
	if err != nil {
		t.Fatalf("failed to create rate limit: %v", err)
	}
}

func TestEngine_Check_BasicRateLimit(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_limit", 10, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result, err := engine.Check("test_limit", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}

	if result.Remaining != 9 {
		t.Errorf("expected remaining to be 9, got %d", result.Remaining)
	}

	if result.Limit != 10 {
		t.Errorf("expected limit to be 10, got %d", result.Limit)
	}
}

func TestEngine_Check_WithBurst(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_burst", 10, "1s", 5)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result, err := engine.Check("test_burst", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}

	if result.Limit != 15 {
		t.Errorf("expected limit to be 15 (10 + 5 burst), got %d", result.Limit)
	}

	if result.Remaining != 14 {
		t.Errorf("expected remaining to be 14, got %d", result.Remaining)
	}
}

func TestEngine_Check_ExceedLimit(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_exceed", 2, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	for i := 0; i < 2; i++ {
		result, err := engine.Check("test_exceed", "user1", 1)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if !result.Allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	result, err := engine.Check("test_exceed", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Allowed {
		t.Errorf("expected request to be denied")
	}

	if result.RetryAfterMs <= 0 {
		t.Errorf("expected RetryAfterMs to be positive, got %d", result.RetryAfterMs)
	}
}

func TestEngine_Check_PerKeyIsolation(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_perkey", 2, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result1, err := engine.Check("test_perkey", "user1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result1.Allowed {
		t.Errorf("user1 request should be allowed")
	}

	result2, err := engine.Check("test_perkey", "user2", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result2.Allowed {
		t.Errorf("user2 request should be allowed (different key)")
	}

	result3, err := engine.Check("test_perkey", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result3.Allowed {
		t.Errorf("user1 additional request should be denied")
	}

	result4, err := engine.Check("test_perkey", "user2", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result4.Allowed {
		t.Errorf("user2 additional request should be denied")
	}
}

func TestEngine_Check_WithCost(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_cost", 10, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result, err := engine.Check("test_cost", "user1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}

	if result.Remaining != 5 {
		t.Errorf("expected remaining to be 5, got %d", result.Remaining)
	}

	result2, err := engine.Check("test_cost", "user1", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result2.Allowed {
		t.Errorf("expected request to be denied (insufficient tokens)")
	}
}

func TestEngine_Check_TokenRefill(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_refill", 10, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result1, err := engine.Check("test_refill", "user1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result1.Allowed {
		t.Errorf("first request should be allowed")
	}

	result2, err := engine.Check("test_refill", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result2.Allowed {
		t.Errorf("second request should be denied (no tokens)")
	}

	time.Sleep(1100 * time.Millisecond)

	result3, err := engine.Check("test_refill", "user1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result3.Allowed {
		t.Errorf("third request should be allowed after refill")
	}
}

func TestEngine_Check_MinuteInterval(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_minute", 60, "1m", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result, err := engine.Check("test_minute", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}

	if result.Limit != 60 {
		t.Errorf("expected limit to be 60, got %d", result.Limit)
	}
}

func TestEngine_Check_HourInterval(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_hour", 3600, "1h", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result, err := engine.Check("test_hour", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}

	if result.Limit != 3600 {
		t.Errorf("expected limit to be 3600, got %d", result.Limit)
	}
}

func TestEngine_Check_NotFound(t *testing.T) {
	db := setupTestDB(t)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	_, err := engine.Check("nonexistent", "user1", 1)
	if err == nil {
		t.Errorf("expected error for nonexistent rate limit")
	}
}

func TestEngine_Check_ZeroCost(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_zero", 10, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	result, err := engine.Check("test_zero", "user1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected request to be allowed")
	}

	if result.Remaining != 9 {
		t.Errorf("expected remaining to be 9 (cost defaults to 1), got %d", result.Remaining)
	}
}

func TestEngine_Cleanup(t *testing.T) {
	db := setupTestDB(t)
	createTestRateLimit(t, db, "test_cleanup", 10, "1s", 0)

	rateLimitRepo := persistence.NewRateLimitRepository(db)
	rateLimitStateRepo := persistence.NewRateLimitStateRepository(db)
	engine := NewEngine(rateLimitRepo, rateLimitStateRepo)

	_, err := engine.Check("test_cleanup", "user1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	db.Model(&persistence.RateLimitState{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 state record, got %d", count)
	}

	err = engine.Cleanup(1 * time.Nanosecond)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	db.Model(&persistence.RateLimitState{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 state records after cleanup, got %d", count)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"1s", 1 * time.Second, false},
		{"10s", 10 * time.Second, false},
		{"1m", 1 * time.Minute, false},
		{"5m", 5 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"24h", 24 * time.Hour, false},
		{"invalid", 0, true},
		{"s", 0, true},
		{"1x", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}
