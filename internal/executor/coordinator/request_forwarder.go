package coordinator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

type RequestForwarder interface {
	ForwardRequestToWorker(ctx context.Context, r *http.Request) (int32, []byte, error)
}

type requestForwarder struct {
	workerManager   WorkerManager
	flowWorkerMap FlowWorkerMap
	flowRepo      persistence.FlowRepository
	pathRegex       *regexp.Regexp
}

func NewRequestForwarder(
	workerManager WorkerManager,
	flowWorkerMap FlowWorkerMap,
	flowRepo persistence.FlowRepository,
) RequestForwarder {
	return &requestForwarder{
		workerManager:   workerManager,
		flowWorkerMap: flowWorkerMap,
		flowRepo:      flowRepo,
		pathRegex:       regexp.MustCompile(`^/ingest/(\d+)(/.*)?$`),
	}
}

func (f *requestForwarder) ForwardRequestToWorker(ctx context.Context, r *http.Request) (int32, []byte, error) {
	var err error
	var id int64
	var componentPath string

	matches := f.pathRegex.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		return 0, nil, fmt.Errorf("invalid path format")
	}

	id, err = strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid ID format")
	}

	if len(matches) > 2 && matches[2] != "" {
		componentPath = matches[2]
	}

	workerFlowID, ok := f.flowWorkerMap.GetFlowWorkerStream(id)
	if !ok {
		flow, err := f.flowRepo.FindByID(id)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to look up flow %d: %w", id, err)
		}
		if flow == nil {
			return 0, nil, fmt.Errorf("flow %d not found", id)
		}
		return 0, nil, fmt.Errorf("flow %d exists but is not currently assigned to any worker", id)
	}

	workerID, ok := f.flowWorkerMap.GetFlowWorker(id)
	if !ok {
		return 0, nil, fmt.Errorf("flow %d exists but is not currently assigned to any worker", id)
	}

	worker, err := f.workerManager.FindWorker(workerID)
	if err != nil || worker == nil {
		return 0, nil, fmt.Errorf("worker %s not found", workerID)
	}

	workerClient, err := f.workerManager.GetWorkerClient(worker)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get worker client: %w", err)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	resp, err := workerClient.Ingest(ctx, &pb.IngestRequest{
		WorkerFlowId: workerFlowID,
		Method:         r.Method,
		Path:           componentPath,
		ContentType:    r.Header.Get("Content-Type"),
		Payload:        bodyBytes,
	})

	if err != nil {
		return 0, nil, fmt.Errorf("failed to forward request to worker: %w", err)
	}

	return resp.StatusCode, resp.Response, nil
}
