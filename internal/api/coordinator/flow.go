package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"

	coordinatorexecutor "github.com/qaynaq/qaynaq/internal/executor/coordinator"
	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func extractCacheResourceName(configYAML string) (string, error) {
	if configYAML == "" {
		return "", nil
	}

	var config map[string]any
	if err := yaml.Unmarshal([]byte(configYAML), &config); err != nil {
		return "", err
	}

	if cacheResource, ok := config["cache_resource"].(string); ok && cacheResource != "" {
		return cacheResource, nil
	}

	return "", nil
}

func extractRateLimitResourceName(configYAML string) (string, error) {
	if configYAML == "" {
		return "", nil
	}

	var config map[string]any
	if err := yaml.Unmarshal([]byte(configYAML), &config); err != nil {
		return "", err
	}

	if rateLimitResource, ok := config["rate_limit"].(string); ok && rateLimitResource != "" {
		return rateLimitResource, nil
	}

	return "", nil
}

func extractBufferResourceName(configYAML string) (string, error) {
	if configYAML == "" {
		return "", nil
	}

	var config map[string]any
	if err := yaml.Unmarshal([]byte(configYAML), &config); err != nil {
		return "", err
	}

	if bufferResource, ok := config["buffer"].(string); ok && bufferResource != "" {
		return bufferResource, nil
	}

	return "", nil
}

func (c *CoordinatorAPI) CreateFlow(_ context.Context, in *pb.Flow) (*pb.FlowResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	flow := &persistence.Flow{
		Processors: make([]persistence.FlowProcessor, len(in.Processors)),
	}
	flow.FromProto(in)

	// Validate buffer_id if provided
	if flow.BufferID != nil && *flow.BufferID != 0 {
		buffer, err := c.bufferRepo.FindByID(*flow.BufferID)
		if err != nil {
			log.Error().Err(err).Int64("buffer_id", *flow.BufferID).Msg("Failed to find buffer")
			return nil, status.Error(codes.Internal, err.Error())
		}
		if buffer == nil {
			return nil, status.Error(codes.InvalidArgument, "Buffer not found")
		}
	} else {
		flow.BufferID = nil
	}

	for i, processor := range in.Processors {
		flow.Processors[i] = persistence.FlowProcessor{
			Label:     processor.GetLabel(),
			Component: processor.GetComponent(),
			Config:    []byte(processor.GetConfig()),
		}
	}
	if !flow.IsReady {
		flow.Status = persistence.FlowStatusPaused
	} else if flow.Status == persistence.FlowStatusActive {
		if validationErr := validateFlowConfig(*flow); validationErr != nil {
			return nil, status.Error(codes.InvalidArgument,
				"Flow configuration is invalid. Set status to \"paused\" to save without running.\n\n"+validationErr.Error())
		}
	}

	if flow.ManagedBy != nil && *flow.ManagedBy != "" {
		existing, _ := c.flowRepo.FindByNameAndManagedBy(flow.Name, *flow.ManagedBy)
		if existing != nil {
			flow.ID = existing.ID
			if err := c.flowRepo.Update(flow); err != nil {
				log.Error().Err(err).Msg("Failed to update managed flow")
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			if err := c.flowRepo.Create(flow); err != nil {
				log.Error().Err(err).Msg("Failed to create flow")
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	} else {
		if err := c.flowRepo.Create(flow); err != nil {
			log.Error().Err(err).Msg("Failed to create flow")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Extract and store cache resources
	cacheNames := make(map[string]bool)

	// Check input config
	if cacheName, err := extractCacheResourceName(in.GetInputConfig()); err == nil && cacheName != "" {
		cacheNames[cacheName] = true
	}

	// Check output config
	if cacheName, err := extractCacheResourceName(in.GetOutputConfig()); err == nil && cacheName != "" {
		cacheNames[cacheName] = true
	}

	// Store cache references
	for cacheName := range cacheNames {
		cache, err := c.cacheRepo.FindByLabel(cacheName)
		if err != nil {
			log.Warn().Err(err).Str("cache_label", cacheName).Msg("Failed to find cache")
			continue
		}
		if cache == nil {
			log.Warn().Str("cache_label", cacheName).Msg("Cache not found")
			continue
		}

		flowCache := persistence.FlowCache{
			FlowID: flow.ID,
			CacheID:  cache.ID,
		}

		if err := c.flowCacheRepo.Create(flowCache); err != nil {
			log.Error().Err(err).Str("cache_label", cacheName).Msg("Failed to store flow cache")
		}
	}

	// Extract and store rate limit resources
	rateLimitNames := make(map[string]bool)

	// Check input config
	if rateLimitName, err := extractRateLimitResourceName(in.GetInputConfig()); err == nil && rateLimitName != "" {
		rateLimitNames[rateLimitName] = true
	}

	// Check output config
	if rateLimitName, err := extractRateLimitResourceName(in.GetOutputConfig()); err == nil && rateLimitName != "" {
		rateLimitNames[rateLimitName] = true
	}

	// Store rate limit references
	for rateLimitName := range rateLimitNames {
		rateLimit, err := c.rateLimitRepo.FindByLabel(rateLimitName)
		if err != nil {
			log.Warn().Err(err).Str("rate_limit_label", rateLimitName).Msg("Failed to find rate limit")
			continue
		}
		if rateLimit == nil {
			log.Warn().Str("rate_limit_label", rateLimitName).Msg("Rate limit not found")
			continue
		}

		flowRateLimit := persistence.FlowRateLimit{
			FlowID:    flow.ID,
			RateLimitID: rateLimit.ID,
		}

		if err := c.flowRateLimitRepo.Create(flowRateLimit); err != nil {
			log.Error().Err(err).Str("rate_limit_label", rateLimitName).Msg("Failed to store flow rate limit")
		}
	}

	// Extract and store buffer resources
	bufferNames := make(map[string]bool)

	// Check input config
	if bufferName, err := extractBufferResourceName(in.GetInputConfig()); err == nil && bufferName != "" {
		bufferNames[bufferName] = true
	}

	// Check output config
	if bufferName, err := extractBufferResourceName(in.GetOutputConfig()); err == nil && bufferName != "" {
		bufferNames[bufferName] = true
	}

	// Store buffer references
	for bufferName := range bufferNames {
		buffer, err := c.bufferRepo.FindByLabel(bufferName)
		if err != nil {
			log.Warn().Err(err).Str("buffer_label", bufferName).Msg("Failed to find buffer")
			continue
		}
		if buffer == nil {
			log.Warn().Str("buffer_label", bufferName).Msg("Buffer not found")
			continue
		}

		flowBuffer := persistence.FlowBuffer{
			FlowID: flow.ID,
			BufferID: buffer.ID,
		}

		if err := c.flowBufferRepo.Create(flowBuffer); err != nil {
			log.Error().Err(err).Str("buffer_label", bufferName).Msg("Failed to store flow buffer")
		}
	}

	return &pb.FlowResponse{
		Data: flow.ToProto(),
		Meta: &pb.CommonResponse{Message: "Flow has been created successfully"},
	}, nil
}

func (c *CoordinatorAPI) GetFlow(_ context.Context, in *pb.GetFlowRequest) (*pb.FlowResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	flow, err := c.flowRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find flow")
		return nil, status.Error(codes.Internal, err.Error())
	} else if flow == nil {
		return nil, status.Error(codes.NotFound, "Flow not found")
	}

	return &pb.FlowResponse{
		Data: flow.ToProto(),
		Meta: &pb.CommonResponse{Message: "OK"},
	}, nil
}

func (c *CoordinatorAPI) DeleteFlow(_ context.Context, in *pb.GetFlowRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	flow, err := c.flowRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find flow")
		return nil, status.Error(codes.Internal, err.Error())
	} else if flow == nil {
		return nil, status.Error(codes.NotFound, "Flow not found")
	}

	if err := c.flowProcessorRepo.DeleteByFlowID(flow.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete flow processors")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := c.flowCacheRepo.DeleteByFlowID(flow.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete flow cache references")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := c.flowRateLimitRepo.DeleteByFlowID(flow.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete flow rate limit references")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := c.flowBufferRepo.DeleteByFlowID(flow.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete flow buffer references")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := c.flowRepo.Delete(flow.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete flow")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{Message: "Flow has been deleted successfully"}, nil
}

func (c *CoordinatorAPI) ListFlows(_ context.Context, in *pb.ListFlowsRequest) (*pb.ListFlowsResponse, error) {
	flowStatuses := []persistence.FlowStatus{
		persistence.FlowStatusActive,
		persistence.FlowStatusCompleted,
		persistence.FlowStatusFailed,
		persistence.FlowStatusPaused,
	}

	if in.GetStatus() != "all" {
		flowStatuses = []persistence.FlowStatus{persistence.FlowStatus(in.GetStatus())}
	}

	flows, err := c.flowRepo.ListAllByStatuses(flowStatuses...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list flows")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListFlowsResponse{
		Data: make([]*pb.Flow, len(flows)),
	}
	for i, flow := range flows {
		result.Data[i] = flow.ToProto()
	}

	return result, nil
}

func (c *CoordinatorAPI) UpdateFlow(_ context.Context, in *pb.Flow) (*pb.FlowResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if in.GetId() == 0 {
		log.Debug().Msg("Invalid request: ID is required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	flow, err := c.flowRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find flow")
		return nil, status.Error(codes.Internal, err.Error())
	} else if flow == nil {
		return nil, status.Error(codes.NotFound, "Flow not found")
	}

	newFlow := &persistence.Flow{
		Processors: make([]persistence.FlowProcessor, len(in.Processors)),
	}
	newFlow.FromProto(in)

	// Validate buffer_id if provided
	if newFlow.BufferID != nil && *newFlow.BufferID != 0 {
		buffer, err := c.bufferRepo.FindByID(*newFlow.BufferID)
		if err != nil {
			log.Error().Err(err).Int64("buffer_id", *newFlow.BufferID).Msg("Failed to find buffer")
			return nil, status.Error(codes.Internal, err.Error())
		}
		if buffer == nil {
			return nil, status.Error(codes.InvalidArgument, "Buffer not found")
		}
	} else {
		newFlow.BufferID = nil
	}

	for i, processor := range in.Processors {
		newFlow.Processors[i] = persistence.FlowProcessor{
			Label:     processor.GetLabel(),
			Component: processor.GetComponent(),
			Config:    []byte(processor.GetConfig()),
		}
	}
	newFlow.ParentID = flow.ParentID
	if newFlow.ManagedBy == nil {
		newFlow.ManagedBy = flow.ManagedBy
	}

	if !newFlow.IsReady {
		newFlow.Status = persistence.FlowStatusPaused
	} else if newFlow.Status == persistence.FlowStatusActive {
		if validationErr := validateFlowConfig(*newFlow); validationErr != nil {
			return nil, status.Error(codes.InvalidArgument,
				"Flow configuration is invalid. Set status to \"paused\" to save without running.\n\n"+validationErr.Error())
		}
	}

	if err = c.flowRepo.Update(newFlow); err != nil {
		log.Error().Err(err).Msg("Failed to update flow")
		return nil, status.Error(codes.Internal, err.Error())
	}

	workerFlows, err := c.workerFlowRepo.ListAllByFlowID(flow.ID)
	if err != nil {
		log.Error().Err(err).Int64("flow_id", flow.ID).Msg("Failed to list worker flows for stopping")
	} else {
		for _, ws := range workerFlows {
			if ws.Status == persistence.WorkerFlowStatusRunning || ws.Status == persistence.WorkerFlowStatusWaiting {
				if err := c.workerFlowRepo.UpdateStatus(ws.ID, persistence.WorkerFlowStatusStopped); err != nil {
					log.Error().Err(err).Int64("worker_flow_id", ws.ID).Msg("Failed to stop worker flow")
				} else {
					log.Info().Int64("flow_id", flow.ID).Int64("worker_flow_id", ws.ID).Str("worker_id", ws.WorkerID).Msg("Stopped worker flow due to flow update")
					c.flowWorkerMap.RemoveFlow(ws.FlowID)
					if ws.Flow.ParentID != nil {
						c.flowWorkerMap.RemoveFlow(*ws.Flow.ParentID)
					}
				}
			}
		}
	}

	// Extract and store cache resources
	cacheNames := make(map[string]bool)

	// Check input config
	if cacheName, err := extractCacheResourceName(in.GetInputConfig()); err == nil && cacheName != "" {
		cacheNames[cacheName] = true
	}

	// Check output config
	if cacheName, err := extractCacheResourceName(in.GetOutputConfig()); err == nil && cacheName != "" {
		cacheNames[cacheName] = true
	}

	// Store cache references
	for cacheName := range cacheNames {
		cache, err := c.cacheRepo.FindByLabel(cacheName)
		if err != nil {
			log.Warn().Err(err).Str("cache_label", cacheName).Msg("Failed to find cache")
			continue
		}
		if cache == nil {
			log.Warn().Str("cache_label", cacheName).Msg("Cache not found")
			continue
		}

		flowCache := persistence.FlowCache{
			FlowID: newFlow.ID,
			CacheID:  cache.ID,
		}

		if err := c.flowCacheRepo.Create(flowCache); err != nil {
			log.Error().Err(err).Str("cache_label", cacheName).Msg("Failed to store flow cache")
		}
	}

	// Delete existing rate limit associations
	if err := c.flowRateLimitRepo.DeleteByFlowID(newFlow.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete existing flow rate limits")
	}

	// Extract and store rate limit resources
	rateLimitNames := make(map[string]bool)

	// Check input config
	if rateLimitName, err := extractRateLimitResourceName(in.GetInputConfig()); err == nil && rateLimitName != "" {
		rateLimitNames[rateLimitName] = true
	}

	// Check output config
	if rateLimitName, err := extractRateLimitResourceName(in.GetOutputConfig()); err == nil && rateLimitName != "" {
		rateLimitNames[rateLimitName] = true
	}

	// Store rate limit references
	for rateLimitName := range rateLimitNames {
		rateLimit, err := c.rateLimitRepo.FindByLabel(rateLimitName)
		if err != nil {
			log.Warn().Err(err).Str("rate_limit_label", rateLimitName).Msg("Failed to find rate limit")
			continue
		}
		if rateLimit == nil {
			log.Warn().Str("rate_limit_label", rateLimitName).Msg("Rate limit not found")
			continue
		}

		flowRateLimit := persistence.FlowRateLimit{
			FlowID:    newFlow.ID,
			RateLimitID: rateLimit.ID,
		}

		if err := c.flowRateLimitRepo.Create(flowRateLimit); err != nil {
			log.Error().Err(err).Str("rate_limit_label", rateLimitName).Msg("Failed to store flow rate limit")
		}
	}

	return &pb.FlowResponse{
		Data: newFlow.ToProto(),
		Meta: &pb.CommonResponse{Message: "Flow has been updated successfully"},
	}, nil
}

func validateFlowConfig(flow persistence.Flow) error {
	return coordinatorexecutor.ValidateFlow(flow)
}
