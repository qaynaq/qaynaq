package coordinator

import (
	"context"
	"fmt"
	"net"

	"time"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func extractClientIP(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no peer found")
	}

	clientAddr := p.Addr.String()
	tcpAddr, err := net.ResolveTCPAddr("tcp", clientAddr)
	if err != nil {
		log.Error().Err(err).Str("client_addr", clientAddr).Msg("failed to resolve tcp address")
		return "", fmt.Errorf("failed to resolve tcp address: %w", err)
	}

	clientAddr = tcpAddr.IP.String()
	if ip := net.ParseIP(clientAddr); ip != nil {
		if ip.IsLoopback() {
			clientAddr = "127.0.0.1"
		}
	}

	return clientAddr, nil
}

func (c *CoordinatorAPI) RegisterWorker(ctx context.Context, in *pb.RegisterWorkerRequest) (*pb.CommonResponse, error) {
	var workerEntity *persistence.Worker
	var err error

	clientAddr, err := extractClientIP(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Debug().Str("client_addr", clientAddr).Msg("Client joining to coordinator")

	workerEntity, err = c.workerRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Str("worker_id", in.GetId()).Msg("Failed to find worker")
		return nil, status.Error(codes.Internal, "failed to find worker")
	}

	if workerEntity != nil && workerEntity.Status == persistence.WorkerStatusActive {
		return &pb.CommonResponse{
			Message: "Worker has already been registered",
		}, nil
	}

	if workerEntity != nil {
		workerEntity.Address = fmt.Sprintf("%s:%d", clientAddr, in.GetPort())
		workerEntity.Status = persistence.WorkerStatusActive
	} else {
		workerEntity = &persistence.Worker{
			ID:      in.GetId(),
			Address: fmt.Sprintf("%s:%d", clientAddr, in.GetPort()),
			Status:  persistence.WorkerStatusActive,
		}
	}

	if err = c.workerRepo.AddOrActivate(workerEntity); err != nil {
		log.Error().Err(err).Str("worker_id", in.GetId()).Msg("Failed to register worker")
		return nil, status.Error(codes.Internal, "failed to register worker")
	}

	log.Info().Str("worker_id", workerEntity.ID).Str("address", workerEntity.Address).Msg("Worker registered")

	return &pb.CommonResponse{
		Message: "Worker registered successfully",
	}, nil
}

func (c *CoordinatorAPI) DeregisterWorker(_ context.Context, in *pb.DeregisterWorkerRequest) (*pb.CommonResponse, error) {
	err := c.workerRepo.Deactivate(in.GetId())
	if err != nil {
		log.Error().Err(err).Str("worker_id", in.GetId()).Msg("Failed to deregister")
		return nil, status.Error(codes.Internal, "failed to deregister worker")
	}

	return &pb.CommonResponse{Message: "Worker deregistered successfully"}, nil
}

func (c *CoordinatorAPI) Heartbeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	clientAddr, err := extractClientIP(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Debug().Str("client_addr", clientAddr).Msg("Client heartbeat from coordinator")

	workerEntity, err := c.workerRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Str("worker_id", in.GetId()).Msg("Failed to find worker for heartbeat")
		return nil, status.Error(codes.Internal, "failed to find worker")
	}

	if workerEntity == nil {
		log.Warn().Str("worker_id", in.GetId()).Msg("Worker not found for heartbeat")
		return nil, status.Error(codes.NotFound, "worker not found")
	}

	workerEntity.Address = fmt.Sprintf("%s:%d", clientAddr, in.GetPort())
	workerEntity.Status = persistence.WorkerStatusActive
	if err = c.workerRepo.AddOrActivate(workerEntity); err != nil {
		log.Error().Err(err).Str("worker_id", in.GetId()).Msg("Failed to update worker heartbeat")
		return nil, status.Error(codes.Internal, "failed to update worker heartbeat")
	}

	response := &pb.HeartbeatResponse{
		Message:                     "Heartbeat acknowledged",
		RenewedLeaseWorkerFlowIds: []int64{},
		ExpiredLeaseWorkerFlowIds: []int64{},
	}

	for _, workerFlowID := range in.GetRunningWorkerFlowIds() {
		workerFlow, err := c.workerFlowRepo.FindByID(workerFlowID)

		if err != nil {
			log.Error().Err(err).Str("worker_id", in.GetId()).Int64("worker_flow_id", workerFlowID).Msg("Failed to find worker flow")
			response.ExpiredLeaseWorkerFlowIds = append(response.ExpiredLeaseWorkerFlowIds, workerFlowID)
			continue
		}

		if workerFlow == nil {
			log.Warn().Str("worker_id", in.GetId()).Int64("worker_flow_id", workerFlowID).Msg("Worker flow not found")
			response.ExpiredLeaseWorkerFlowIds = append(response.ExpiredLeaseWorkerFlowIds, workerFlowID)
			continue
		}

		if workerFlow.WorkerID != in.GetId() {
			log.Warn().Str("worker_id", in.GetId()).Int64("worker_flow_id", workerFlowID).Str("actual_worker", workerFlow.WorkerID).Msg("Worker stream belongs to different worker")
			response.ExpiredLeaseWorkerFlowIds = append(response.ExpiredLeaseWorkerFlowIds, workerFlowID)
			continue
		}

		if workerFlow.Status != persistence.WorkerFlowStatusRunning {
			log.Warn().Str("worker_id", in.GetId()).Int64("worker_flow_id", workerFlowID).Str("status", string(workerFlow.Status)).Msg("Flow should not be running")
			response.ExpiredLeaseWorkerFlowIds = append(response.ExpiredLeaseWorkerFlowIds, workerFlowID)
			continue
		}

		newExpiry := time.Now().Add(persistence.FlowLeaseInterval)

		if workerFlow.LeaseExpiresAt.IsZero() {
			log.Info().
				Str("worker_id", in.GetId()).
				Int64("worker_flow_id", workerFlowID).
				Int64("flow_id", workerFlow.FlowID).
				Msg("Initializing lease for existing flow")
		}

		err = c.workerFlowRepo.UpdateLeaseExpiry(workerFlow.ID, newExpiry)
		if err != nil {
			log.Error().Err(err).Int64("worker_flow_id", workerFlow.ID).Msg("Failed to update lease expiry")
			continue
		}

		response.RenewedLeaseWorkerFlowIds = append(response.RenewedLeaseWorkerFlowIds, workerFlowID)
	}

	log.Debug().
		Str("worker_id", in.GetId()).
		Str("address", workerEntity.Address).
		Int("renewed", len(response.RenewedLeaseWorkerFlowIds)).
		Int("expired", len(response.ExpiredLeaseWorkerFlowIds)).
		Msg("Worker heartbeat received")

	return response, nil
}

func (c *CoordinatorAPI) ListWorkers(_ context.Context, in *pb.ListWorkersRequest) (*pb.ListWorkersResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	workerStatuses := []persistence.WorkerStatus{persistence.WorkerStatusActive, persistence.WorkerStatusInactive}

	if in.GetStatus() != "all" {
		workerStatuses = []persistence.WorkerStatus{persistence.WorkerStatus(in.GetStatus())}
	}

	workers, err := c.workerRepo.FindAllByStatuses(workerStatuses...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workers")
		return nil, status.Error(codes.Internal, "failed to list workers")
	}

	result := &pb.ListWorkersResponse{
		Data: make([]*pb.ListWorkersResponse_Worker, 0, len(workers)),
	}

	for _, worker := range workers {
		result.Data = append(result.Data, &pb.ListWorkersResponse_Worker{
			Id:            worker.ID,
			Address:       worker.Address,
			LastHeartbeat: timestamppb.New(worker.LastHeartbeat),
			Status:        string(worker.Status),
		})
	}

	return result, nil
}
