package cdc_mysql

import (
	"context"
	"sync"
	"testing"

	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/warpstreamlabs/bento/public/service"
)

func setupTestCheckpointTracker(t *testing.T, posMode string, maxPending int) (*checkpointTracker, *positionCache) {
	t.Helper()

	resources := service.MockResources(
		service.MockResourcesOptAddCache("test_positions"),
	)

	log := logger.NewFromZerolog("DEBUG", map[string]any{
		"test": "checkpoint_tracker",
	})

	posCache := newPositionCache(resources, "test_positions", "test_key", 0, log)
	tracker := newCheckpointTracker(posCache, posMode, maxPending, log)

	return tracker, posCache
}

func TestNewCheckpointTracker(t *testing.T) {
	tracker, _ := setupTestCheckpointTracker(t, "gtid", 100)

	if tracker == nil {
		t.Fatal("Expected tracker to be created, got nil")
	}

	if tracker.checkpointer == nil {
		t.Error("Expected checkpointer to be initialized")
	}

	if tracker.posMode != "gtid" {
		t.Errorf("Expected posMode 'gtid', got '%s'", tracker.posMode)
	}

	if tracker.maxPendingCheckpoints != 100 {
		t.Errorf("Expected maxPendingCheckpoints 100, got %d", tracker.maxPendingCheckpoints)
	}

	if tracker.logger == nil {
		t.Error("Expected logger to be set")
	}

	if tracker.posCache == nil {
		t.Error("Expected posCache to be set")
	}
}

func TestTrackEventAndAcknowledge(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	gtid := "3E11FA47-71CA-11E1-9E33-C80AA9429562:10"
	ackFn := tracker.trackEvent(gtid, "", 0, 1)

	pending := tracker.getPendingCount()
	if pending != 1 {
		t.Errorf("Expected 1 pending checkpoint, got %d", pending)
	}

	ackFn()

	pending = tracker.getPendingCount()
	if pending != 0 {
		t.Errorf("Expected 0 pending checkpoints after ack, got %d", pending)
	}

	pos, err := posCache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	expectedGTID := "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-10"
	if pos.GTIDSet != expectedGTID {
		t.Errorf("Expected GTID '%s', got '%s'", expectedGTID, pos.GTIDSet)
	}
}

func TestTrackMultipleEventsInOrder(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	ack1 := tracker.trackEvent("UUID:10", "", 0, 1)
	ack2 := tracker.trackEvent("UUID:20", "", 0, 1)
	ack3 := tracker.trackEvent("UUID:30", "", 0, 1)

	if tracker.getPendingCount() != 3 {
		t.Errorf("Expected 3 pending checkpoints, got %d", tracker.getPendingCount())
	}

	ack1()
	pos, _ := posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-10" {
		t.Errorf("Expected position after first ack: UUID:1-10, got %s", pos.GTIDSet)
	}

	ack2()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-20" {
		t.Errorf("Expected position after second ack: UUID:1-20, got %s", pos.GTIDSet)
	}

	ack3()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-30" {
		t.Errorf("Expected position after third ack: UUID:1-30, got %s", pos.GTIDSet)
	}

	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending checkpoints, got %d", tracker.getPendingCount())
	}
}

func TestTrackMultipleEventsOutOfOrder(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	ack1 := tracker.trackEvent("UUID:10", "", 0, 1)
	ack2 := tracker.trackEvent("UUID:20", "", 0, 1)
	ack3 := tracker.trackEvent("UUID:30", "", 0, 1)

	ack2()
	if tracker.getPendingCount() != 3 {
		t.Errorf("Expected 3 pending (not released), got %d", tracker.getPendingCount())
	}

	_, err := posCache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected no position saved when acknowledging out of order")
	}

	ack1()
	pos, _ := posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-20" {
		t.Errorf("Expected position UUID:1-20 after ack1 (which releases both), got %s", pos.GTIDSet)
	}

	ack3()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-30" {
		t.Errorf("Expected position UUID:1-30, got %s", pos.GTIDSet)
	}
}

func TestTrackFilePosition(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "file", 100)
	ctx := context.Background()

	binlogFile := "mysql-bin.000123"
	binlogPos := uint32(4567)

	ackFn := tracker.trackEvent("", binlogFile, binlogPos, 1)
	ackFn()

	pos, err := posCache.loadPosition(ctx)
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
}

func TestTrackEmptyAdvancementNoPending(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	gtid := "UUID:50"
	ackFn := tracker.trackEmptyAdvancement(gtid, "", 0)

	pos, err := posCache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Expected position to be saved immediately, got error: %v", err)
	}

	if pos.GTIDSet != "UUID:1-50" {
		t.Errorf("Expected GTID 'UUID:1-50', got '%s'", pos.GTIDSet)
	}

	ackFn()

	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending checkpoints, got %d", tracker.getPendingCount())
	}
}

func TestTrackEmptyAdvancementWithPending(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	ack1 := tracker.trackEvent("UUID:10", "", 0, 1)
	emptyAck := tracker.trackEmptyAdvancement("UUID:20", "", 0)

	_, err := posCache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected no position saved when empty advancement is queued")
	}

	if tracker.getPendingCount() != 1 {
		t.Errorf("Expected 1 pending checkpoint (empty advancement has 0 batch size), got %d", tracker.getPendingCount())
	}

	ack1()
	pos, _ := posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-10" {
		t.Errorf("Expected position UUID:1-10, got %s", pos.GTIDSet)
	}

	emptyAck()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-20" {
		t.Errorf("Expected position UUID:1-20 after empty ack, got %s", pos.GTIDSet)
	}
}

func TestCanAcceptMoreCheckpoints(t *testing.T) {
	tracker, _ := setupTestCheckpointTracker(t, "gtid", 3)

	if !tracker.canAcceptMoreCheckpoints() {
		t.Error("Expected to accept checkpoints when empty")
	}

	ack1 := tracker.trackEvent("UUID:1", "", 0, 1)
	ack2 := tracker.trackEvent("UUID:2", "", 0, 1)

	if !tracker.canAcceptMoreCheckpoints() {
		t.Error("Expected to accept checkpoints with 2/3 pending")
	}

	ack3 := tracker.trackEvent("UUID:3", "", 0, 1)

	if tracker.canAcceptMoreCheckpoints() {
		t.Error("Expected to reject checkpoints when at max limit")
	}

	ack1()

	if !tracker.canAcceptMoreCheckpoints() {
		t.Error("Expected to accept checkpoints after acknowledging one")
	}

	ack2()
	ack3()

	if !tracker.canAcceptMoreCheckpoints() {
		t.Error("Expected to accept checkpoints when all acknowledged")
	}
}

func TestGetPendingCount(t *testing.T) {
	tracker, _ := setupTestCheckpointTracker(t, "gtid", 100)

	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending initially, got %d", tracker.getPendingCount())
	}

	ack1 := tracker.trackEvent("UUID:1", "", 0, 5)
	if tracker.getPendingCount() != 5 {
		t.Errorf("Expected 5 pending, got %d", tracker.getPendingCount())
	}

	ack2 := tracker.trackEvent("UUID:2", "", 0, 3)
	if tracker.getPendingCount() != 8 {
		t.Errorf("Expected 8 pending, got %d", tracker.getPendingCount())
	}

	ack1()
	if tracker.getPendingCount() != 3 {
		t.Errorf("Expected 3 pending after first ack, got %d", tracker.getPendingCount())
	}

	ack2()
	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending after all acks, got %d", tracker.getPendingCount())
	}
}

func TestSavePositionGTIDMode(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	checkpoint := &binlogCheckpoint{
		gtid:      "UUID:100",
		file:      "",
		position:  0,
		batchSize: 1,
	}

	tracker.savePosition(checkpoint)

	pos, err := posCache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.Mode != "gtid" {
		t.Errorf("Expected mode 'gtid', got '%s'", pos.Mode)
	}

	if pos.GTIDSet != "UUID:1-100" {
		t.Errorf("Expected GTID 'UUID:1-100', got '%s'", pos.GTIDSet)
	}
}

func TestSavePositionFileMode(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "file", 100)
	ctx := context.Background()

	checkpoint := &binlogCheckpoint{
		gtid:      "",
		file:      "mysql-bin.000456",
		position:  8888,
		batchSize: 1,
	}

	tracker.savePosition(checkpoint)

	pos, err := posCache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	if pos.Mode != "file" {
		t.Errorf("Expected mode 'file', got '%s'", pos.Mode)
	}

	if pos.BinlogFile != "mysql-bin.000456" {
		t.Errorf("Expected binlog file 'mysql-bin.000456', got '%s'", pos.BinlogFile)
	}

	if pos.BinlogPos != 8888 {
		t.Errorf("Expected binlog position 8888, got %d", pos.BinlogPos)
	}
}

func TestSavePositionIgnoresEmptyGTID(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	checkpoint := &binlogCheckpoint{
		gtid:      "",
		file:      "mysql-bin.000001",
		position:  100,
		batchSize: 1,
	}

	tracker.savePosition(checkpoint)

	_, err := posCache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected no position saved when GTID is empty in GTID mode")
	}
}

func TestSavePositionIgnoresZeroPosition(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "file", 100)
	ctx := context.Background()

	checkpoint := &binlogCheckpoint{
		gtid:      "UUID:10",
		file:      "mysql-bin.000001",
		position:  0,
		batchSize: 1,
	}

	tracker.savePosition(checkpoint)

	_, err := posCache.loadPosition(ctx)
	if err == nil {
		t.Error("Expected no position saved when position is 0 in file mode")
	}
}

func TestConcurrentTracking(t *testing.T) {
	tracker, _ := setupTestCheckpointTracker(t, "gtid", 1000)

	var wg sync.WaitGroup
	acks := make([]func(), 100)

	for i := range 100 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ackFn := tracker.trackEvent("UUID:"+string(rune(id)), "", 0, 1)
			acks[id] = ackFn
		}(i)
	}

	wg.Wait()

	if tracker.getPendingCount() != 100 {
		t.Errorf("Expected 100 pending checkpoints, got %d", tracker.getPendingCount())
	}

	for i := range 100 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			acks[id]()
		}(i)
	}

	wg.Wait()

	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending checkpoints after all acks, got %d", tracker.getPendingCount())
	}
}

func TestBackpressureThreshold(t *testing.T) {
	maxPending := 5
	tracker, _ := setupTestCheckpointTracker(t, "gtid", maxPending)

	acks := make([]func(), maxPending+1)

	for i := range maxPending {
		if !tracker.canAcceptMoreCheckpoints() {
			t.Errorf("Should accept checkpoint %d/%d", i+1, maxPending)
		}
		acks[i] = tracker.trackEvent("UUID:"+string(rune(i)), "", 0, 1)
	}

	if tracker.canAcceptMoreCheckpoints() {
		t.Error("Should not accept more checkpoints when at limit")
	}

	acks[0]()

	if !tracker.canAcceptMoreCheckpoints() {
		t.Error("Should accept more checkpoints after acknowledging one")
	}

	for i := 1; i < maxPending; i++ {
		acks[i]()
	}
}

func TestEmptyAdvancementMultiple(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	emptyAck1 := tracker.trackEmptyAdvancement("UUID:10", "", 0)

	pos, err := posCache.loadPosition(ctx)
	if err != nil {
		t.Fatalf("Expected immediate save for first empty advancement: %v", err)
	}

	if pos.GTIDSet != "UUID:1-10" {
		t.Errorf("Expected first position UUID:1-10, got %s", pos.GTIDSet)
	}

	emptyAck2 := tracker.trackEmptyAdvancement("UUID:20", "", 0)

	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-20" {
		t.Errorf("Expected position UUID:1-20 after second empty advancement, got %s", pos.GTIDSet)
	}

	emptyAck3 := tracker.trackEmptyAdvancement("UUID:30", "", 0)

	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-30" {
		t.Errorf("Expected position UUID:1-30 after third empty advancement, got %s", pos.GTIDSet)
	}

	emptyAck1()
	emptyAck2()
	emptyAck3()

	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending after all acks, got %d", tracker.getPendingCount())
	}
}

func TestLargeBatchSizes(t *testing.T) {
	tracker, _ := setupTestCheckpointTracker(t, "gtid", 10000)

	ack1 := tracker.trackEvent("UUID:1", "", 0, 1000)
	ack2 := tracker.trackEvent("UUID:2", "", 0, 2000)

	pending := tracker.getPendingCount()
	if pending != 3000 {
		t.Errorf("Expected 3000 pending, got %d", pending)
	}

	ack1()
	pending = tracker.getPendingCount()
	if pending != 2000 {
		t.Errorf("Expected 2000 pending after first ack, got %d", pending)
	}

	ack2()
	pending = tracker.getPendingCount()
	if pending != 0 {
		t.Errorf("Expected 0 pending after all acks, got %d", pending)
	}
}

func TestMixedEventsAndEmptyAdvancements(t *testing.T) {
	tracker, posCache := setupTestCheckpointTracker(t, "gtid", 100)
	ctx := context.Background()

	ack1 := tracker.trackEvent("UUID:10", "", 0, 5)
	emptyAck1 := tracker.trackEmptyAdvancement("UUID:20", "", 0)
	ack2 := tracker.trackEvent("UUID:30", "", 0, 3)
	emptyAck2 := tracker.trackEmptyAdvancement("UUID:40", "", 0)

	ack1()
	pos, _ := posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-10" {
		t.Errorf("Expected UUID:1-10, got %s", pos.GTIDSet)
	}

	emptyAck1()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-20" {
		t.Errorf("Expected UUID:1-20, got %s", pos.GTIDSet)
	}

	ack2()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-30" {
		t.Errorf("Expected UUID:1-30, got %s", pos.GTIDSet)
	}

	emptyAck2()
	pos, _ = posCache.loadPosition(ctx)
	if pos.GTIDSet != "UUID:1-40" {
		t.Errorf("Expected UUID:1-40, got %s", pos.GTIDSet)
	}

	if tracker.getPendingCount() != 0 {
		t.Errorf("Expected 0 pending, got %d", tracker.getPendingCount())
	}
}
