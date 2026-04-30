package main

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	_ "github.com/warpstreamlabs/bento/public/components/all"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/qaynaq/qaynaq/internal/analytics"
	"github.com/qaynaq/qaynaq/internal/sampledata"
	"github.com/qaynaq/qaynaq/internal/api"
	"github.com/qaynaq/qaynaq/internal/api/coordinator"
	"github.com/qaynaq/qaynaq/internal/auth"
	"github.com/qaynaq/qaynaq/internal/connection"
	intcli "github.com/qaynaq/qaynaq/internal/cli"
	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/qaynaq/qaynaq/internal/executor"
	executorcoordinator "github.com/qaynaq/qaynaq/internal/executor/coordinator"
	mcppkg "github.com/qaynaq/qaynaq/internal/mcp"
	mcpoauth "github.com/qaynaq/qaynaq/internal/mcp/oauth"
	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/ratelimiter"
	"github.com/qaynaq/qaynaq/internal/vault"
)

func expandStr(ctx *cli.Context, name string) string {
	return os.Expand(ctx.String(name), expandVar)
}

func expandVar(key string) string {
	if idx := strings.Index(key, ":-"); idx != -1 {
		if val, ok := os.LookupEnv(key[:idx]); ok {
			return val
		}
		return key[idx+2:]
	}
	return os.Getenv(key)
}

func buildDatabaseConfig(ctx *cli.Context) *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Driver: expandStr(ctx, "database.driver"),
		URI:    expandStr(ctx, "database.uri"),
	}
}

func buildSecretConfig(ctx *cli.Context) *config.SecretConfig {
	return &config.SecretConfig{
		Provider: config.SecretProviderLocal,
		Key:      expandStr(ctx, "secret.key"),
	}
}

func buildAuthConfig(ctx *cli.Context) *config.AuthConfig {
	return &config.AuthConfig{
		Type:                    config.AuthType(expandStr(ctx, "auth.type")),
		BasicUsername:           expandStr(ctx, "auth.basic-username"),
		BasicPassword:           expandStr(ctx, "auth.basic-password"),
		OAuth2ClientID:          expandStr(ctx, "auth.oauth2-client-id"),
		OAuth2ClientSecret:      expandStr(ctx, "auth.oauth2-client-secret"),
		OAuth2AuthorizationURL:  expandStr(ctx, "auth.oauth2-authorization-url"),
		OAuth2TokenURL:          expandStr(ctx, "auth.oauth2-token-url"),
		OAuth2RedirectURL:       expandStr(ctx, "auth.oauth2-redirect-url"),
		OAuth2Scopes:            splitComma(expandStr(ctx, "auth.oauth2-scopes")),
		OAuth2UserInfoURL:       expandStr(ctx, "auth.oauth2-user-info-url"),
		OAuth2AllowedUsers:      splitComma(expandStr(ctx, "auth.oauth2-allowed-users")),
		OAuth2AllowedDomains:    splitComma(expandStr(ctx, "auth.oauth2-allowed-domains")),
		OAuth2SessionCookieName: expandStr(ctx, "auth.oauth2-session-cookie-name"),
	}
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func InitializeCoordinatorCommand(ctx *cli.Context) *intcli.CoordinatorCLI {
	databaseConfig := buildDatabaseConfig(ctx)
	db := persistence.NewGormDB(databaseConfig)
	secretConfig := buildSecretConfig(ctx)
	authConfig := buildAuthConfig(ctx)
	authManager, err := auth.NewManager(authConfig, secretConfig.Key)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create auth manager")
		return nil
	}
	aesgcm, err := vault.NewAESGCM([]byte(secretConfig.Key))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create AESGCM")
		return nil
	}
	eventRepository := persistence.NewEventRepository(db)
	flowRepository := persistence.NewFlowRepository(db)
	workerRepository := persistence.NewWorkerRepository(db)
	workerFlowRepository := persistence.NewWorkerFlowRepository(db)
	secretRepository := persistence.NewSecretRepository(db)
	cacheRepository := persistence.NewCacheRepository(db)
	flowCacheRepository := persistence.NewFlowCacheRepository(db)
	bufferRepository := persistence.NewBufferRepository(db)
	flowBufferRepository := persistence.NewFlowBufferRepository(db)
	rateLimitRepository := persistence.NewRateLimitRepository(db)
	rateLimitStateRepository := persistence.NewRateLimitStateRepository(db)
	flowRateLimitRepository := persistence.NewFlowRateLimitRepository(db)
	flowProcessorRepository := persistence.NewFlowProcessorRepository(db)
	fileRepository := persistence.NewFileRepository(db)
	rateLimiterEngine := ratelimiter.NewEngine(rateLimitRepository, rateLimitStateRepository)
	analyticsProvider := analytics.NewLocalProvider(db)
	flowWorkerMap := executorcoordinator.NewFlowWorkerMap()
	settingRepository := persistence.NewSettingRepository(db)
	apiTokenRepository := persistence.NewAPITokenRepository(db)
	oauthClientRepository := persistence.NewOAuthClientRepository(db)
	oauthRefreshTokenRepository := persistence.NewOAuthRefreshTokenRepository(db)
	oauthConsentRepository := persistence.NewOAuthConsentRepository(db)
	connectionRepository := persistence.NewConnectionRepository(db)
	connManager := connection.NewManager(connectionRepository, aesgcm)
	mcpOAuthEnabled := ctx.Bool("mcp.oauth-enabled")
	coordinatorAPI := coordinator.NewCoordinatorAPI(eventRepository, flowRepository, flowCacheRepository, flowRateLimitRepository, flowBufferRepository, flowProcessorRepository, workerRepository, workerFlowRepository, secretRepository, cacheRepository, bufferRepository, rateLimitRepository, fileRepository, settingRepository, apiTokenRepository, oauthClientRepository, oauthRefreshTokenRepository, oauthConsentRepository, rateLimiterEngine, aesgcm, analyticsProvider, connManager, flowWorkerMap, authConfig.Type, mcpOAuthEnabled)
	coordinatorExecutor := executor.NewCoordinatorExecutor(workerRepository, flowRepository, flowCacheRepository, flowRateLimitRepository, workerFlowRepository, fileRepository, flowWorkerMap)
	oauthHandler := connection.NewOAuthHandler(connManager)
	mcpHandler := mcppkg.NewMCPHandler(flowRepository, coordinatorExecutor, Version)

	var mcpOAuthServer *mcpoauth.Server
	if mcpOAuthEnabled {
		sessionResolver := auth.NewSessionResolver(authManager)
		mcpOAuthServer = mcpoauth.NewServer(oauthClientRepository, oauthRefreshTokenRepository, oauthConsentRepository, sessionResolver, secretConfig.Key)
		log.Info().Msg("MCP OAuth Authorization Server enabled")
	}

	httpPort := uint32(ctx.Uint("http-port"))
	grpcPort := uint32(ctx.Uint("grpc-port"))
	sampledata.Init()
	coordinatorCLI := intcli.NewCoordinatorCLI(coordinatorAPI, coordinatorExecutor, rateLimiterEngine, authManager, oauthHandler, mcpHandler, mcpOAuthServer, httpPort, grpcPort)
	return coordinatorCLI
}

func InitializeWorkerCommand(appCtx context.Context, ctx *cli.Context) *intcli.WorkerCLI {
	secretConfig := buildSecretConfig(ctx)

	discoveryUri := ctx.String("discovery-uri")
	grpcConn, err := grpc.NewClient(discoveryUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create grpc client")
	}

	grpcPort := uint32(ctx.Uint("grpc-port"))
	vaultProvider := vault.NewLocalProvider(secretConfig, grpcConn)
	workerExecutor := executor.NewWorkerExecutor(appCtx, grpcConn, grpcPort, vaultProvider)
	workerAPI := api.NewWorkerAPI(workerExecutor)
	workerCLI := intcli.NewWorkerCLI(workerAPI, workerExecutor, grpcPort)
	return workerCLI
}
