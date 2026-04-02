package worker

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

type CoordinatorConnection interface {
	JoinToCoordinator(ctx context.Context) error
	LeaveCoordinator(ctx context.Context) error
	SendHeartbeat(ctx context.Context) error
	SetFlowManager(flowManager any)
	GetClient() pb.CoordinatorClient
	UpdateWorkerFlowStatus(ctx context.Context, workerFlowID int64, status pb.WorkerFlowStatus) error
	IngestMetrics(ctx context.Context, workerFlowID int64, inputEvents, processorErrors, outputEvents uint64) error
}

type flowManagerInterface interface {
	GetRunningFlowIDs() []int64
	StopFlow(workerFlowID int64) error
}

type coordinatorConnection struct {
	mu                sync.Mutex
	grpcConn          *grpc.ClientConn
	coordinatorClient pb.CoordinatorClient
	flowManager     flowManagerInterface
	grpcPort          uint32
	joined            bool
}

func NewCoordinatorConnection(ctx context.Context, grpcConn *grpc.ClientConn, grpcPort uint32) CoordinatorConnection {
	coordinatorGRPCClient := pb.NewCoordinatorClient(grpcConn)

	connection := &coordinatorConnection{
		grpcConn:          grpcConn,
		coordinatorClient: coordinatorGRPCClient,
		grpcPort:          grpcPort,
		joined:            false,
	}

	go connection.monitorConnection(ctx)

	return connection
}

func (c *coordinatorConnection) monitorConnection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			state := c.grpcConn.GetState()
			if state == connectivity.TransientFailure || state == connectivity.Shutdown {
				log.Warn().Str("state", state.String()).Msg("gRPC connection is down")
				c.mu.Lock()
				c.joined = false
				c.mu.Unlock()
			}
		}
	}
}

func (c *coordinatorConnection) JoinToCoordinator(ctx context.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	c.mu.Lock()
	if c.joined {
		log.Debug().Str("worker_id", hostname).Msg("Worker already joined to coordinator")
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	r, err := c.coordinatorClient.RegisterWorker(ctx, &pb.RegisterWorkerRequest{
		Id:   hostname,
		Port: c.grpcPort,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to register on coordinator")
		return err
	}

	c.mu.Lock()
	c.joined = true
	c.mu.Unlock()

	log.Info().Str("worker_id", hostname).Msg(r.GetMessage())
	return nil
}

func (c *coordinatorConnection) LeaveCoordinator(ctx context.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	resp, err := c.coordinatorClient.DeregisterWorker(ctx, &pb.DeregisterWorkerRequest{Id: hostname})
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.joined = false
	c.mu.Unlock()

	log.Info().Str("worker_id", hostname).Str("coordinator_response", resp.Message).Msg("Left coordinator")
	return nil
}

func (c *coordinatorConnection) SetFlowManager(flowManager any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sm, ok := flowManager.(flowManagerInterface); ok {
		c.flowManager = sm
	}
}

func (c *coordinatorConnection) SendHeartbeat(ctx context.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	c.mu.Lock()
	if !c.joined {
		log.Debug().Str("worker_id", hostname).Msg("Worker not joined, skipping heartbeat")
		c.mu.Unlock()
		return nil
	}

	var runningFlowIDs []int64
	if c.flowManager != nil {
		runningFlowIDs = c.flowManager.GetRunningFlowIDs()
	}
	c.mu.Unlock()

	resp, err := c.coordinatorClient.Heartbeat(ctx, &pb.HeartbeatRequest{
		Id:                     hostname,
		Port:                   c.grpcPort,
		RunningWorkerFlowIds: runningFlowIDs,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send heartbeat to coordinator")
		c.mu.Lock()
		c.joined = false
		c.mu.Unlock()
		return err
	}

	for _, workerFlowID := range resp.ExpiredLeaseWorkerFlowIds {
		log.Warn().Int64("worker_flow_id", workerFlowID).Msg("Lease expired - stopping stream")
		if c.flowManager != nil {
			if err := c.flowManager.StopFlow(workerFlowID); err != nil {
				log.Error().Err(err).Int64("worker_flow_id", workerFlowID).Msg("Failed to stop expired flow")
			}
		}
	}

	log.Debug().
		Str("worker_id", hostname).
		Int("running_streams", len(runningFlowIDs)).
		Int("renewed", len(resp.RenewedLeaseWorkerFlowIds)).
		Int("expired", len(resp.ExpiredLeaseWorkerFlowIds)).
		Msg("Heartbeat sent")

	return nil
}

func (c *coordinatorConnection) GetClient() pb.CoordinatorClient {
	return c.coordinatorClient
}

func (c *coordinatorConnection) UpdateWorkerFlowStatus(ctx context.Context, workerFlowID int64, status pb.WorkerFlowStatus) error {
	resp, err := c.coordinatorClient.UpdateWorkerFlowStatus(
		ctx,
		&pb.WorkerFlowStatusRequest{
			WorkerFlowId: workerFlowID,
			Status:         status,
		},
	)
	if err != nil {
		return err
	}

	log.Info().
		Int64("worker_flow_id", workerFlowID).
		Str("coordinator_response", resp.Message).
		Msg("Updated worker flow status")

	return nil
}

func (c *coordinatorConnection) IngestMetrics(ctx context.Context, workerFlowID int64, inputEvents, processorErrors, outputEvents uint64) error {
	_, err := c.coordinatorClient.IngestMetrics(ctx, &pb.MetricsRequest{
		WorkerFlowId:  workerFlowID,
		InputEvents:     inputEvents,
		ProcessorErrors: processorErrors,
		OutputEvents:    outputEvents,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send metrics")
		return err
	}

	return nil
}
