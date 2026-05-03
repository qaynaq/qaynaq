package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/warpstreamlabs/bento/public/service"

	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"github.com/qaynaq/qaynaq/internal/vault"
)

type IngestResult struct {
	StatusCode int
	Response   []byte
}

type ServiceFlow struct {
	Stream         *service.Stream
	Cancel         context.CancelFunc
	Mux            http.Handler
	Status         persistence.WorkerFlowStatus
	TracingSummary *service.TracingSummary
}

type FlowManager interface {
	AddFlow(workerFlowID int64, config string) error
	WriteFiles(files []*pb.FlowFile) error
	GetFlow(workerFlowID int64) (*ServiceFlow, bool)
	GetFlowStatus(workerFlowID int64) (*persistence.WorkerFlowStatus, error)
	DeleteFlow(workerFlowID int64) error
	IngestData(workerFlowID int64, method, path, contentType string, payload []byte) (*IngestResult, error)
	GetAllFlows() map[int64]*ServiceFlow
	GetRunningFlowIDs() []int64
	StopFlow(workerFlowID int64) error
	StopAllFlows()
	StartFlow(ctx context.Context, workerFlowID int64)
}

type flowManager struct {
	mu                    sync.RWMutex
	flows                 map[int64]*ServiceFlow
	coordinatorConnection CoordinatorConnection
	vaultProvider         vault.VaultProvider
}

func NewFlowManager(coordinatorConnection CoordinatorConnection, vaultProvider vault.VaultProvider) FlowManager {
	return &flowManager{
		flows:                 make(map[int64]*ServiceFlow),
		coordinatorConnection: coordinatorConnection,
		vaultProvider:         vaultProvider,
	}
}

func (m *flowManager) WriteFiles(files []*pb.FlowFile) error {
	for _, f := range files {
		dest := filepath.Join("/tmp/qaynaq/files", f.Key)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", f.Key, err)
		}
		if err := os.WriteFile(dest, f.Content, 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", f.Key, err)
		}
		log.Debug().Str("key", f.Key).Str("path", dest).Msg("Wrote file to disk")
	}
	return nil
}

func (m *flowManager) AddFlow(workerFlowID int64, config string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.flows[workerFlowID]; exists {
		log.Debug().Int64("worker_flow_id", workerFlowID).Msg("Flow already exists")
		return nil
	}

	streamBuilder := service.NewStreamBuilder()
	streamMux := newSafeMux()
	streamBuilder.SetHTTPMux(streamMux)

	slogLogger := logger.NewSlogLogger("INFO", map[string]any{
		"@service":       "qaynaq",
		"worker_flow_id": workerFlowID,
	})
	streamBuilder.SetLogger(slogLogger)

	streamBuilder.SetEnvVarLookupFunc(func(key string) (string, bool) {
		// QAYNAQ_CONN_<name> resolves to the connection name itself. Components
		// receive the name and call vault.GetAccessToken(name) at request time so
		// the access token is always fresh and refresh-token rotations stay
		// consistent across the system.
		if connName, ok := strings.CutPrefix(key, "QAYNAQ_CONN_"); ok {
			return connName, true
		}
		secret, err := m.vaultProvider.GetSecret(key)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get secret")
			return "", false
		}
		return secret, true
	})

	if err := streamBuilder.SetYAML(config); err != nil {
		log.Error().Err(err).Msg("Failed to set flow YAML")
		return err
	}

	flow, tracingSummary, err := streamBuilder.BuildTracedV2()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build flow")
		return err
	}

	serviceStream := &ServiceFlow{
		Stream:         flow,
		Mux:            streamMux,
		Status:         persistence.WorkerFlowStatusRunning,
		TracingSummary: tracingSummary,
	}

	m.flows[workerFlowID] = serviceStream
	log.Info().Int64("worker_flow_id", workerFlowID).Msg("Flow added to manager")

	return nil
}

func (m *flowManager) GetFlow(workerFlowID int64) (*ServiceFlow, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flow, exists := m.flows[workerFlowID]
	return flow, exists
}

func (m *flowManager) GetFlowStatus(workerFlowID int64) (*persistence.WorkerFlowStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flow, exists := m.flows[workerFlowID]
	if !exists {
		return nil, nil
	}

	return &flow.Status, nil
}

func (m *flowManager) DeleteFlow(workerFlowID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flow, exists := m.flows[workerFlowID]
	if !exists {
		return fmt.Errorf("flow with ID %d not found", workerFlowID)
	}

	if flow.Stream != nil {
		if err := flow.Stream.StopWithin(5 * time.Second); err != nil {
			log.Warn().Err(err).Int64("worker_flow_id", workerFlowID).Msg("Failed to stop flow gracefully, forcing shutdown")
		}
	}

	if flow.Cancel != nil {
		flow.Cancel()
	}

	delete(m.flows, workerFlowID)
	log.Info().Int64("worker_flow_id", workerFlowID).Msg("Flow deleted from manager")

	return nil
}

func (m *flowManager) IngestData(workerFlowID int64, method, path, contentType string, payload []byte) (*IngestResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := &IngestResult{}

	flow, exists := m.flows[workerFlowID]
	if !exists {
		log.Debug().Int64("worker_flow_id", workerFlowID).Msg("flow is not running on this worker")
		return nil, fmt.Errorf("flow is assigned but not yet running on this worker (worker_flow_id: %d)", workerFlowID)
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(method, path, bytes.NewBuffer(payload))
	if err != nil {
		log.Error().
			Err(err).
			Int64("worker_flow_id", workerFlowID).
			Str("content_type", contentType).
			Str("method", method).
			Str("path", path).
			Bytes("payload", payload).
			Msg("Failed to create new request")
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	flow.Mux.ServeHTTP(rr, req)
	result.StatusCode = rr.Code

	if rr.Body != nil {
		result.Response, err = io.ReadAll(rr.Body)
		if err != nil {
			log.Error().
				Err(err).
				Int64("worker_flow_id", workerFlowID).
				Str("content_type", contentType).
				Str("method", method).
				Str("path", path).
				Bytes("payload", payload).
				Msg("Failed to read response body")
			result.Response = []byte("Failed to read response body")
		}
	}

	return result, nil
}

func (m *flowManager) GetAllFlows() map[int64]*ServiceFlow {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flows := make(map[int64]*ServiceFlow)
	maps.Copy(flows, m.flows)

	return flows
}

func (m *flowManager) GetRunningFlowIDs() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flowIDs := make([]int64, 0, len(m.flows))
	for workerFlowID := range m.flows {
		flowIDs = append(flowIDs, workerFlowID)
	}

	return flowIDs
}

func (m *flowManager) StopFlow(workerFlowID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flow, exists := m.flows[workerFlowID]
	if !exists {
		log.Debug().Int64("worker_flow_id", workerFlowID).Msg("Flow not found, already stopped")
		return nil
	}

	log.Info().Int64("worker_flow_id", workerFlowID).Msg("Stopping flow")

	if flow.Cancel != nil {
		flow.Cancel()
	}

	if flow.Stream != nil {
		if err := flow.Stream.StopWithin(5 * time.Second); err != nil {
			log.Warn().Err(err).Int64("worker_flow_id", workerFlowID).Msg("Failed to stop flow gracefully, forcing shutdown")
		}
	}

	delete(m.flows, workerFlowID)
	log.Info().Int64("worker_flow_id", workerFlowID).Msg("Stream stopped")

	return nil
}

func (m *flowManager) StopAllFlows() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for workerFlowID, flow := range m.flows {
		log.Info().Int64("worker_flow_id", workerFlowID).Msg("Stopping flow")
		if flow.Cancel != nil {
			flow.Cancel()
		}
	}

	m.flows = make(map[int64]*ServiceFlow)
	log.Info().Msg("All flows stopped")
}

func (m *flowManager) StartFlow(ctx context.Context, workerFlowID int64) {
	flow, exists := m.GetFlow(workerFlowID)
	if !exists {
		log.Error().Int64("worker_flow_id", workerFlowID).Msg("Flow not found for starting")
		return
	}

	log.Info().Int64("worker_flow_id", workerFlowID).Msg("Starting flow")

	if err := m.coordinatorConnection.UpdateWorkerFlowStatus(ctx, workerFlowID, pb.WorkerFlowStatus(pb.WorkerFlowStatus_value[string(persistence.WorkerFlowStatusRunning)])); err != nil {
		log.Warn().
			Err(err).
			Int64("worker_flow_id", workerFlowID).
			Str("status", string(persistence.WorkerFlowStatusRunning)).
			Msg("Failed to update worker flow status")
	}

	go func() {
		var flowStatus persistence.WorkerFlowStatus

		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Err(fmt.Errorf("%v", r)).
					Int64("worker_flow_id", workerFlowID).
					Msg("Flow panicked")
			}

			log.Info().Int64("worker_flow_id", workerFlowID).Str("status", string(flowStatus)).Msg("Finishing flow")

			m.shipMetrics(ctx, workerFlowID, flow.TracingSummary)

			if err := m.DeleteFlow(workerFlowID); err != nil {
				log.Debug().Err(err).Int64("worker_flow_id", workerFlowID).Msg("Flow already deleted")
			}

			if err := m.coordinatorConnection.UpdateWorkerFlowStatus(ctx, workerFlowID, pb.WorkerFlowStatus(pb.WorkerFlowStatus_value[string(flowStatus)])); err != nil {
				log.Warn().
					Err(err).
					Int64("worker_flow_id", workerFlowID).
					Str("status", string(flowStatus)).
					Msg("Failed to update worker flow status")
			}

			log.Info().Int64("worker_flow_id", workerFlowID).Str("status", string(flowStatus)).Msg("Flow has been finished")
		}()

		streamCtx, cancel := context.WithCancel(ctx)
		flow.Cancel = cancel

		m.mu.Lock()
		flow.Status = persistence.WorkerFlowStatusRunning
		m.mu.Unlock()

		log.Info().Int64("worker_flow_id", workerFlowID).Msg("Flow started running")

		if err := flow.Stream.Run(streamCtx); err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				log.Info().Int64("worker_flow_id", workerFlowID).Msg("Flow stopped: context canceled")
				flowStatus = persistence.WorkerFlowStatusStopped
			case errors.Is(err, context.DeadlineExceeded):
				log.Info().Int64("worker_flow_id", workerFlowID).Msg("Flow stopped: timeout exceeded")
				flowStatus = persistence.WorkerFlowStatusStopped
			default:
				log.Error().Err(err).Msg("Failed to run flow")
				flowStatus = persistence.WorkerFlowStatusFailed
			}
		} else {
			log.Info().Int64("worker_flow_id", workerFlowID).Msg("Flow has been completed")
			flowStatus = persistence.WorkerFlowStatusCompleted
		}
	}()
}

func (m *flowManager) shipMetrics(ctx context.Context, workerFlowID int64, tracingSummary *service.TracingSummary) {
	if tracingSummary == nil {
		return
	}

	err := m.coordinatorConnection.IngestMetrics(
		ctx,
		workerFlowID,
		tracingSummary.TotalInput(),
		tracingSummary.TotalProcessorErrors(),
		tracingSummary.TotalOutput(),
	)
	if err != nil {
		log.Error().Err(err).Int64("worker_flow_id", workerFlowID).Msg("Failed to ship metrics")
	}
}
