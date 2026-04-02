package cdc_mysql

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/warpstreamlabs/bento/public/service"
)

type batchReader struct {
	maxBatchSize int
	positionMode string
	posCache     *positionCache
	eventBuffer  []binlogEvent
	bufferMutex  sync.Mutex
	pendingEvent *replication.BinlogEvent

	connManager       *connectionManager
	eventProcessor    *eventProcessor
	checkpointTracker *checkpointTracker
	logger            *logger.Logger
}

func newBatchReader(
	maxBatchSize int,
	positionMode string,
	posCache *positionCache,
	connManager *connectionManager,
	eventProcessor *eventProcessor,
	checkpointTracker *checkpointTracker,
	logger *logger.Logger,
) *batchReader {
	return &batchReader{
		maxBatchSize:      maxBatchSize,
		positionMode:      positionMode,
		posCache:          posCache,
		connManager:       connManager,
		eventProcessor:    eventProcessor,
		checkpointTracker: checkpointTracker,
		logger:            logger,
	}
}

func (br *batchReader) readBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	br.logger.Debug("ReadBatch method called")
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	connected, streamer := br.connManager.getConnectionState()
	if !connected || streamer == nil || br.checkpointTracker == nil {
		return nil, nil, service.ErrNotConnected
	}

	for !br.checkpointTracker.canAcceptMoreCheckpoints() {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			br.logger.Debug("Waiting for backpressure relief", "pending_checkpoints", br.checkpointTracker.getPendingCount())
		}
	}

	var allMessages []*service.Message
	var currentGTID string
	var currentFile string
	var currentPosition uint32
	var isGTIDComplete bool

	br.bufferMutex.Lock()
	for len(br.eventBuffer) > 0 && len(allMessages) < br.maxBatchSize {
		event := br.eventBuffer[0]
		br.eventBuffer = br.eventBuffer[1:]

		msg := service.NewMessage(nil)
		msg.SetStructured(event.ToMap())
		if event.GTID != "" {
			msg.MetaSetMut("mysql_gtid", event.GTID)
			currentGTID = event.GTID
		}
		allMessages = append(allMessages, msg)
	}

	bufferEmpty := len(br.eventBuffer) == 0
	br.bufferMutex.Unlock()

	if len(allMessages) > 0 {
		br.logger.Debug("Consumed events from buffer",
			"consumed_from_buffer", len(allMessages),
			"buffer_empty", bufferEmpty,
			"gtid", currentGTID,
			"file", currentFile,
			"position", currentPosition)
	}

	if bufferEmpty && len(allMessages) > 0 {
		isGTIDComplete = true
		br.logger.Debug("Buffer empty - GTID transaction complete", "messages", len(allMessages), "gtid", currentGTID)
	} else if len(allMessages) >= br.maxBatchSize {
		isGTIDComplete = false
		br.logger.Debug("Batch full, buffer has more events - partial GTID", "messages", len(allMessages), "gtid", currentGTID)
	} else {
		var err error
		br.logger.Debug("Collecting new events from stream")
		isGTIDComplete, err = br.collectNewEventsFromStream(ctx, &allMessages, &currentGTID, &currentFile, &currentPosition)
		if err != nil {
			if errors.Is(err, ErrBinlogNotAvailable) {
				br.logger.Error("Binlog position corrupted, purging and requesting reconnection")

				if purgeErr := br.posCache.purgePosition(ctx); purgeErr != nil {
					br.logger.Error("Failed to purge position from cache", "error", purgeErr)
				}

				return nil, nil, service.ErrNotConnected
			}

			if errors.Is(err, replication.ErrNeedSyncAgain) {
				br.logger.Warn("Streamer requested resync during ReadBatch, closing connection and requesting reconnect", "error", err)
				br.connManager.markDisconnected()
				return nil, nil, service.ErrNotConnected
			}

			br.logger.Error("Error from collectNewEventsFromStream, propagating to ReadBatch", "error", err)
			return nil, nil, err
		}
	}

	if len(allMessages) == 0 {
		if isGTIDComplete && br.checkpointTracker != nil {
			pendingCount := br.checkpointTracker.getPendingCount()
			br.logger.Debug("Processing empty advancement for complete filtered GTID", "gtid", currentGTID, "file", currentFile, "pos", currentPosition, "pending_checkpoints", pendingCount)
			ackEventFn := br.checkpointTracker.trackEmptyAdvancement(currentGTID, currentFile, currentPosition)

			ackEventFn()

			return service.MessageBatch{}, func(ctx context.Context, err error) error { return nil }, nil
		} else {
			br.logger.Debug("Empty batch with incomplete GTID, no position advance", "gtid", currentGTID, "file", currentFile, "pos", currentPosition, "isGTIDComplete", isGTIDComplete)
			return service.MessageBatch{}, func(ctx context.Context, err error) error { return nil }, nil
		}
	}

	if isGTIDComplete {
		br.logger.Debug("Created complete GTID batch", "messages", len(allMessages), "gtid", currentGTID, "file", currentFile, "pos", currentPosition, "pending_checkpoints", br.checkpointTracker.getPendingCount())
	} else {
		br.logger.Debug("Created partial GTID batch", "messages", len(allMessages), "gtid", currentGTID, "file", currentFile, "pos", currentPosition, "pending_checkpoints", br.checkpointTracker.getPendingCount())
	}

	var ackEventFn func()
	if isGTIDComplete {
		ackEventFn = br.checkpointTracker.trackEvent(currentGTID, currentFile, currentPosition, 1)
	} else {
		ackEventFn = func() {
			br.logger.Debug("Partial GTID batch acknowledged, no position advance")
		}
	}

	checkpointAckFunc := func(ctx context.Context, err error) error {
		if err == nil {
			if isGTIDComplete {
				br.logger.Debug("Acknowledging complete GTID batch, advancing position", "messages", len(allMessages), "gtid", currentGTID)
			} else {
				br.logger.Debug("Acknowledging partial GTID batch, no position advance", "messages", len(allMessages))
			}
			ackEventFn()
		} else {
			br.logger.Error("Batch processing failed, not advancing position", "error", err)
		}
		return nil
	}

	return allMessages, checkpointAckFunc, nil
}

func (br *batchReader) collectNewEventsFromStream(ctx context.Context, allMessages *[]*service.Message, currentGTID *string, currentFile *string, currentPosition *uint32) (bool, error) {
	if br.pendingEvent != nil {
		events := br.eventProcessor.processEvent(br.pendingEvent)

		if br.positionMode == "gtid" {
			*currentGTID = br.eventProcessor.getCurrentGTID()
		} else {
			*currentFile, *currentPosition = br.eventProcessor.getCurrentFilePosition()
		}

		br.addEventsToMessagesOrBuffer(events, *currentGTID, allMessages)
		br.pendingEvent = nil
	}

	_, streamer := br.connManager.getConnectionState()
	if streamer == nil {
		return false, service.ErrNotConnected
	}

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		eventCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		ev, err := streamer.GetEvent(eventCtx)
		cancel()

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if len(*allMessages) > 0 && *currentGTID != "" {
					br.logger.Debug("Timeout waiting for next event, completing current GTID transaction", "gtid", *currentGTID, "messages", len(*allMessages))
					return true, nil
				}
				return false, nil
			}

			if isBinlogNotAvailableError(err) {
				br.logger.Error("Binlog position/GTID no longer available during streaming", "error", err)
				return false, ErrBinlogNotAvailable
			}

			if errors.Is(err, replication.ErrNeedSyncAgain) {
				br.logger.Warn("Streamer requested resync (ErrNeedSyncAgain)", "error", err)
				return false, replication.ErrNeedSyncAgain
			}

			br.logger.Error("Error getting event", "error", err)
			return false, err
		}

		br.connManager.resetRetryInterval()

		events := br.eventProcessor.processEvent(ev)

		var gtid string
		var file string
		var position uint32

		if br.positionMode == "gtid" {
			gtid = br.eventProcessor.getCurrentGTID()
		} else {
			file, position = br.eventProcessor.getCurrentFilePosition()
		}

		if *currentGTID == "" && gtid != "" {
			*currentGTID = gtid
			*currentFile = file
			*currentPosition = position
		}

		if gtid != "" && gtid != *currentGTID {
			br.pendingEvent = ev
			br.logger.Debug("GTID changed, completing current transaction", "current_gtid", *currentGTID, "new_gtid", gtid)
			return true, nil
		}

		if gtid != "" || (file != "" && position > 0) {
			*currentGTID = gtid
			*currentFile = file
			*currentPosition = position
		}

		if len(events) == 0 {
			br.logger.Debug("Skipping filtered event", "gtid", gtid, "file", file, "pos", position)
			continue
		}

		br.logger.Debug("Adding events to GTID batch", "events", len(events), "gtid", gtid, "file", file, "pos", position)

		if br.addEventsToMessagesOrBuffer(events, gtid, allMessages) {
			return false, nil
		}
	}
}

func (br *batchReader) addEventsToMessagesOrBuffer(events []binlogEvent, gtid string, allMessages *[]*service.Message) bool {
	for i, event := range events {
		if len(*allMessages) < br.maxBatchSize {
			msg := service.NewMessage(nil)
			msg.SetStructured(event.ToMap())
			if gtid != "" {
				msg.MetaSetMut("mysql_gtid", gtid)
			}
			*allMessages = append(*allMessages, msg)
		} else {
			br.bufferMutex.Lock()
			for j := i; j < len(events); j++ {
				br.eventBuffer = append(br.eventBuffer, events[j])
			}
			br.bufferMutex.Unlock()
			br.logger.Debug("Buffered overflow events", "buffered_events", len(events)-i, "gtid", gtid)
			return true
		}
	}
	return false
}
