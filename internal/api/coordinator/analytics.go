package coordinator

import (
	"context"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CoordinatorAPI) GetAnalytics(_ context.Context, _ *pb.GetAnalyticsRequest) (*pb.GetAnalyticsResponse, error) {
	if s.analyticsProvider == nil {
		return nil, status.Error(codes.Unimplemented, "analytics provider not configured")
	}

	result, err := s.analyticsProvider.GetAnalytics()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get analytics: %v", err)
	}

	resp := &pb.GetAnalyticsResponse{
		TotalFlows:         result.TotalFlows,
		TotalInputEvents:    result.TotalInputEvents,
		TotalOutputEvents:   result.TotalOutputEvents,
		TotalProcessorErrors: result.TotalProcessorErrors,
		ActiveWorkers:       result.ActiveWorkers,
		TotalEvents:         result.TotalEvents,
		ErrorEvents:         result.ErrorEvents,
	}

	resp.FlowsByStatus = make([]*pb.GetAnalyticsResponse_FlowStatusCount, len(result.FlowsByStatus))
	for i, s := range result.FlowsByStatus {
		resp.FlowsByStatus[i] = &pb.GetAnalyticsResponse_FlowStatusCount{
			Status: s.Status,
			Count:  s.Count,
		}
	}

	resp.EventsOverTime = make([]*pb.GetAnalyticsResponse_TimeSeriesPoint, len(result.EventsOverTime))
	for i, pt := range result.EventsOverTime {
		resp.EventsOverTime[i] = &pb.GetAnalyticsResponse_TimeSeriesPoint{
			Timestamp:   pt.Timestamp.Format("2006-01-02"),
			InputEvents: pt.InputEvents,
			OutputEvents: pt.OutputEvents,
			ErrorEvents: pt.ErrorEvents,
		}
	}

	resp.TopInputComponents = make([]*pb.GetAnalyticsResponse_ComponentCount, len(result.TopInputComponents))
	for i, c := range result.TopInputComponents {
		resp.TopInputComponents[i] = &pb.GetAnalyticsResponse_ComponentCount{
			Component: c.Component,
			Count:     c.Count,
		}
	}

	resp.TopOutputComponents = make([]*pb.GetAnalyticsResponse_ComponentCount, len(result.TopOutputComponents))
	for i, c := range result.TopOutputComponents {
		resp.TopOutputComponents[i] = &pb.GetAnalyticsResponse_ComponentCount{
			Component: c.Component,
			Count:     c.Count,
		}
	}

	return resp, nil
}
