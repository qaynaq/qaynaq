package coordinator

import (
	"context"
	"encoding/json"
	"io"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *CoordinatorAPI) IngestEvents(stream grpc.BidiStreamingServer[pb.Event, emptypb.Empty]) error {
	flowIDCache := make(map[int64]int64)

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive event: %v", err)
		}

		meta, err := event.GetMeta().MarshalJSON()
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid event format: %v", err)
		}

		wsID := event.GetWorkerFlowId()
		flowID, ok := flowIDCache[wsID]
		if !ok {
			ws, wsErr := c.workerFlowRepo.FindByID(wsID)
			if wsErr != nil {
				log.Error().Err(wsErr).Int64("worker_flow_id", wsID).Msg("failed to find worker flow")
				return status.Error(codes.Internal, "failed to resolve flow ID")
			}
			if ws == nil {
				return status.Errorf(codes.NotFound, "worker flow %d not found", wsID)
			}
			flowID = ws.FlowID
			flowIDCache[wsID] = flowID
		}

		evetEntity := &persistence.Event{
			WorkerFlowID: wsID,
			FlowID:       flowID,
			TraceID:      event.GetTraceId(),
			Section:        event.GetSection(),
			ComponentLabel: event.GetComponentLabel(),
			Type:           persistence.EventType(event.GetType()),
			Meta:           meta,
			Content:        event.GetContent(),
		}
		if err = c.eventRepo.AddEvent(evetEntity); err != nil {
			return status.Error(codes.InvalidArgument, "invalid event format")
		}

		if err = stream.Send(&emptypb.Empty{}); err != nil {
			log.Error().Err(err).Msg("failed to ack event")
			return status.Error(codes.Internal, "failed to ack event")
		}
	}
}

func (c *CoordinatorAPI) ListEvents(ctx context.Context, in *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	limit := int(in.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := int(in.GetOffset())

	startTime := in.GetStartTime().AsTime()
	endTime := in.GetEndTime().AsTime()

	// The request sends the parent (initial) flow ID.
	// Fetch all versions: the parent itself + all versions with that parent_id.
	parentID := in.GetFlowId()
	allFlows, err := c.flowRepo.ListAllVersionsByParentID(parentID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list flow versions")
		return nil, status.Error(codes.Internal, "failed to list flow versions")
	}

	flowIDs := make([]int64, 0, len(allFlows)+1)
	flowIDs = append(flowIDs, parentID)
	for _, s := range allFlows {
		if s.ID != parentID {
			flowIDs = append(flowIDs, s.ID)
		}
	}

	events, total, err := c.eventRepo.ListEventsByFlowIDs(flowIDs, limit, offset, startTime, endTime)
	if err != nil {
		log.Error().Err(err).Msg("failed to list events")
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	pbEvents := make([]*pb.Event, 0, len(events))
	for _, e := range events {
		var metaMap map[string]any
		if len(e.Meta) > 0 {
			_ = json.Unmarshal(e.Meta, &metaMap)
		}
		metaStruct, _ := structpb.NewStruct(metaMap)

		pbEvents = append(pbEvents, &pb.Event{
			Id:             e.ID,
			WorkerFlowId: e.WorkerFlowID,
			TraceId:         e.TraceID,
			Section:        e.Section,
			ComponentLabel: e.ComponentLabel,
			Type:           string(e.Type),
			Content:        e.Content,
			Meta:           metaStruct,
			CreatedAt:      timestamppb.New(e.CreatedAt),
		})
	}

	return &pb.ListEventsResponse{
		Data:  pbEvents,
		Total: total,
	}, nil
}

func (c *CoordinatorAPI) IngestMetrics(ctx context.Context, in *pb.MetricsRequest) (*emptypb.Empty, error) {
	var err error

	if err = in.Validate(); err != nil {
		log.Error().Err(err).Msg("failed to validate metrics request")
		return nil, status.Error(codes.InvalidArgument, "invalid metrics request")
	}

	err = c.workerFlowRepo.UpdateMetrics(
		in.GetWorkerFlowId(),
		in.GetInputEvents(),
		in.GetProcessorErrors(),
		in.GetOutputEvents(),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to update metrics")
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &emptypb.Empty{}, nil
}
