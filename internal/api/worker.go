package api

import (
	"context"

	"github.com/rs/zerolog/log"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/qaynaq/qaynaq/internal/executor"
)

type WorkerAPI struct {
	pb.UnimplementedWorkerServer
	workerExecutor executor.WorkerExecutor
}

func NewWorkerAPI(workerExecutor executor.WorkerExecutor) *WorkerAPI {
	return &WorkerAPI{
		workerExecutor: workerExecutor,
	}
}

func (a *WorkerAPI) AssignFlow(ctx context.Context, in *pb.AssignFlowRequest) (*pb.CommonResponse, error) {
	log.Debug().
		Int64("worker_flow_id", in.GetWorkerFlowId()).
		Str("config", in.GetConfig()).
		Msg("Starting flow for processing")

	err := a.workerExecutor.AddFlowToQueue(ctx, in.GetWorkerFlowId(), in.GetConfig(), in.GetFiles())
	if err != nil {
		log.Error().Err(err).Int64("worker_flow_id", in.GetWorkerFlowId()).Msg("Failed to queue flow")
		return nil, status.Error(codes.Internal, "Failed to queue flow")
	}

	return &pb.CommonResponse{Message: "Flow queued for processing"}, nil
}

func (a *WorkerAPI) FetchFlow(ctx context.Context, in *pb.FetchFlowRequest) (*pb.FetchFlowResponse, error) {
	log.Debug().
		Int64("worker_flow_id", in.GetWorkerFlowId()).
		Msg("Fetching flow for processing")

	flowStatus, err := a.workerExecutor.FetchWorkerFlowStatus(ctx, in.GetWorkerFlowId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch flow")
		return nil, status.Error(codes.Internal, "Failed to fetch flow")
	} else if flowStatus == nil {
		return nil, status.Error(codes.NotFound, "Flow not found in worker")
	}
	return &pb.FetchFlowResponse{
		Status: pb.WorkerFlowStatus(pb.WorkerFlowStatus_value[string(*flowStatus)]),
	}, nil
}

func (a *WorkerAPI) CompleteFlow(ctx context.Context, in *pb.CompleteFlowRequest) (*pb.CommonResponse, error) {
	var err error

	log.Debug().
		Int64("worker_flow_id", in.GetWorkerFlowId()).
		Msg("Starting flow for processing")

	if err = a.workerExecutor.DeleteWorkerFlow(ctx, in.GetWorkerFlowId()); err != nil {
		log.Error().Err(err).Int64("worker_flow_id", in.GetWorkerFlowId()).Msg("Failed to delete worker flow")
		return nil, status.Error(codes.Internal, "Failed to delete flow in worker")
	}

	return &pb.CommonResponse{Message: "Worker Flow has been deleted successfully"}, nil
}

func (a *WorkerAPI) HealthCheck(_ context.Context, _ *emptypb.Empty) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{
		Message: "OK",
	}, nil
}

func (a *WorkerAPI) Ingest(ctx context.Context, in *pb.IngestRequest) (*pb.IngestResponse, error) {

	log.Debug().
		Int64("worker_flow_id", in.GetWorkerFlowId()).
		Bytes("data", in.GetPayload()).
		Msg("Ingesting data")
	ingestResult, err := a.workerExecutor.IngestData(
		ctx,
		in.GetWorkerFlowId(),
		in.GetMethod(),
		in.GetPath(),
		in.GetContentType(),
		in.GetPayload(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Int64("worker_flow_id", in.GetWorkerFlowId()).
			Msg("Failed to ingest data")
		return nil, status.Errorf(codes.Internal, "Failed to ingest data: %v", err)
	}

	return &pb.IngestResponse{
		StatusCode: int32(ingestResult.StatusCode),
		Response:   ingestResult.Response,
	}, nil
}
