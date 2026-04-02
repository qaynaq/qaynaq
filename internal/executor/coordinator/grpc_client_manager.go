package coordinator

import (
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

type GRPCClientManager interface {
	GetClient(worker *persistence.Worker) (pb.WorkerClient, error)
	RemoveClient(workerID string)
}

type workerClientEntry struct {
	client  pb.WorkerClient
	conn    *grpc.ClientConn
	address string
}

type grpcClientManager struct {
	mu            sync.Mutex
	workerClients map[string]*workerClientEntry
}

func NewGRPCClientManager() GRPCClientManager {
	return &grpcClientManager{
		workerClients: make(map[string]*workerClientEntry),
	}
}

func (m *grpcClientManager) GetClient(worker *persistence.Worker) (pb.WorkerClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, exists := m.workerClients[worker.ID]; exists {
		if entry.address == worker.Address {
			return entry.client, nil
		}
		// worker address changed (e.g. after restart) - close stale connection
		log.Info().Str("worker_id", worker.ID).Str("old", entry.address).Str("new", worker.Address).Msg("Worker address changed, reconnecting")
		entry.conn.Close()
		delete(m.workerClients, worker.ID)
	}

	log.Debug().Str("worker_id", worker.ID).Str("address", worker.Address).Msg("Creating new grpc client for worker")
	grpcConn, err := grpc.NewClient(worker.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewWorkerClient(grpcConn)
	m.workerClients[worker.ID] = &workerClientEntry{
		client:  client,
		conn:    grpcConn,
		address: worker.Address,
	}

	return client, nil
}

func (m *grpcClientManager) RemoveClient(workerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, exists := m.workerClients[workerID]; exists {
		entry.conn.Close()
		delete(m.workerClients, workerID)
	}
	log.Debug().Str("worker_id", workerID).Msg("Removed grpc client for worker")
}
