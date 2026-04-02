package coordinator

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) CreateSecret(_ context.Context, in *pb.SecretRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if strings.HasPrefix(in.GetKey(), "QAYNAQ_CONN_") {
		return nil, status.Error(codes.InvalidArgument, "secret keys with the prefix 'QAYNAQ_CONN_' are reserved for connections")
	}

	encryptedValue, err := c.aesgcm.Encrypt(in.GetValue())
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt secret")
		return nil, status.Error(codes.Internal, err.Error())
	}

	secret := &persistence.Secret{
		Key:            in.GetKey(),
		EncryptedValue: encryptedValue,
	}

	keyExists, err := c.secretRepo.Create(secret)
	if keyExists && err != nil {
		return nil, status.Error(codes.AlreadyExists, "Secret with this key already exists")
	} else if err != nil {
		log.Error().Err(err).Msg("Failed to create secret")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{Message: "Secret has been created successfully"}, nil
}

func (c *CoordinatorAPI) ListSecrets(_ context.Context, _ *emptypb.Empty) (*pb.ListSecretsResponse, error) {
	secrets, err := c.secretRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list secrets")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListSecretsResponse{
		Data: make([]*pb.Secret, len(secrets)),
	}

	for i, secret := range secrets {
		result.Data[i] = &pb.Secret{
			Key:       secret.Key,
			CreatedAt: timestamppb.New(secret.CreatedAt),
		}
	}

	return result, nil
}

func (c *CoordinatorAPI) GetSecret(_ context.Context, in *pb.SecretRequest) (*pb.SecretResponse, error) {
	secret, err := c.secretRepo.GetByKey(in.GetKey())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get secret")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SecretResponse{
		Data: &pb.Secret{
			Key:            secret.Key,
			EncryptedValue: secret.EncryptedValue,
			CreatedAt:      timestamppb.New(secret.CreatedAt),
		},
	}, nil
}

func (c *CoordinatorAPI) DeleteSecret(_ context.Context, in *pb.SecretRequest) (*pb.CommonResponse, error) {
	if err := c.secretRepo.Delete(in.GetKey()); err != nil {
		log.Error().Err(err).Msg("Failed to delete secret")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{Message: "Secret has been deleted successfully"}, nil
}
