package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

const (
	MaxItemsInFlowQueue = 10
	EventStreamMaxDelay = 30 * time.Second
)

type FlowQueueItem struct {
	WorkerFlowID int64
	Config       string
	Files        []*pb.FlowFile
}

type FlowQueue interface {
	AddFlowToQueue(workerFlowID int64, config string, files []*pb.FlowFile) error
	ConsumeFlowQueue(ctx context.Context)
}

type flowQueue struct {
	queue                 chan FlowQueueItem
	flowManager           FlowManager
	coordinatorConnection CoordinatorConnection
}

func NewFlowQueue(flowManager FlowManager, coordinatorConnection CoordinatorConnection) FlowQueue {
	return &flowQueue{
		queue:                 make(chan FlowQueueItem, MaxItemsInFlowQueue),
		flowManager:           flowManager,
		coordinatorConnection: coordinatorConnection,
	}
}

func (q *flowQueue) AddFlowToQueue(workerFlowID int64, config string, files []*pb.FlowFile) error {
	item := FlowQueueItem{
		WorkerFlowID: workerFlowID,
		Config:       config,
		Files:        files,
	}

	select {
	case q.queue <- item:
		log.Info().Int64("worker_flow_id", workerFlowID).Msg("Stream added to queue")
		return nil
	case <-time.After(EventStreamMaxDelay):
		log.Error().Int64("worker_flow_id", workerFlowID).Msg("Failed to add stream to queue: timeout")
		return ErrFlowQueueTimeout
	}
}

func (q *flowQueue) ConsumeFlowQueue(ctx context.Context) {
	for {
		select {
		case item := <-q.queue:
			log.Info().Int64("worker_flow_id", item.WorkerFlowID).Msg("Processing stream from queue")

			if err := q.flowManager.WriteFiles(item.Files); err != nil {
				log.Error().Err(err).Int64("worker_flow_id", item.WorkerFlowID).Msg("Failed to write files to disk")
				q.reportFailed(ctx, item.WorkerFlowID)
				continue
			}

			if err := q.flowManager.AddFlow(item.WorkerFlowID, item.Config); err != nil {
				log.Error().Err(err).Int64("worker_flow_id", item.WorkerFlowID).Msg("Failed to add stream to manager")
				q.reportFailed(ctx, item.WorkerFlowID)
				continue
			}

			q.flowManager.StartFlow(ctx, item.WorkerFlowID)

		case <-ctx.Done():
			log.Info().Msg("Stopping flow queue processing...")
			q.flowManager.StopAllFlows()

			time.Sleep(1 * time.Second)

			log.Info().Msg("Flow queue processing stopped")
			return
		}
	}
}

func (q *flowQueue) reportFailed(ctx context.Context, workerFlowID int64) {
	if q.coordinatorConnection == nil {
		return
	}
	rpcCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	failed := pb.WorkerFlowStatus(pb.WorkerFlowStatus_value[string(persistence.WorkerFlowStatusFailed)])
	if err := q.coordinatorConnection.UpdateWorkerFlowStatus(rpcCtx, workerFlowID, failed); err != nil {
		log.Warn().Err(err).Int64("worker_flow_id", workerFlowID).Msg("Failed to report worker flow init failure")
	}
}

var (
	ErrFlowQueueTimeout = fmt.Errorf("flow queue timeout")
)
