package executor

import (
	"context"
	"net/http"

	"github.com/qaynaq/qaynaq/internal/executor/coordinator"
	"github.com/qaynaq/qaynaq/internal/persistence"
)

type CoordinatorExecutor interface {
	CheckWorkersAndAssignFlows(context.Context) error
	CheckWorkerHeartbeats(context.Context) error
	CheckFlowLeases(context.Context) error
	ForwardRequestToWorker(context.Context, *http.Request) (int32, []byte, error)
}

type coordinatorExecutor struct {
	coordinator coordinator.CoordinatorExecutor
}

func NewCoordinatorExecutor(
	workerRepo persistence.WorkerRepository,
	flowRepo persistence.FlowRepository,
	flowCacheRepo persistence.FlowCacheRepository,
	flowRateLimitRepo persistence.FlowRateLimitRepository,
	workerFlowRepo persistence.WorkerFlowRepository,
	fileRepo persistence.FileRepository,
	flowWorkerMap coordinator.FlowWorkerMap,
) CoordinatorExecutor {
	return &coordinatorExecutor{
		coordinator: coordinator.NewCoordinatorExecutor(workerRepo, flowRepo, flowCacheRepo, flowRateLimitRepo, workerFlowRepo, fileRepo, flowWorkerMap),
	}
}

func (e *coordinatorExecutor) CheckWorkersAndAssignFlows(ctx context.Context) error {
	return e.coordinator.CheckWorkersAndAssignFlows(ctx)
}

func (e *coordinatorExecutor) CheckWorkerHeartbeats(ctx context.Context) error {
	return e.coordinator.CheckWorkerHeartbeats(ctx)
}

func (e *coordinatorExecutor) CheckFlowLeases(ctx context.Context) error {
	return e.coordinator.CheckFlowLeases(ctx)
}

func (e *coordinatorExecutor) ForwardRequestToWorker(ctx context.Context, r *http.Request) (int32, []byte, error) {
	return e.coordinator.ForwardRequestToWorker(ctx, r)
}
