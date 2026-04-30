package coordinator

import (
	"sync"
	"time"

	"github.com/qaynaq/qaynaq/internal/analytics"
	"github.com/qaynaq/qaynaq/internal/config"
	"github.com/qaynaq/qaynaq/internal/connection"
	pb "github.com/qaynaq/qaynaq/internal/protogen"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/ratelimiter"
	"github.com/qaynaq/qaynaq/internal/vault"
)

type settingsCache struct {
	mu           sync.RWMutex
	mcpProtected bool
	expiresAt    time.Time
}

type tokenUsageTracker struct {
	mu      sync.Mutex
	pending map[int64]time.Time
}

const settingsCacheTTL = 10 * time.Second

type FlowWorkerMap interface {
	SetFlowWorker(flowID int64, workerID string, workerFlowID int64)
	RemoveFlow(flowID int64)
	RemoveFlowIfMatches(flowID int64, workerFlowID int64)
}

type CoordinatorAPI struct {
	pb.UnimplementedCoordinatorServer
	eventRepo           persistence.EventRepository
	workerRepo          persistence.WorkerRepository
	flowRepo          persistence.FlowRepository
	flowCacheRepo     persistence.FlowCacheRepository
	flowRateLimitRepo persistence.FlowRateLimitRepository
	flowBufferRepo      persistence.FlowBufferRepository
	flowProcessorRepo persistence.FlowProcessorRepository
	workerFlowRepo    persistence.WorkerFlowRepository
	secretRepo          persistence.SecretRepository
	cacheRepo           persistence.CacheRepository
	bufferRepo          persistence.BufferRepository
	rateLimitRepo       persistence.RateLimitRepository
	fileRepo            persistence.FileRepository
	settingRepo         persistence.SettingRepository
	apiTokenRepo        persistence.APITokenRepository
	oauthClientRepo     persistence.OAuthClientRepository
	oauthRefreshRepo    persistence.OAuthRefreshTokenRepository
	oauthConsentRepo    persistence.OAuthConsentRepository
	rateLimiterEngine   *ratelimiter.Engine
	aesgcm              *vault.AESGCM
	analyticsProvider   analytics.Provider
	flowWorkerMap     FlowWorkerMap
	connManager       *connection.Manager
	authType          config.AuthType
	mcpOAuthEnabled   bool
	cache             settingsCache
	tokenUsage        tokenUsageTracker
}

func NewCoordinatorAPI(
	eventRepo persistence.EventRepository,
	flowRepo persistence.FlowRepository,
	flowCacheRepo persistence.FlowCacheRepository,
	flowRateLimitRepo persistence.FlowRateLimitRepository,
	flowBufferRepo persistence.FlowBufferRepository,
	flowProcessorRepo persistence.FlowProcessorRepository,
	workerRepo persistence.WorkerRepository,
	workerFlowRepo persistence.WorkerFlowRepository,
	secretRepo persistence.SecretRepository,
	cacheRepo persistence.CacheRepository,
	bufferRepo persistence.BufferRepository,
	rateLimitRepo persistence.RateLimitRepository,
	fileRepo persistence.FileRepository,
	settingRepo persistence.SettingRepository,
	apiTokenRepo persistence.APITokenRepository,
	oauthClientRepo persistence.OAuthClientRepository,
	oauthRefreshRepo persistence.OAuthRefreshTokenRepository,
	oauthConsentRepo persistence.OAuthConsentRepository,
	rateLimiterEngine *ratelimiter.Engine,
	aesgcm *vault.AESGCM,
	analyticsProvider analytics.Provider,
	connManager *connection.Manager,
	flowWorkerMap FlowWorkerMap,
	authType config.AuthType,
	mcpOAuthEnabled bool,
) *CoordinatorAPI {
	return &CoordinatorAPI{
		eventRepo:           eventRepo,
		flowRepo:          flowRepo,
		flowCacheRepo:     flowCacheRepo,
		flowRateLimitRepo: flowRateLimitRepo,
		flowBufferRepo:      flowBufferRepo,
		flowProcessorRepo: flowProcessorRepo,
		workerRepo:          workerRepo,
		workerFlowRepo:    workerFlowRepo,
		secretRepo:          secretRepo,
		cacheRepo:           cacheRepo,
		bufferRepo:          bufferRepo,
		rateLimitRepo:       rateLimitRepo,
		fileRepo:            fileRepo,
		settingRepo:         settingRepo,
		apiTokenRepo:        apiTokenRepo,
		oauthClientRepo:     oauthClientRepo,
		oauthRefreshRepo:    oauthRefreshRepo,
		oauthConsentRepo:    oauthConsentRepo,
		rateLimiterEngine:   rateLimiterEngine,
		aesgcm:              aesgcm,
		analyticsProvider:   analyticsProvider,
		connManager:       connManager,
		flowWorkerMap:     flowWorkerMap,
		authType:          authType,
		mcpOAuthEnabled:   mcpOAuthEnabled,
		tokenUsage:        tokenUsageTracker{pending: make(map[int64]time.Time)},
	}
}
