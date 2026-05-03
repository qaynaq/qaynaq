package vault

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/config"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/grpc"
)

// localCacheSafetyMargin: worker won't return a cached access token if it's
// closer than this to expiry. Slightly tighter than coordinator's margin so
// the worker calls coordinator before coordinator's own cache goes stale.
const localCacheSafetyMargin = 90 * time.Second

type LocalProvider struct {
	secretConfig          *config.SecretConfig
	coordinatorGRPCClient pb.CoordinatorClient
	aesgcm                *AESGCM

	cacheMu sync.RWMutex
	cache   map[string]AccessToken
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
		cache:                 make(map[string]AccessToken),
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

func (p *LocalProvider) GetAccessToken(name string) (AccessToken, error) {
	if v, ok := p.cacheLookup(name); ok {
		return v, nil
	}
	return p.fetchAccessToken(name, false)
}

func (p *LocalProvider) ForceRefreshAccessToken(name string) (AccessToken, error) {
	p.cacheMu.Lock()
	delete(p.cache, name)
	p.cacheMu.Unlock()
	return p.fetchAccessToken(name, true)
}

func (p *LocalProvider) fetchAccessToken(name string, forceRefresh bool) (AccessToken, error) {
	resp, err := p.coordinatorGRPCClient.GetAccessToken(context.Background(), &pb.AccessTokenRequest{
		Name:         name,
		ForceRefresh: forceRefresh,
	})
	if err != nil {
		return AccessToken{}, err
	}

	tok := AccessToken{AccessToken: resp.AccessToken}
	if resp.ExpiresAt != nil {
		tok.ExpiresAt = resp.ExpiresAt.AsTime()
	}

	p.cacheStore(name, tok)
	return tok, nil
}

func (p *LocalProvider) cacheLookup(name string) (AccessToken, bool) {
	p.cacheMu.RLock()
	defer p.cacheMu.RUnlock()
	v, ok := p.cache[name]
	if !ok {
		return AccessToken{}, false
	}
	if !v.ExpiresAt.IsZero() && time.Until(v.ExpiresAt) < localCacheSafetyMargin {
		return AccessToken{}, false
	}
	return v, true
}

func (p *LocalProvider) cacheStore(name string, tok AccessToken) {
	p.cacheMu.Lock()
	p.cache[name] = tok
	p.cacheMu.Unlock()
}
