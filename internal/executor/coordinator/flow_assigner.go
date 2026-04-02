package coordinator

import (
	"container/heap"
	"context"

	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"github.com/qaynaq/qaynaq/internal/utils"
)

type FlowAssigner interface {
	AssignFlows(ctx context.Context) error
}

type flowAssigner struct {
	workerManager    WorkerManager
	flowRepo       persistence.FlowRepository
	workerFlowRepo persistence.WorkerFlowRepository
	configBuilder    ConfigBuilder
	flowWorkerMap  FlowWorkerMap
	noWorkersLogged bool
}

func NewFlowAssigner(
	workerManager WorkerManager,
	flowRepo persistence.FlowRepository,
	workerFlowRepo persistence.WorkerFlowRepository,
	configBuilder ConfigBuilder,
	flowWorkerMap FlowWorkerMap,
) FlowAssigner {
	return &flowAssigner{
		workerManager:    workerManager,
		flowRepo:       flowRepo,
		workerFlowRepo: workerFlowRepo,
		configBuilder:    configBuilder,
		flowWorkerMap:  flowWorkerMap,
	}
}

func (s *flowAssigner) AssignFlows(ctx context.Context) error {
	healthyWorkers, err := s.workerManager.GetHealthyWorkers(ctx)
	if err != nil {
		return err
	}

	workerHeap := &utils.WorkerHeap{}
	heap.Init(workerHeap)
	for _, worker := range healthyWorkers {
		heap.Push(workerHeap, worker)
	}

	flows, err := s.flowRepo.ListAllActiveAndNonAssigned()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list all active and non-assigned flows")
		return err
	}

	if len(flows) > 0 && workerHeap.Len() == 0 {
		if !s.noWorkersLogged {
			log.Warn().Int("waiting_flow_count", len(flows)).Msg("No active workers to assign flows")
			s.noWorkersLogged = true
		}
		return nil
	}

	if s.noWorkersLogged {
		log.Info().Int("worker_count", workerHeap.Len()).Msg("Workers available again")
		s.noWorkersLogged = false
	}

	for _, flow := range flows {
		worker := heap.Pop(workerHeap).(persistence.Worker)
		err = s.assignFlowToWorker(ctx, worker, flow)
		if err != nil {
			log.Error().
				Err(err).
				Str("worker_id", worker.ID).
				Int64("flow_id", flow.ID).
				Msg("Failed to assign job")
		} else {
			log.Debug().
				Str("worker_id", worker.ID).
				Int64("flow_id", flow.ID).
				Msg("Assigned job to worker")
			worker.RunningFlowCount++
		}
		workerHeap.Push(worker)
	}

	return nil
}

func (s *flowAssigner) assignFlowToWorker(ctx context.Context, worker persistence.Worker, flow persistence.Flow) error {
	workerClient, err := s.workerManager.GetWorkerClient(&worker)
	if err != nil {
		log.Error().Err(err).Str("worker_id", worker.ID).Msg("Failed to get worker grpc client")
		return err
	}

	buildResult, err := s.configBuilder.BuildFlowConfig(flow)
	if err != nil {
		return err
	}

	log.Debug().
		Str("worker_id", worker.ID).
		Int64("flow_id", flow.ID).
		Str("config", buildResult.Config).
		Int("files_count", len(buildResult.Files)).
		Msg("Config for worker flow")

	workerFlow, err := s.workerFlowRepo.Queue(worker.ID, flow.ID)
	if err != nil {
		return err
	}

	var flowFiles []*pb.FlowFile
	for _, f := range buildResult.Files {
		flowFiles = append(flowFiles, &pb.FlowFile{
			Key:     f.Key,
			Content: f.Content,
		})
	}

	resp, err := workerClient.AssignFlow(ctx, &pb.AssignFlowRequest{
		WorkerFlowId: workerFlow.ID,
		Config:         buildResult.Config,
		Files:          flowFiles,
	})
	if err != nil {
		if err := s.workerFlowRepo.UpdateStatus(workerFlow.ID, persistence.WorkerFlowStatusFailed); err != nil {
			log.Warn().Err(err).Int64("worker_flow_id", workerFlow.ID).Msg("Failed to update worker flow status after failed assignment")
		}
		// If the worker is unreachable, deactivate it so we stop retrying
		s.workerManager.DeactivateWorker(worker.ID)
		return err
	}

	s.flowWorkerMap.SetFlowWorker(flow.ID, worker.ID, workerFlow.ID)
	if flow.ParentID != nil {
		s.flowWorkerMap.SetFlowWorker(*flow.ParentID, worker.ID, workerFlow.ID)
	}

	log.Info().
		Str("worker_id", worker.ID).
		Int64("flow_id", flow.ID).
		Int64("worker_flow_id", workerFlow.ID).
		Str("worker_stream_assign_response", resp.Message).
		Msg("Assigned job to worker")

	return nil
}
