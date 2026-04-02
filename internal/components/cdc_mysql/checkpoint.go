package cdc_mysql

import (
	"context"
	"sync"

	"github.com/Jeffail/checkpoint"
	"github.com/qaynaq/qaynaq/internal/logger"
)

type binlogCheckpoint struct {
	gtid      string
	file      string
	position  uint32
	batchSize int64
}

type checkpointTracker struct {
	checkpointer          *checkpoint.Uncapped[*binlogCheckpoint]
	mu                    sync.Mutex
	posCache              *positionCache
	posMode               string
	maxPendingCheckpoints int
	logger                *logger.Logger
}

func newCheckpointTracker(posCache *positionCache, posMode string, maxPendingCheckpoints int, log *logger.Logger) *checkpointTracker {
	return &checkpointTracker{
		checkpointer:          checkpoint.NewUncapped[*binlogCheckpoint](),
		posCache:              posCache,
		posMode:               posMode,
		maxPendingCheckpoints: maxPendingCheckpoints,
		logger:                log,
	}
}

func (ct *checkpointTracker) trackEvent(gtid string, file string, position uint32, batchSize int64) func() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	checkpoint := &binlogCheckpoint{
		gtid:      gtid,
		file:      file,
		position:  position,
		batchSize: batchSize,
	}

	releaseFn := ct.checkpointer.Track(checkpoint, batchSize)

	ct.logger.Debug("Tracking checkpoint batch", "gtid", gtid, "file", file, "pos", position, "batch_size", batchSize)

	return func() {
		ct.mu.Lock()
		defer ct.mu.Unlock()

		ct.logger.Debug("Acknowledging checkpoint batch", "gtid", gtid, "batch_size", batchSize, "pending_before", ct.checkpointer.Pending())

		highestCheckpoint := releaseFn()

		ct.logger.Debug("Checkpoint batch acknowledged", "gtid", gtid, "batch_size", batchSize, "pending_after", ct.checkpointer.Pending())

		if highestCheckpoint != nil && *highestCheckpoint != nil {
			ct.logger.Debug("New highest checkpoint achieved, saving position", "highest_gtid", (*highestCheckpoint).gtid)
			ct.savePosition(*highestCheckpoint)
		} else {
			ct.logger.Debug("No new highest checkpoint, waiting for previous batches to complete")
		}
	}
}

func (ct *checkpointTracker) trackEmptyAdvancement(gtid string, file string, position uint32) func() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	pending := ct.checkpointer.Pending()

	if pending == 0 {
		ct.logger.Debug("Immediately saving empty advancement (no pending checkpoints)", "gtid", gtid, "file", file, "pos", position)

		checkpoint := &binlogCheckpoint{
			gtid:      gtid,
			file:      file,
			position:  position,
			batchSize: 0,
		}

		ct.savePosition(checkpoint)

		return func() {
			ct.logger.Debug("Empty advancement already saved", "gtid", gtid)
		}
	}

	ct.logger.Debug("Queueing empty advancement (pending checkpoints exist)", "gtid", gtid, "file", file, "pos", position, "pending_checkpoints", pending)

	checkpoint := &binlogCheckpoint{
		gtid:      gtid,
		file:      file,
		position:  position,
		batchSize: 0,
	}

	releaseFn := ct.checkpointer.Track(checkpoint, 0)

	return func() {
		ct.mu.Lock()
		defer ct.mu.Unlock()

		ct.logger.Debug("Acknowledging queued empty advancement checkpoint", "gtid", gtid, "pending_before", ct.checkpointer.Pending())

		highestCheckpoint := releaseFn()

		ct.logger.Debug("Queued empty advancement acknowledged", "gtid", gtid, "pending_after", ct.checkpointer.Pending())

		if highestCheckpoint != nil && *highestCheckpoint != nil {
			ct.logger.Debug("New highest checkpoint achieved from empty advancement, saving position", "highest_gtid", (*highestCheckpoint).gtid)
			ct.savePosition(*highestCheckpoint)
		} else {
			ct.logger.Debug("No new highest checkpoint from empty advancement, waiting for previous batches to complete")
		}
	}
}

func (ct *checkpointTracker) canAcceptMoreCheckpoints() bool {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	pending := ct.checkpointer.Pending()
	canAccept := int(pending) < ct.maxPendingCheckpoints

	if !canAccept {
		ct.logger.Debug("Backpressure: pending checkpoints at or exceeding max limit", "pending_checkpoints", pending, "max_limit", ct.maxPendingCheckpoints)
	}

	return canAccept
}

func (ct *checkpointTracker) getPendingCount() int64 {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.checkpointer.Pending()
}

func (ct *checkpointTracker) savePosition(checkpoint *binlogCheckpoint) {
	ctx := context.Background()
	if ct.posMode == "gtid" && checkpoint.gtid != "" {
		ct.logger.Debug("Saving GTID position", "gtid_position", checkpoint.gtid)
		ct.posCache.saveGTIDPosition(ctx, checkpoint.gtid)
	} else if ct.posMode == "file" && checkpoint.file != "" && checkpoint.position > 0 {
		ct.logger.Debug("Saving file position", "binlog_file", checkpoint.file, "binlog_pos", checkpoint.position)
		ct.posCache.saveFilePosition(ctx, checkpoint.file, checkpoint.position)
	}
}
