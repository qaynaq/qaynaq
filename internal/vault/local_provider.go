package vault

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/config"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/grpc"
)

type LocalProvider struct {
	secretConfig          *config.SecretConfig
	coordinatorGRPCClient pb.CoordinatorClient
	aesgcm                *AESGCM
}

func NewLocalProvider(secretConfig *config.SecretConfig, grpcConn *grpc.ClientConn) VaultProvider {
	if secretConfig.Provider != config.SecretProviderLocal {
		log.Fatal().Msg("Invalid secret provider")
		return nil
	}
	aesgcm, err := NewAESGCM([]byte(secretConfig.Key))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create AESGCM")
		return nil
	}

	return &LocalProvider{
		secretConfig:          secretConfig,
		coordinatorGRPCClient: pb.NewCoordinatorClient(grpcConn),
		aesgcm:                aesgcm,
	}
}

func (p *LocalProvider) GetSecret(key string) (string, error) {
	secret, err := p.coordinatorGRPCClient.GetSecret(context.Background(), &pb.SecretRequest{Key: key})
	if err != nil {
		return "", err
	}

	decryptedValue, err := p.aesgcm.Decrypt(secret.Data.EncryptedValue)
	if err != nil {
		return "", err
	}

	return decryptedValue, nil
}

func (p *LocalProvider) GetConnectionToken(name string) (string, error) {
	resp, err := p.coordinatorGRPCClient.GetConnectionToken(context.Background(), &pb.ConnectionRequest{Name: name})
	if err != nil {
		return "", err
	}

	return resp.Data, nil
}
