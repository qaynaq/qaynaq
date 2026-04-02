package coordinator

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

type CoordinatorExecutor interface {
	CheckWorkersAndAssignFlows(context.Context) error
	CheckWorkerHeartbeats(context.Context) error
	CheckFlowLeases(context.Context) error
	ForwardRequestToWorker(context.Context, *http.Request) (int32, []byte, error)
}

type coordinatorExecutor struct {
	flowAssigner   FlowAssigner
	requestForwarder RequestForwarder
	flowWorkerMap  FlowWorkerMap
	workerFlowRepo persistence.WorkerFlowRepository
	workerRepo       persistence.WorkerRepository
}

func NewCoordinatorExecutor(
	workerRepo persistence.WorkerRepository,
	flowRepo persistence.FlowRepository,
	flowCacheRepo persistence.FlowCacheRepository,
	flowRateLimitRepo persistence.FlowRateLimitRepository,
	workerFlowRepo persistence.WorkerFlowRepository,
	fileRepo persistence.FileRepository,
	flowWorkerMap FlowWorkerMap,
) CoordinatorExecutor {
	clientManager := NewGRPCClientManager()
	workerManager := NewWorkerManager(workerRepo, workerFlowRepo, clientManager)
	configBuilder := NewConfigBuilder(flowCacheRepo, flowRateLimitRepo, fileRepo)

	err := initializeFlowWorkerMapping(workerFlowRepo, flowWorkerMap)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize flow-worker mapping")
	}

	flowAssigner := NewFlowAssigner(workerManager, flowRepo, workerFlowRepo, configBuilder, flowWorkerMap)
	requestForwarder := NewRequestForwarder(workerManager, flowWorkerMap, flowRepo)

	return &coordinatorExecutor{
		flowAssigner:   flowAssigner,
		requestForwarder: requestForwarder,
		flowWorkerMap:  flowWorkerMap,
		workerFlowRepo: workerFlowRepo,
		workerRepo:       workerRepo,
	}
}

func (e *coordinatorExecutor) CheckWorkersAndAssignFlows(ctx context.Context) error {
	return e.flowAssigner.AssignFlows(ctx)
}

func (e *coordinatorExecutor) CheckWorkerHeartbeats(ctx context.Context) error {
	workers, err := e.workerRepo.FindActiveWithStaleHeartbeat()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch workers with stale heartbeats")
		return err
	}

	for _, worker := range workers {
		timeSinceHeartbeat := time.Since(worker.LastHeartbeat)

		log.Warn().
			Str("worker_id", worker.ID).
			Dur("time_since_heartbeat", timeSinceHeartbeat).
			Msg("Worker heartbeat timeout - marking as inactive")

		err := e.workerRepo.Deactivate(worker.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("worker_id", worker.ID).
				Msg("Failed to deactivate worker")
			continue
		}

		err = e.workerFlowRepo.StopAllRunningAndWaitingByWorkerID(worker.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("worker_id", worker.ID).
				Msg("Failed to stop worker flows")
			continue
		}

		log.Info().
			Str("worker_id", worker.ID).
			Msg("Worker marked as inactive due to heartbeat timeout")
	}

	return nil
}

func (e *coordinatorExecutor) CheckFlowLeases(ctx context.Context) error {
	expiredFlows, err := e.workerFlowRepo.FindRunningWithExpiredLeases()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch flows with expired leases")
		return err
	}

	for _, workerFlow := range expiredFlows {
		timeSinceExpiry := time.Since(workerFlow.LeaseExpiresAt)

		log.Warn().
			Str("worker_id", workerFlow.WorkerID).
			Int64("flow_id", workerFlow.FlowID).
			Int64("worker_flow_id", workerFlow.ID).
			Dur("time_since_expiry", timeSinceExpiry).
			Msg("Flow lease expired - marking as stopped")

		err := e.workerFlowRepo.UpdateStatus(
			workerFlow.ID,
			persistence.WorkerFlowStatusStopped,
		)
		if err != nil {
			log.Error().
				Err(err).
				Int64("worker_flow_id", workerFlow.ID).
				Msg("Failed to mark expired flow as stopped")
			continue
		}

		log.Info().
			Str("worker_id", workerFlow.WorkerID).
			Int64("flow_id", workerFlow.FlowID).
			Msg("Expired flow marked as stopped - available for reassignment")
	}

	return nil
}

func (e *coordinatorExecutor) ForwardRequestToWorker(ctx context.Context, r *http.Request) (int32, []byte, error) {
	return e.requestForwarder.ForwardRequestToWorker(ctx, r)
}

func initializeFlowWorkerMapping(workerFlowRepo persistence.WorkerFlowRepository, flowWorkerMap FlowWorkerMap) error {
	workerFlows, err := workerFlowRepo.ListAllByStatuses(persistence.WorkerFlowStatusRunning)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list all running worker flows, suggesting to restart the coordinator, otherwise HTTP streams will not be reachable")
		return err
	}

	for _, workerFlow := range workerFlows {
		flowWorkerMap.SetFlowWorker(workerFlow.FlowID, workerFlow.Worker.ID, workerFlow.ID)
		if workerFlow.Flow.ParentID != nil {
			flowWorkerMap.SetFlowWorker(*workerFlow.Flow.ParentID, workerFlow.Worker.ID, workerFlow.ID)
		}
	}

	log.Info().Int("running_worker_stream_count", len(workerFlows)).Msg("Loaded running worker flows")
	return nil
}
