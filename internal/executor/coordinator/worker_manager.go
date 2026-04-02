package coordinator

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

type WorkerManager interface {
	GetHealthyWorkers(ctx context.Context) ([]persistence.Worker, error)
	DeactivateWorker(workerID string) error
	FindWorker(workerID string) (*persistence.Worker, error)
	GetWorkerClient(worker *persistence.Worker) (pb.WorkerClient, error)
}

type workerManager struct {
	mu               sync.Mutex
	workerRepo       persistence.WorkerRepository
	workerFlowRepo persistence.WorkerFlowRepository
	clientManager    GRPCClientManager
}

func NewWorkerManager(
	workerRepo persistence.WorkerRepository,
	workerFlowRepo persistence.WorkerFlowRepository,
	clientManager GRPCClientManager,
) WorkerManager {
	return &workerManager{
		workerRepo:       workerRepo,
		workerFlowRepo: workerFlowRepo,
		clientManager:    clientManager,
	}
}

func (m *workerManager) GetHealthyWorkers(ctx context.Context) ([]persistence.Worker, error) {
	workers, err := m.workerRepo.FindAllActiveWithRunningFlowCount()
	if err != nil {
		return nil, err
	}

	return workers, nil
}

func (m *workerManager) DeactivateWorker(workerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.workerRepo.Deactivate(workerID); err != nil {
		log.Error().Err(err).Str("worker_id", workerID).Msg("Failed to deactivate worker")
		return err
	}

	if err := m.workerFlowRepo.StopAllRunningAndWaitingByWorkerID(workerID); err != nil {
		log.Warn().Err(err).Str("worker_id", workerID).Msg("Failed to update all worker flows statuses in worker")
		return err
	}

	m.clientManager.RemoveClient(workerID)

	log.Info().Str("worker_id", workerID).Msg("Worker is unhealthy and deactivated")
	return nil
}

func (m *workerManager) FindWorker(workerID string) (*persistence.Worker, error) {
	return m.workerRepo.FindByID(workerID)
}

func (m *workerManager) GetWorkerClient(worker *persistence.Worker) (pb.WorkerClient, error) {
	return m.clientManager.GetClient(worker)
}
