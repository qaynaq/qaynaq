package coordinator

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

const (
	settingMCPProtected  = "mcp_protected"
	settingSetupComplete = "setup_complete"
)

func (c *CoordinatorAPI) GetSetupStatus(_ context.Context, _ *emptypb.Empty) (*pb.SetupStatusResponse, error) {
	val, err := c.settingRepo.Get(settingSetupComplete)
	return &pb.SetupStatusResponse{FirstRunComplete: err == nil && val == "true"}, nil
}

func (c *CoordinatorAPI) CompleteSetup(_ context.Context, _ *emptypb.Empty) (*pb.CommonResponse, error) {
	if err := c.settingRepo.Set(settingSetupComplete, "true"); err != nil {
		log.Error().Err(err).Msg("Failed to complete setup")
		return nil, status.Error(codes.Internal, "failed to complete setup")
	}
	return &pb.CommonResponse{Message: "Setup completed"}, nil
}

func (c *CoordinatorAPI) GetMCPSettings(_ context.Context, _ *emptypb.Empty) (*pb.GetMCPSettingsResponse, error) {
	protected := false
	val, err := c.settingRepo.Get(settingMCPProtected)
	if err == nil {
		protected = val == "true"
	}

	tokens, err := c.apiTokenRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list API tokens")
		return nil, status.Error(codes.Internal, "failed to list tokens")
	}

	return &pb.GetMCPSettingsResponse{
		Protected:    protected,
		AuthEnabled:  c.authType != config.AuthTypeNone,
		Tokens:       c.toProtoTokens(tokens),
		OauthEnabled: c.mcpOAuthEnabled,
	}, nil
}

func (c *CoordinatorAPI) UpdateMCPProtected(_ context.Context, in *pb.UpdateMCPProtectedRequest) (*pb.UpdateMCPProtectedResponse, error) {
	val := "false"
	if in.GetProtected() {
		val = "true"
	}

	if err := c.settingRepo.Set(settingMCPProtected, val); err != nil {
		log.Error().Err(err).Msg("Failed to update MCP protected setting")
		return nil, status.Error(codes.Internal, "failed to update setting")
	}

	c.cache.mu.Lock()
	c.cache.mcpProtected = in.GetProtected()
	c.cache.expiresAt = time.Now().Add(settingsCacheTTL)
	c.cache.mu.Unlock()

	return &pb.UpdateMCPProtectedResponse{Protected: in.GetProtected()}, nil
}

func (c *CoordinatorAPI) ListAPITokens(_ context.Context, _ *emptypb.Empty) (*pb.ListAPITokensResponse, error) {
	tokens, err := c.apiTokenRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list API tokens")
		return nil, status.Error(codes.Internal, "failed to list tokens")
	}

	return &pb.ListAPITokensResponse{Tokens: c.toProtoTokens(tokens)}, nil
}

func (c *CoordinatorAPI) CreateAPIToken(_ context.Context, in *pb.CreateAPITokenRequest) (*pb.CreateAPITokenResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rawToken, err := generateToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	hash := hashToken(rawToken)

	token := &persistence.APIToken{
		Name:      strings.TrimSpace(in.GetName()),
		TokenHash: hash,
		Scopes:    in.GetScopes(),
		CreatedAt: time.Now(),
	}

	if err := c.apiTokenRepo.Create(token); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
			return nil, status.Error(codes.AlreadyExists, "a token with this name already exists")
		}
		log.Error().Err(err).Msg("Failed to create API token")
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	return &pb.CreateAPITokenResponse{
		Data: &pb.APIToken{
			Id:        token.ID,
			Name:      token.Name,
			Token:     rawToken,
			Scopes:    token.Scopes,
			CreatedAt: timestamppb.New(token.CreatedAt),
		},
	}, nil
}

func (c *CoordinatorAPI) DeleteAPIToken(_ context.Context, in *pb.DeleteAPITokenRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := c.apiTokenRepo.Delete(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete API token")
		return nil, status.Error(codes.Internal, "failed to delete token")
	}

	return &pb.CommonResponse{Message: "Token has been deleted successfully"}, nil
}

func (c *CoordinatorAPI) IsMCPProtected() bool {
	c.cache.mu.RLock()
	if time.Now().Before(c.cache.expiresAt) {
		val := c.cache.mcpProtected
		c.cache.mu.RUnlock()
		return val
	}
	c.cache.mu.RUnlock()

	val, err := c.settingRepo.Get(settingMCPProtected)
	protected := err == nil && val == "true"

	c.cache.mu.Lock()
	c.cache.mcpProtected = protected
	c.cache.expiresAt = time.Now().Add(settingsCacheTTL)
	c.cache.mu.Unlock()

	return protected
}

func (c *CoordinatorAPI) ValidateMCPToken(rawToken string) bool {
	hash := hashToken(rawToken)
	token, err := c.apiTokenRepo.FindByHash(hash)
	if err != nil {
		return false
	}
	if !hasScope(token.Scopes, "mcp") {
		return false
	}

	now := time.Now()
	c.tokenUsage.mu.Lock()
	c.tokenUsage.pending[token.ID] = now
	c.tokenUsage.mu.Unlock()

	return true
}

func (c *CoordinatorAPI) FlushTokenUsage() {
	c.tokenUsage.mu.Lock()
	if len(c.tokenUsage.pending) == 0 {
		c.tokenUsage.mu.Unlock()
		return
	}
	pending := c.tokenUsage.pending
	c.tokenUsage.pending = make(map[int64]time.Time)
	c.tokenUsage.mu.Unlock()

	if err := c.apiTokenRepo.BatchUpdateLastUsedAt(pending); err != nil {
		log.Error().Err(err).Msg("Failed to flush token usage data")
		c.tokenUsage.mu.Lock()
		for id, usedAt := range pending {
			if _, exists := c.tokenUsage.pending[id]; !exists {
				c.tokenUsage.pending[id] = usedAt
			}
		}
		c.tokenUsage.mu.Unlock()
	} else {
		log.Debug().Int("count", len(pending)).Msg("Flushed token usage data")
	}
}

func (c *CoordinatorAPI) ListOAuthClients(_ context.Context, _ *emptypb.Empty) (*pb.ListOAuthClientsResponse, error) {
	if !c.mcpOAuthEnabled {
		return &pb.ListOAuthClientsResponse{OauthEnabled: false}, nil
	}
	clients, err := c.oauthClientRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list oauth clients")
		return nil, status.Error(codes.Internal, "failed to list oauth clients")
	}
	consented, err := c.oauthConsentRepo.ClientIDsWithAnyConsent()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load consent state")
		return nil, status.Error(codes.Internal, "failed to load consent state")
	}
	out := make([]*pb.OAuthClient, len(clients))
	for i, cl := range clients {
		oc := &pb.OAuthClient{
			Id:           cl.ID,
			Name:         cl.Name,
			RedirectUris: cl.RedirectURIs,
			CreatedAt:    timestamppb.New(cl.CreatedAt),
			Consented:    consented[cl.ID],
		}
		if cl.LastUsedAt != nil {
			oc.LastUsedAt = timestamppb.New(*cl.LastUsedAt)
		}
		out[i] = oc
	}
	return &pb.ListOAuthClientsResponse{Clients: out, OauthEnabled: true}, nil
}

// RevokeOAuthConsent removes the user's consent for a client and revokes the
// client's refresh tokens. The client registration row is kept; the next
// connection from this MCP client will prompt for consent again.
func (c *CoordinatorAPI) RevokeOAuthConsent(_ context.Context, in *pb.RevokeOAuthConsentRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !c.mcpOAuthEnabled {
		return nil, status.Error(codes.FailedPrecondition, "MCP OAuth is not enabled")
	}
	if err := c.oauthConsentRepo.DeleteByClient(in.GetClientId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete oauth consent")
		return nil, status.Error(codes.Internal, "failed to revoke consent")
	}
	if err := c.oauthRefreshRepo.DeleteByClient(in.GetClientId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete oauth refresh tokens")
		return nil, status.Error(codes.Internal, "failed to revoke client sessions")
	}
	return &pb.CommonResponse{Message: "Consent revoked. The MCP client will prompt for consent on next connection."}, nil
}

// DeleteOAuthClient hard-deletes a client and all of its refresh tokens.
// Refresh tokens are dropped first so in-flight access tokens stop being
// accepted before the parent row is removed.
func (c *CoordinatorAPI) DeleteOAuthClient(_ context.Context, in *pb.DeleteOAuthClientRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !c.mcpOAuthEnabled {
		return nil, status.Error(codes.FailedPrecondition, "MCP OAuth is not enabled")
	}
	if err := c.oauthRefreshRepo.DeleteByClient(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete oauth refresh tokens for client")
		return nil, status.Error(codes.Internal, "failed to delete client sessions")
	}
	if err := c.oauthConsentRepo.DeleteByClient(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete oauth consents for client")
		return nil, status.Error(codes.Internal, "failed to delete client consents")
	}
	if err := c.oauthClientRepo.Delete(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete oauth client")
		return nil, status.Error(codes.Internal, "failed to delete client")
	}
	return &pb.CommonResponse{Message: "Client deleted. The MCP client will need to register again on next connect."}, nil
}

func (c *CoordinatorAPI) ListOAuthSessions(_ context.Context, _ *emptypb.Empty) (*pb.ListOAuthSessionsResponse, error) {
	if !c.mcpOAuthEnabled {
		return &pb.ListOAuthSessionsResponse{OauthEnabled: false}, nil
	}
	sessions, err := c.oauthRefreshRepo.ListActiveSessions()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list oauth sessions")
		return nil, status.Error(codes.Internal, "failed to list sessions")
	}
	out := make([]*pb.OAuthSession, len(sessions))
	for i, s := range sessions {
		out[i] = &pb.OAuthSession{
			Id:         s.ID,
			ClientId:   s.ClientID,
			ClientName: s.ClientName,
			UserEmail:  s.UserEmail,
			CreatedAt:  timestamppb.New(s.CreatedAt),
			ExpiresAt:  timestamppb.New(s.ExpiresAt),
		}
	}
	return &pb.ListOAuthSessionsResponse{Sessions: out, OauthEnabled: true}, nil
}

// RevokeOAuthSession invalidates a single refresh token. The client's
// access token keeps working until expiry; after that the SDK runs a fresh
// auth flow against the same client_id.
func (c *CoordinatorAPI) RevokeOAuthSession(_ context.Context, in *pb.RevokeOAuthSessionRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !c.mcpOAuthEnabled {
		return nil, status.Error(codes.FailedPrecondition, "MCP OAuth is not enabled")
	}
	if err := c.oauthRefreshRepo.Revoke(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to revoke oauth session")
		return nil, status.Error(codes.Internal, "failed to revoke session")
	}
	return &pb.CommonResponse{Message: "Session revoked. The user will be asked to log in again within an hour."}, nil
}

func (c *CoordinatorAPI) toProtoTokens(tokens []persistence.APIToken) []*pb.APIToken {
	pbTokens := make([]*pb.APIToken, len(tokens))
	for i, t := range tokens {
		pt := &pb.APIToken{
			Id:        t.ID,
			Name:      t.Name,
			Scopes:    t.Scopes,
			CreatedAt: timestamppb.New(t.CreatedAt),
		}
		if t.LastUsedAt != nil {
			pt.LastUsedAt = timestamppb.New(*t.LastUsedAt)
		}
		pbTokens[i] = pt
	}
	return pbTokens
}

func hasScope(scopes []string, target string) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, s := range scopes {
		if s == target || s == "*" {
			return true
		}
	}
	return false
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return "at_" + hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
