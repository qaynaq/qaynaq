package coordinator

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) ListGroups(_ context.Context, _ *emptypb.Empty) (*pb.ListGroupsResponse, error) {
	groups, err := c.groupRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list groups")
		return nil, status.Error(codes.Internal, "failed to list groups")
	}

	data := make([]*pb.GroupInfo, len(groups))
	for i, g := range groups {
		data[i] = &pb.GroupInfo{
			Name:      g.Name,
			Source:    g.Source,
			CreatedAt: timestamppb.New(g.CreatedAt),
		}
		if g.LastSeenAt != nil {
			data[i].LastSeenAt = timestamppb.New(*g.LastSeenAt)
		}
	}
	return &pb.ListGroupsResponse{Data: data}, nil
}

func (c *CoordinatorAPI) ImportGroups(_ context.Context, req *pb.ImportGroupsRequest) (*pb.ImportGroupsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	seen := make(map[string]bool, len(req.Names))
	names := make([]string, 0, len(req.Names))
	for _, name := range req.Names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no group names provided")
	}

	imported, err := c.groupRepo.ImportManual(names)
	if err != nil {
		log.Error().Err(err).Msg("Failed to import groups")
		return nil, status.Error(codes.Internal, "failed to import groups")
	}
	return &pb.ImportGroupsResponse{Imported: int32(imported)}, nil
}

func (c *CoordinatorAPI) DeleteGroup(_ context.Context, req *pb.DeleteGroupRequest) (*pb.CommonResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := c.groupRepo.Delete(req.Name); err != nil {
		log.Error().Err(err).Str("name", req.Name).Msg("Failed to delete group")
		return nil, status.Error(codes.Internal, "failed to delete group")
	}
	return &pb.CommonResponse{Message: "Group has been deleted successfully"}, nil
}

func (c *CoordinatorAPI) ListMCPCallLogs(_ context.Context, req *pb.ListMCPCallLogsRequest) (*pb.ListMCPCallLogsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	entries, total, err := c.mcpCallLogRepo.List(limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list MCP call logs")
		return nil, status.Error(codes.Internal, "failed to list MCP call logs")
	}

	data := make([]*pb.MCPCallLogEntry, len(entries))
	for i, e := range entries {
		data[i] = &pb.MCPCallLogEntry{
			Id:         e.ID,
			Actor:      e.Actor,
			ActorKind:  e.ActorKind,
			Groups:     e.Groups,
			ToolName:   e.ToolName,
			TargetKind: e.TargetKind,
			TargetId:   e.TargetID,
			Status:     e.Status,
			Error:      e.Error,
			DurationMs: e.DurationMs,
			CreatedAt:  timestamppb.New(e.CreatedAt),
		}
	}
	return &pb.ListMCPCallLogsResponse{Data: data, Total: total}, nil
}
