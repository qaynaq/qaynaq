package cdc_mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/warpstreamlabs/bento/public/service"
)

type positionCache struct {
	resources    *service.Resources
	cacheName    string
	cacheKey     string
	saveInterval time.Duration
	logger       *logger.Logger

	mu                sync.RWMutex
	pendingPosition   *binlogPosition
	lastSavedPosition *binlogPosition
	ticker            *time.Ticker
	stopChan          chan struct{}
	stoppedChan       chan struct{}
	started           bool
}

func newPositionCache(resources *service.Resources, cacheName string, cacheKey string, saveInterval time.Duration, log *logger.Logger) *positionCache {
	return &positionCache{
		resources:    resources,
		cacheName:    cacheName,
		cacheKey:     cacheKey,
		saveInterval: saveInterval,
		logger:       log,
		stopChan:     make(chan struct{}),
		stoppedChan:  make(chan struct{}),
	}
}

// Start begins the periodic save goroutine if saveInterval > 0
func (pc *positionCache) Start(ctx context.Context) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.started {
		pc.logger.Warn("Position cache already started")
		return
	}

	pc.started = true

	if pc.saveInterval > 0 {
		pc.logger.Info("Starting position cache with periodic save", "interval", pc.saveInterval)
		pc.ticker = time.NewTicker(pc.saveInterval)
		go pc.periodicSaveLoop(ctx)
	} else {
		pc.logger.Info("Position cache configured for immediate save mode (interval=0)")
		close(pc.stoppedChan)
	}
}

// Stop halts the periodic save goroutine and flushes any pending position
func (pc *positionCache) Stop(ctx context.Context) {
	pc.mu.Lock()
	if !pc.started {
		pc.mu.Unlock()
		return
	}
	started := pc.started
	pc.mu.Unlock()

	if !started {
		return
	}

	pc.logger.Info("Stopping position cache and flushing pending position")

	close(pc.stopChan)

	if pc.saveInterval > 0 {
		<-pc.stoppedChan
	}

	pc.flush(ctx)

	pc.logger.Info("Position cache stopped")
}

func (pc *positionCache) periodicSaveLoop(ctx context.Context) {
	defer close(pc.stoppedChan)
	defer pc.ticker.Stop()

	for {
		select {
		case <-pc.ticker.C:
			pc.flush(ctx)

		case <-ctx.Done():
			pc.logger.Info("Context cancelled, flushing position before exit")
			pc.flush(context.Background())
			return

		case <-pc.stopChan:
			pc.logger.Debug("Stop signal received in periodic save loop")
			return
		}
	}
}

func (pc *positionCache) flush(ctx context.Context) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.pendingPosition == nil {
		pc.logger.Debug("No pending position to flush")
		return
	}

	if pc.lastSavedPosition != nil && pc.positionsEqual(pc.pendingPosition, pc.lastSavedPosition) {
		pc.logger.Debug("Position unchanged, skipping save")
		return
	}

	pc.logger.Debug("Flushing pending position to cache")
	pc.savePositionLocked(ctx, *pc.pendingPosition)

	positionCopy := *pc.pendingPosition
	pc.lastSavedPosition = &positionCopy
}

func (pc *positionCache) positionsEqual(a, b *binlogPosition) bool {
	if a.Mode != b.Mode {
		return false
	}
	if a.Mode == "gtid" {
		return a.GTIDSet == b.GTIDSet
	}
	return a.BinlogFile == b.BinlogFile && a.BinlogPos == b.BinlogPos
}

func (pc *positionCache) saveGTIDPosition(ctx context.Context, gtid string) {
	gtidRange, err := buildGTIDRange(gtid)
	if err != nil {
		pc.logger.Error("Error building GTID range", "error", err)
		gtidRange = gtid
	}

	position := binlogPosition{
		GTIDSet: gtidRange,
		Mode:    "gtid",
	}

	if pc.saveInterval == 0 {
		pc.mu.Lock()
		pc.savePositionLocked(ctx, position)
		pc.mu.Unlock()
	} else {
		pc.mu.Lock()
		pc.pendingPosition = &position
		pc.mu.Unlock()
	}
}

func (pc *positionCache) saveFilePosition(ctx context.Context, binlogFile string, binlogPos uint32) {
	position := binlogPosition{
		BinlogFile: binlogFile,
		BinlogPos:  binlogPos,
		Mode:       "file",
	}

	if pc.saveInterval == 0 {
		pc.mu.Lock()
		pc.savePositionLocked(ctx, position)
		pc.mu.Unlock()
	} else {
		pc.mu.Lock()
		pc.pendingPosition = &position
		pc.mu.Unlock()
	}
}

// savePositionLocked saves the position to cache (caller must hold lock)
func (pc *positionCache) savePositionLocked(ctx context.Context, position binlogPosition) {
	data, err := json.Marshal(position)
	if err != nil {
		pc.logger.Error("Error marshaling position", "error", err)
		return
	}

	err = pc.resources.AccessCache(ctx, pc.cacheName, func(c service.Cache) {
		if err := c.Set(ctx, pc.cacheKey, data, nil); err != nil {
			pc.logger.Error("Error saving position to cache", "cache_name", pc.cacheName, "cache_key", pc.cacheKey, "error", err)
		} else {
			if position.Mode == "gtid" {
				pc.logger.Debug("Saved GTID position", "gtid_set", position.GTIDSet)
			} else {
				pc.logger.Debug("Saved file position", "binlog_file", position.BinlogFile, "binlog_pos", position.BinlogPos)
			}
		}
	})

	if err != nil {
		pc.logger.Error("Error accessing cache", "cache_name", pc.cacheName, "error", err)
	}
}

func (pc *positionCache) loadPosition(ctx context.Context) (binlogPosition, error) {
	var pos binlogPosition
	var data []byte
	var getErr error

	err := pc.resources.AccessCache(ctx, pc.cacheName, func(c service.Cache) {
		data, getErr = c.Get(ctx, pc.cacheKey)
	})

	if err != nil {
		return binlogPosition{}, fmt.Errorf("failed to access cache: %w", err)
	}

	if getErr != nil {
		return binlogPosition{}, fmt.Errorf("failed to get position from cache: %w", getErr)
	}

	if err := json.Unmarshal(data, &pos); err != nil {
		return binlogPosition{}, fmt.Errorf("failed to parse position data: %w", err)
	}

	return pos, nil
}

func (pc *positionCache) purgePosition(ctx context.Context) error {
	pc.logger.Info("Attempting to purge position from cache", "cache_name", pc.cacheName, "cache_key", pc.cacheKey)

	var deleteErr error
	err := pc.resources.AccessCache(ctx, pc.cacheName, func(c service.Cache) {
		deleteErr = c.Delete(ctx, pc.cacheKey)
	})

	if err != nil {
		pc.logger.Error("Failed to access cache for purge", "cache_name", pc.cacheName, "error", err)
		return fmt.Errorf("failed to access cache: %w", err)
	}

	if deleteErr != nil {
		pc.logger.Error("Failed to delete position from cache", "cache_name", pc.cacheName, "cache_key", pc.cacheKey, "error", deleteErr)
		return fmt.Errorf("failed to delete position: %w", deleteErr)
	}

	pc.logger.Warn("Successfully purged position from cache due to binlog unavailable error", "cache_name", pc.cacheName, "cache_key", pc.cacheKey)
	return nil
}

func buildGTIDRange(gtid string) (string, error) {
	parts := strings.Split(gtid, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid GTID format: %s", gtid)
	}

	sid := parts[0]
	gnoStr := parts[1]

	gno, err := strconv.ParseInt(gnoStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid global number in GTID %s: %v", gtid, err)
	}

	if gno <= 0 {
		return gtid, nil
	}

	return fmt.Sprintf("%s:1-%d", sid, gno), nil
}
