package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) UpdateWorkerFlowStatus(_ context.Context, in *pb.WorkerFlowStatusRequest) (*pb.CommonResponse, error) {
	var err error

	workerFlow, err := c.workerFlowRepo.FindByID(in.GetWorkerFlowId())
	if err != nil {
		log.Error().Err(err).Int64("worker_flow_id", in.GetWorkerFlowId()).Msg("Failed to find worker flow")
		return nil, status.Error(codes.Internal, "Failed to find worker flow")
	} else if workerFlow == nil {
		log.Error().Err(err).Int64("worker_flow_id", in.GetWorkerFlowId()).Msg("Worker flow not found")
		return nil, status.Error(codes.NotFound, "worker flow not found")
	}

	newStatus := persistence.WorkerFlowStatus(in.GetStatus().String())

	if err = c.workerFlowRepo.UpdateStatus(in.GetWorkerFlowId(), newStatus); err != nil {
		log.Error().Err(err).Str("target_status", string(newStatus)).Msg("Failed to update worker flow")
		return nil, status.Error(codes.Internal, "Failed to update worker flow status")
	}

	switch newStatus {
	case persistence.WorkerFlowStatusRunning:
		c.flowWorkerMap.SetFlowWorker(workerFlow.FlowID, workerFlow.WorkerID, workerFlow.ID)
		if workerFlow.Flow.ParentID != nil {
			c.flowWorkerMap.SetFlowWorker(*workerFlow.Flow.ParentID, workerFlow.WorkerID, workerFlow.ID)
		}
	case persistence.WorkerFlowStatusFailed, persistence.WorkerFlowStatusStopped:
		c.flowWorkerMap.RemoveFlowIfMatches(workerFlow.FlowID, workerFlow.ID)
		if workerFlow.Flow.ParentID != nil {
			c.flowWorkerMap.RemoveFlowIfMatches(*workerFlow.Flow.ParentID, workerFlow.ID)
		}
	case persistence.WorkerFlowStatusCompleted:
		c.flowWorkerMap.RemoveFlowIfMatches(workerFlow.FlowID, workerFlow.ID)
		if workerFlow.Flow.ParentID != nil {
			c.flowWorkerMap.RemoveFlowIfMatches(*workerFlow.Flow.ParentID, workerFlow.ID)
		}
		if err = c.flowRepo.UpdateStatus(workerFlow.FlowID, persistence.FlowStatusCompleted); err != nil {
			log.Error().Err(err).Str("target_status", string(persistence.FlowStatusCompleted)).Msg("Failed to update flow status")
			return nil, status.Error(codes.Internal, "Failed to update worker flow status")
		}
	}

	return &pb.CommonResponse{
		Message: "Worker Flow status has been updated successfully",
	}, nil
}
