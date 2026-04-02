package worker

import (
	"context"

	"google.golang.org/grpc"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"github.com/qaynaq/qaynaq/internal/vault"
)

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
	coordinatorConnection CoordinatorConnection
	flowManager         FlowManager
	flowQueue           FlowQueue
	telemetryManager      TelemetryManager
}

func NewWorkerExecutor(ctx context.Context, grpcConn *grpc.ClientConn, grpcPort uint32, vaultProvider vault.VaultProvider) WorkerExecutor {
	coordinatorConnection := NewCoordinatorConnection(ctx, grpcConn, grpcPort)

	flowManager := NewFlowManager(coordinatorConnection, vaultProvider)

	coordinatorConnection.SetFlowManager(flowManager)

	flowQueue := NewFlowQueue(flowManager)

	telemetryManager := NewTelemetryManager(coordinatorConnection, flowManager)

	return &workerExecutor{
		coordinatorConnection: coordinatorConnection,
		flowManager:         flowManager,
		flowQueue:           flowQueue,
		telemetryManager:      telemetryManager,
	}
}

func (e *workerExecutor) JoinToCoordinator(ctx context.Context) error {
	return e.coordinatorConnection.JoinToCoordinator(ctx)
}

func (e *workerExecutor) LeaveCoordinator(ctx context.Context) error {
	return e.coordinatorConnection.LeaveCoordinator(ctx)
}

func (e *workerExecutor) SendHeartbeat(ctx context.Context) error {
	return e.coordinatorConnection.SendHeartbeat(ctx)
}

func (e *workerExecutor) AddFlowToQueue(ctx context.Context, workerFlowID int64, config string, files []*pb.FlowFile) error {
	return e.flowQueue.AddFlowToQueue(workerFlowID, config, files)
}

func (e *workerExecutor) FetchWorkerFlowStatus(ctx context.Context, workerFlowID int64) (*persistence.WorkerFlowStatus, error) {
	return e.flowManager.GetFlowStatus(workerFlowID)
}

func (e *workerExecutor) DeleteWorkerFlow(ctx context.Context, workerFlowID int64) error {
	return e.flowManager.DeleteFlow(workerFlowID)
}

func (e *workerExecutor) ShipLogs(ctx context.Context) {
	e.telemetryManager.ShipLogs(ctx)
}

func (e *workerExecutor) ShipMetrics(ctx context.Context) {
	e.telemetryManager.ShipMetrics(ctx)
}

func (e *workerExecutor) ConsumeFlowQueue(ctx context.Context) {
	e.flowQueue.ConsumeFlowQueue(ctx)
}

func (e *workerExecutor) IngestData(ctx context.Context, workerFlowID int64, method, path, contentType string, payload []byte) (*IngestResult, error) {
	return e.flowManager.IngestData(workerFlowID, method, path, contentType, payload)
}
