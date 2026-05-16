package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/qaynaq/qaynaq/internal/api/coordinator"
	"github.com/qaynaq/qaynaq/internal/auth"
	"github.com/qaynaq/qaynaq/internal/connection"
	"github.com/qaynaq/qaynaq/internal/executor"
	mcppkg "github.com/qaynaq/qaynaq/internal/mcp"
	mcpoauth "github.com/qaynaq/qaynaq/internal/mcp/oauth"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"github.com/qaynaq/qaynaq/internal/web"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type MCPSyncer interface {
	SyncTools()
}

type CoordinatorCLI struct {
	api                *coordinator.CoordinatorAPI
	executor           executor.CoordinatorExecutor
	rateLimiterEngine  interface{ Cleanup(time.Duration) error }
	authManager        *auth.Manager
	oauthHandler       *connection.OAuthHandler
	connRefreshJob     *connection.RefreshJob
	mcpHandler         http.Handler
	mcpSyncer          MCPSyncer
	mcpOAuthServer     *mcpoauth.Server
	httpPort, grpcPort uint32
}

func NewCoordinatorCLI(api *coordinator.CoordinatorAPI, executor executor.CoordinatorExecutor, rateLimiterEngine interface{ Cleanup(time.Duration) error }, authManager *auth.Manager, oauthHandler *connection.OAuthHandler, connRefreshJob *connection.RefreshJob, mcpHandler interface {
	http.Handler
	MCPSyncer
}, mcpOAuthServer *mcpoauth.Server, httpPort, grpcPort uint32) *CoordinatorCLI {
	return &CoordinatorCLI{api, executor, rateLimiterEngine, authManager, oauthHandler, connRefreshJob, mcpHandler, mcpHandler, mcpOAuthServer, httpPort, grpcPort}
}

func (c *CoordinatorCLI) Run(ctx context.Context) {
	g, ctx := errgroup.WithContext(ctx)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping worker health check / flow assignment routine...")
				return ctx.Err()
			case <-ticker.C:
				err := c.executor.CheckWorkersAndAssignFlows(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to perform worker health check and assign flows")
				}
			}
		}
	})

	leaseTicker := time.NewTicker(5 * time.Second)
	defer leaseTicker.Stop()

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping flow lease expiration checker routine...")
				return ctx.Err()
			case <-leaseTicker.C:
				err := c.executor.CheckFlowLeases(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to check flow leases")
				}
			}
		}
	})

	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping worker heartbeat timeout checker routine...")
				return ctx.Err()
			case <-heartbeatTicker.C:
				err := c.executor.CheckWorkerHeartbeats(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to check worker heartbeats")
				}
			}
		}
	})

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping rate limit state cleanup routine...")
				return ctx.Err()
			case <-cleanupTicker.C:
				err := c.rateLimiterEngine.Cleanup(24 * time.Hour)
				if err != nil {
					log.Error().Err(err).Msg("Failed to cleanup old rate limit states")
				} else {
					log.Debug().Msg("Rate limit state cleanup completed")
				}
			}
		}
	})

	mcpSyncTicker := time.NewTicker(30 * time.Second)
	defer mcpSyncTicker.Stop()

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping MCP tool sync routine...")
				return ctx.Err()
			case <-mcpSyncTicker.C:
				c.mcpSyncer.SyncTools()
			}
		}
	})

	g.Go(func() error {
		c.connRefreshJob.Run(ctx)
		return ctx.Err()
	})

	tokenUsageTicker := time.NewTicker(5 * time.Minute)
	defer tokenUsageTicker.Stop()

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Flushing token usage before shutdown...")
				c.api.FlushTokenUsage()
				return ctx.Err()
			case <-tokenUsageTicker.C:
				c.api.FlushTokenUsage()
			}
		}
	})

	coordinatorServerAddress := fmt.Sprintf(":%d", c.grpcPort)
	lis, err := net.Listen("tcp", coordinatorServerAddress)
	if err != nil {
		log.Fatal().Err(err).Uint32("port", c.grpcPort).Msg("failed to listen GRPC port")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, c.api)

	g.Go(func() error {
		log.Info().Uint32("port", c.grpcPort).Msg("starting coordinator GRPC server")
		errCh := make(chan error, 1)
		go func() {
			errCh <- grpcServer.Serve(lis)
		}()

		select {
		case err := <-errCh:
			log.Error().Err(err).Msg("gRPC server failed")
			return err
		case <-ctx.Done():
			log.Info().Msg("Shutting down gRPC server...")
			grpcServer.GracefulStop()
			log.Info().Msg("gRPC server stopped gracefully")
			return ctx.Err()
		}
	})

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		}}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err = pb.RegisterCoordinatorHandlerFromEndpoint(context.Background(), mux, coordinatorServerAddress, opts); err != nil {
		log.Fatal().Err(err).Msg("failed to register coordinator handler endpoint")
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Accept-Encoding", "Authorization", "Content-Type", "Origin", "Mcp-Session-Id"},
		ExposedHeaders:   []string{"Content-Length", "Mcp-Session-Id"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	})

	uiFS := web.AssetsFS()
	mainMux := http.NewServeMux()

	c.authManager.SetupAuthRoutes(mainMux)
	mainMux.HandleFunc("/connections/oauth/authorize", c.oauthHandler.HandleAuthorize)
	mainMux.HandleFunc("/connections/oauth/callback", c.oauthHandler.HandleCallback)

	apiMux := http.NewServeMux()
	apiMux.Handle("/", mux)
	if c.mcpOAuthServer != nil {
		c.mcpOAuthServer.MountAPIRoutes(apiMux)
	}
	protectedAPI := c.authManager.Middleware(apiMux)
	mainMux.Handle("/api/", http.StripPrefix("/api", protectedAPI))
	mainMux.HandleFunc("/ingest/", func(w http.ResponseWriter, r *http.Request) {
		statusCode, response, err := c.executor.ForwardRequestToWorker(r.Context(), r)
		if err != nil {
			log.Error().Err(err).Msg("failed to ingest flow")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(int(statusCode))
		_, _ = w.Write(response) //nolint:gosec // forwarding worker response payload to caller
	})
	mainMux.HandleFunc("/.well-known/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	if c.mcpOAuthServer != nil {
		c.mcpOAuthServer.MountRoutes(mainMux)
	}
	var mcpOAuthValidator mcppkg.OAuthValidator
	if c.mcpOAuthServer != nil {
		mcpOAuthValidator = c.mcpOAuthServer
	}
	mcpWithAuth := mcppkg.AuthMiddleware(c.api, mcpOAuthValidator, c.mcpHandler)
	mainMux.Handle("/mcp", mcpWithAuth)
	mainMux.Handle("/mcp/", mcpWithAuth)
	mainMux.HandleFunc("/", web.ServeSPA(uiFS, "index.html"))

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", c.httpPort),
		Handler:           corsMiddleware.Handler(mainMux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		log.Info().Uint32("port", c.httpPort).Msg("API gateway server starting")
		errCh := make(chan error, 1)
		go func() {
			errCh <- httpServer.ListenAndServe()
		}()

		select {
		case err := <-errCh:
			if err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("HTTP gateway server failed")
				return err
			}
			log.Info().Msg("HTTP gateway server stopped.")
			return nil
		case <-ctx.Done():
			log.Info().Msg("Shutting down HTTP gateway server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if c.mcpOAuthServer != nil {
				c.mcpOAuthServer.Shutdown(shutdownCtx)
			}
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				// Streaming MCP connections block graceful drain; force-close.
				log.Warn().Err(err).Msg("HTTP gateway graceful shutdown timed out, forcing close")
				if closeErr := httpServer.Close(); closeErr != nil {
					log.Error().Err(closeErr).Msg("HTTP gateway force close failed")
				}
			} else {
				log.Info().Msg("HTTP gateway server stopped gracefully")
			}
			return ctx.Err()
		}
	})

	log.Info().Msg("Coordinator running. Press Ctrl+C to stop.")
	if err := g.Wait(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		log.Error().Err(err).Msg("Coordinator encountered an error")
	} else {
		log.Info().Msg("Coordinator shutdown complete.")
	}
}
