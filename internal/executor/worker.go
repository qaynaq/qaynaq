package executor

import (
	"context"

	"github.com/qaynaq/qaynaq/internal/executor/worker"
	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"github.com/qaynaq/qaynaq/internal/vault"

	"google.golang.org/grpc"
)

type IngestResult = worker.IngestResult

type WorkerExecutor interface {
	JoinToCoordinator(context.Context) error
	LeaveCoordinator(context.Context) error
	SendHeartbeat(context.Context) error
	AddFlowToQueue(ctx context.Context, workerFlowID int64, config string, files []*pb.FlowFile) error
	FetchWorkerFlowStatus(ctx context.Context, workerFlowID int64) (*persistence.WorkerFlowStatus, error)
	DeleteWorkerFlow(ctx context.Context, workerFlowID int64) error
	ShipLogs(context.Context)
	ShipMetrics(context.Context)
	ConsumeFlowQueue(context.Context)
	IngestData(ctx context.Context, workerFlowID int64, method, path, contentType string, payload []byte) (*IngestResult, error)
}

type workerExecutor struct {
	worker worker.WorkerExecutor
}

func NewWorkerExecutor(ctx context.Context, grpcConn *grpc.ClientConn, grpcPort uint32, vaultProvider vault.VaultProvider) WorkerExecutor {
	return &workerExecutor{
		worker: worker.NewWorkerExecutor(ctx, grpcConn, grpcPort, vaultProvider),
	}
}

func (e *workerExecutor) JoinToCoordinator(ctx context.Context) error {
	return e.worker.JoinToCoordinator(ctx)
}

func (e *workerExecutor) LeaveCoordinator(ctx context.Context) error {
	return e.worker.LeaveCoordinator(ctx)
}

func (e *workerExecutor) SendHeartbeat(ctx context.Context) error {
	return e.worker.SendHeartbeat(ctx)
}

func (e *workerExecutor) AddFlowToQueue(ctx context.Context, workerFlowID int64, config string, files []*pb.FlowFile) error {
	return e.worker.AddFlowToQueue(ctx, workerFlowID, config, files)
}

func (e *workerExecutor) FetchWorkerFlowStatus(ctx context.Context, workerFlowID int64) (*persistence.WorkerFlowStatus, error) {
	return e.worker.FetchWorkerFlowStatus(ctx, workerFlowID)
}

func (e *workerExecutor) DeleteWorkerFlow(ctx context.Context, workerFlowID int64) error {
	return e.worker.DeleteWorkerFlow(ctx, workerFlowID)
}

func (e *workerExecutor) ShipLogs(ctx context.Context) {
	e.worker.ShipLogs(ctx)
}

func (e *workerExecutor) ShipMetrics(ctx context.Context) {
	e.worker.ShipMetrics(ctx)
}

func (e *workerExecutor) ConsumeFlowQueue(ctx context.Context) {
	e.worker.ConsumeFlowQueue(ctx)
}

func (e *workerExecutor) IngestData(ctx context.Context, workerFlowID int64, method, path, contentType string, payload []byte) (*IngestResult, error) {
	return e.worker.IngestData(ctx, workerFlowID, method, path, contentType, payload)
}
