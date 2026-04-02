package cdc_mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/warpstreamlabs/bento/public/service"
)

func setupTestPositionCache(t *testing.T) (*positionCache, *service.Resources, *logger.Logger) {
	t.Helper()

	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)

	log := logger.NewFromZerolog("DEBUG", map[string]any{
		"test": "position_cache",
	})

	cache := newPositionCache(resources, "test_positions", "test_key", 0, log)
	return cache, resources, log
}

func TestNewPositionCache(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("positions"),
	)

	log := logger.NewFromZerolog("INFO", nil)

	cache := newPositionCache(resources, "positions", "db1", 0, log)

	if cache == nil {
		t.Fatal("Expected cache to be created, got nil")
	}

	if cache.cacheName != "positions" {
		t.Errorf("Expected cacheName 'positions', got '%s'", cache.cacheName)
	}

	if cache.cacheKey != "db1" {
		t.Errorf("Expected cacheKey 'db1', got '%s'", cache.cacheKey)
	}

	if cache.resources == nil {
		t.Error("Expected resources to be set")
	}

	if cache.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestSaveAndLoadGTIDPosition(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	gtid := "3E11FA47-71CA-11E1-9E33-C80AA9429562:23"

	cache.saveGTIDPosition(ctx, gtid)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.Mode != "gtid" {
		t.Errorf("Expected mode 'gtid', got '%s'", pos.Mode)
	}

	expectedGTIDSet := "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-23"
	if pos.GTIDSet != expectedGTIDSet {
		t.Errorf("Expected GTID set '%s', got '%s'", expectedGTIDSet, pos.GTIDSet)
	}

	if pos.BinlogFile != "" {
		t.Errorf("Expected empty BinlogFile, got '%s'", pos.BinlogFile)
	}

	if pos.BinlogPos != 0 {
		t.Errorf("Expected BinlogPos 0, got %d", pos.BinlogPos)
	}
}

func TestSaveAndLoadFilePosition(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	binlogFile := "mysql-bin.000123"
	binlogPos := uint32(4567)

	cache.saveFilePosition(ctx, binlogFile, binlogPos)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.Mode != "file" {
		t.Errorf("Expected mode 'file', got '%s'", pos.Mode)
	}

	if pos.BinlogFile != binlogFile {
		t.Errorf("Expected binlog file '%s', got '%s'", binlogFile, pos.BinlogFile)
	}

	if pos.BinlogPos != binlogPos {
		t.Errorf("Expected binlog position %d, got %d", binlogPos, pos.BinlogPos)
	}

	if pos.GTIDSet != "" {
		t.Errorf("Expected empty GTIDSet, got '%s'", pos.GTIDSet)
	}
}

func TestSavePositionOverwrite(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	cache.saveFilePosition(ctx, "mysql-bin.000001", 100)

	pos1, _ := cache.loadPosition(ctx)
	if pos1.BinlogFile != "mysql-bin.000001" || pos1.BinlogPos != 100 {
		t.Fatal("First save failed")
	}

	cache.saveFilePosition(ctx, "mysql-bin.000002", 200)

	pos2, _ := cache.loadPosition(ctx)
	if pos2.BinlogFile != "mysql-bin.000002" {
		t.Errorf("Expected binlog file 'mysql-bin.000002', got '%s'", pos2.BinlogFile)
	}
	if pos2.BinlogPos != 200 {
		t.Errorf("Expected binlog position 200, got %d", pos2.BinlogPos)
	}
}

func TestLoadPositionNotFound(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	_, err := cache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected error when loading non-existent position, got nil")
	}
}

func TestPurgePosition(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")

	_, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Position should exist before purge: %v", err)
	}

	err = cache.purgePosition(ctx)
	if err != nil {
		t.Fatalf("Failed to purge position: %v", err)
	}

	_, err = cache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected error after purge, position should not exist")
	}
}

func TestPurgePositionNonExistent(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	err := cache.purgePosition(ctx)
	if err != nil {
		t.Errorf("Purging non-existent position should not error, got: %v", err)
	}
}

func TestSavePositionWithInvalidCache(t *testing.T) {
	resources := service.MockResources()
	log := logger.NewFromZerolog("ERROR", nil)

	cache := newPositionCache(resources, "nonexistent_cache", "test_key", 0, log)
	ctx := context.Background()

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")

	_, err := cache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected error when accessing non-existent cache")
	}
}

func TestBuildGTIDRange(t *testing.T) {
	tests := []struct {
		name        string
		gtid        string
		expected    string
		expectError bool
	}{
		{
			name:        "valid GTID with positive number",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562:23",
			expected:    "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-23",
			expectError: false,
		},
		{
			name:        "valid GTID with large number",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562:999999",
			expected:    "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-999999",
			expectError: false,
		},
		{
			name:        "GTID with zero",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562:0",
			expected:    "3E11FA47-71CA-11E1-9E33-C80AA9429562:0",
			expectError: false,
		},
		{
			name:        "GTID with negative number",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562:-1",
			expected:    "3E11FA47-71CA-11E1-9E33-C80AA9429562:-1",
			expectError: false,
		},
		{
			name:        "invalid GTID format - no colon",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid GTID format - multiple colons",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562:23:45",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid GTID format - non-numeric transaction number",
			gtid:        "3E11FA47-71CA-11E1-9E33-C80AA9429562:abc",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty string",
			gtid:        "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildGTIDRange(tt.gtid)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for GTID '%s', got nil", tt.gtid)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for GTID '%s': %v", tt.gtid, err)
				}
				if result != tt.expected {
					t.Errorf("Expected GTID range '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestSaveGTIDPositionWithInvalidGTID(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	invalidGTID := "invalid-gtid-format"

	cache.saveGTIDPosition(ctx, invalidGTID)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.GTIDSet != invalidGTID {
		t.Errorf("Expected GTID set to be saved as-is when invalid: '%s', got '%s'", invalidGTID, pos.GTIDSet)
	}
}

func TestMultipleCacheKeys(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx := context.Background()

	cache1 := newPositionCache(resources, "positions", "db1", 0, log)
	cache2 := newPositionCache(resources, "positions", "db2", 0, log)

	cache1.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")
	cache2.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:20")

	pos1, err := cache1.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position for cache1: %v", err)
	}

	pos2, err := cache2.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position for cache2: %v", err)
	}

	if pos1.GTIDSet == pos2.GTIDSet {
		t.Error("Expected different GTID sets for different cache keys")
	}

	if pos1.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10" {
		t.Errorf("Expected cache1 GTID '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10', got '%s'", pos1.GTIDSet)
	}

	if pos2.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-20" {
		t.Errorf("Expected cache2 GTID '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-20', got '%s'", pos2.GTIDSet)
	}
}

func TestPositionJSONSerialization(t *testing.T) {
	cache, resources, _ := setupTestPositionCache(t)
	ctx := context.Background()

	testCases := []struct {
		name     string
		position binlogPosition
	}{
		{
			name: "GTID position",
			position: binlogPosition{
				Mode:    "gtid",
				GTIDSet: "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-100",
			},
		},
		{
			name: "File position",
			position: binlogPosition{
				Mode:       "file",
				BinlogFile: "mysql-bin.000123",
				BinlogPos:  4567,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.position)
			if err != nil {
				t.Fatalf("Failed to marshal position: %v", err)
			}

			err = resources.AccessCache(ctx, "test_positions", func(c service.Cache) {
				if err := c.Set(ctx, "test_key", data, nil); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
			})
			if err != nil {
				t.Fatalf("Failed to access cache: %v", err)
			}

			loaded, err := cache.loadPosition(ctx)
			if err != nil {
				t.Fatalf("Failed to load position: %v", err)
			}

			if loaded.Mode != tc.position.Mode {
				t.Errorf("Mode mismatch: expected '%s', got '%s'", tc.position.Mode, loaded.Mode)
			}

			if loaded.GTIDSet != tc.position.GTIDSet {
				t.Errorf("GTIDSet mismatch: expected '%s', got '%s'", tc.position.GTIDSet, loaded.GTIDSet)
			}

			if loaded.BinlogFile != tc.position.BinlogFile {
				t.Errorf("BinlogFile mismatch: expected '%s', got '%s'", tc.position.BinlogFile, loaded.BinlogFile)
			}

			if loaded.BinlogPos != tc.position.BinlogPos {
				t.Errorf("BinlogPos mismatch: expected %d, got %d", tc.position.BinlogPos, loaded.BinlogPos)
			}
		})
	}
}

func TestLoadPositionWithCorruptedData(t *testing.T) {
	cache, resources, _ := setupTestPositionCache(t)
	ctx := context.Background()

	corruptedData := []byte("this is not valid JSON")

	err := resources.AccessCache(ctx, "test_positions", func(c service.Cache) {
		if err := c.Set(ctx, "test_key", corruptedData, nil); err != nil {
			t.Fatalf("Failed to set corrupted data: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to access cache: %v", err)
	}

	_, err = cache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected error when loading corrupted JSON data, got nil")
	}
}

func TestConcurrentAccessToCache(t *testing.T) {
	cache, _, _ := setupTestPositionCache(t)
	ctx := context.Background()

	done := make(chan bool)

	for i := range 10 {
		go func(id int) {
			gtid := fmt.Sprintf("3E11FA47-71CA-11E1-9E33-C80AA9429562:%d", id)
			cache.saveGTIDPosition(ctx, gtid)
			done <- true
		}(i)
	}

	for range 10 {
		<-done
	}

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position after concurrent writes: %v", err)
	}

	if pos.Mode != "gtid" {
		t.Errorf("Expected mode 'gtid', got '%s'", pos.Mode)
	}

	if pos.GTIDSet == "" {
		t.Error("Expected non-empty GTID set after concurrent writes")
	}
}

func TestIntervalBasedCaching(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx := context.Background()

	cache := newPositionCache(resources, "test_positions", "test_key", 100*time.Millisecond, log)
	cache.Start(ctx)
	defer cache.Stop(ctx)

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")

	time.Sleep(10 * time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position after interval: %v", err)
	}

	if pos.Mode != "gtid" {
		t.Errorf("Expected mode 'gtid', got '%s'", pos.Mode)
	}

	if pos.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10" {
		t.Errorf("Expected GTID set '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10', got '%s'", pos.GTIDSet)
	}
}

func TestImmediateSaveMode(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx := context.Background()

	cache := newPositionCache(resources, "test_positions", "test_key", 0, log)
	cache.Start(ctx)
	defer cache.Stop(ctx)

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position immediately: %v", err)
	}

	if pos.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10" {
		t.Errorf("Expected GTID set '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10', got '%s'", pos.GTIDSet)
	}
}

func TestStopFlushesPendingPosition(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx := context.Background()

	cache := newPositionCache(resources, "test_positions", "test_key", 1*time.Hour, log)
	cache.Start(ctx)

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:50")

	cache.Stop(ctx)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position after stop: %v", err)
	}

	if pos.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-50" {
		t.Errorf("Expected GTID set '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-50', got '%s'", pos.GTIDSet)
	}
}

func TestContextCancellationFlushes(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx, cancel := context.WithCancel(context.Background())

	cache := newPositionCache(resources, "test_positions", "test_key", 1*time.Hour, log)
	cache.Start(ctx)

	cache.saveFilePosition(ctx, "mysql-bin.000123", 4567)

	cancel()
	time.Sleep(50 * time.Millisecond)

	cache.Stop(context.Background())

	pos, err := cache.loadPosition(context.Background())
	if err != nil {
		t.Fatalf("Failed to load position after context cancellation: %v", err)
	}

	if pos.BinlogFile != "mysql-bin.000123" {
		t.Errorf("Expected binlog file 'mysql-bin.000123', got '%s'", pos.BinlogFile)
	}

	if pos.BinlogPos != 4567 {
		t.Errorf("Expected binlog pos 4567, got %d", pos.BinlogPos)
	}
}

func TestMultipleUpdatesWithinInterval(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx := context.Background()

	cache := newPositionCache(resources, "test_positions", "test_key", 200*time.Millisecond, log)
	cache.Start(ctx)
	defer cache.Stop(ctx)

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")
	time.Sleep(10 * time.Millisecond)
	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:20")
	time.Sleep(10 * time.Millisecond)
	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:30")

	time.Sleep(250 * time.Millisecond)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-30" {
		t.Errorf("Expected GTID set '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-30', got '%s'", pos.GTIDSet)
	}
}

func TestNoSaveWhenPositionUnchanged(t *testing.T) {
	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)
	log := logger.NewFromZerolog("DEBUG", nil)
	ctx := context.Background()

	cache := newPositionCache(resources, "test_positions", "test_key", 100*time.Millisecond, log)
	cache.Start(ctx)
	defer cache.Stop(ctx)

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")
	time.Sleep(150 * time.Millisecond)

	cache.saveGTIDPosition(ctx, "3E11FA47-71CA-11E1-9E33-C80AA9429562:10")
	time.Sleep(150 * time.Millisecond)

	pos, err := cache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.GTIDSet != "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10" {
		t.Errorf("Expected GTID set '3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10', got '%s'", pos.GTIDSet)
	}
}
