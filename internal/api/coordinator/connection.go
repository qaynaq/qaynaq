package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/qaynaq/qaynaq/internal/connection"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) ListProviders(_ context.Context, _ *emptypb.Empty) (*pb.ListProvidersResponse, error) {
	providers := connection.GetProviders()

	result := &pb.ListProvidersResponse{}
	for id, scopes := range providers {
		p := &pb.Provider{Id: id}
		for _, s := range scopes {
			p.Scopes = append(p.Scopes, &pb.ProviderScope{
				Scope:       s.Scope,
				Label:       s.Label,
				Description: s.Description,
			})
		}
		result.Data = append(result.Data, p)
	}

	return result, nil
}

func (c *CoordinatorAPI) ListConnections(_ context.Context, _ *emptypb.Empty) (*pb.ListConnectionsResponse, error) {
	connections, err := c.connManager.ListConnections()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list connections")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListConnectionsResponse{
		Data: make([]*pb.ConnectionInfo, len(connections)),
	}

	for i, conn := range connections {
		result.Data[i] = &pb.ConnectionInfo{
			Name:             conn.Name,
			Provider:         conn.Provider,
			Scopes:           conn.Scopes,
			ClientId:         conn.ClientID,
			ClientSecretHint: conn.ClientSecretHint,
			CreatedAt:        timestamppb.New(conn.CreatedAt),
			UpdatedAt:        timestamppb.New(conn.UpdatedAt),
		}
	}

	return result, nil
}

func (c *CoordinatorAPI) DeleteConnection(_ context.Context, in *pb.ConnectionRequest) (*pb.CommonResponse, error) {
	if err := c.connManager.DeleteConnection(in.GetName()); err != nil {
		log.Error().Err(err).Msg("Failed to delete connection")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{Message: "Connection has been deleted successfully"}, nil
}

func (c *CoordinatorAPI) GetConnectionToken(_ context.Context, in *pb.ConnectionRequest) (*pb.ConnectionTokenResponse, error) {
	data, err := c.connManager.GetConnectionData(in.GetName())
	if err != nil {
		log.Error().Err(err).Str("name", in.GetName()).Msg("Failed to get connection token")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ConnectionTokenResponse{Data: data}, nil
}

func (c *CoordinatorAPI) GetAccessToken(ctx context.Context, in *pb.ConnectionRequest) (*pb.AccessTokenResponse, error) {
	tok, err := c.connManager.GetAccessToken(ctx, in.GetName())
	if err != nil {
		log.Error().Err(err).Str("name", in.GetName()).Msg("Failed to get access token")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AccessTokenResponse{
		AccessToken: tok.AccessToken,
		ExpiresAt:   timestamppb.New(tok.ExpiresAt),
	}, nil
}
