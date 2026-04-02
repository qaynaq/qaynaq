package cli

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/qaynaq/qaynaq/internal/api"
	"github.com/qaynaq/qaynaq/internal/executor"
	pb "github.com/qaynaq/qaynaq/internal/protogen"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type WorkerCLI struct {
	api      *api.WorkerAPI
	executor executor.WorkerExecutor
	grpcPort uint32
}

func NewWorkerCLI(api *api.WorkerAPI, executor executor.WorkerExecutor, grpcPort uint32) *WorkerCLI {
	return &WorkerCLI{api, executor, grpcPort}
}

func (c *WorkerCLI) Run(ctx context.Context) {
	g, ctx := errgroup.WithContext(ctx)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	workerServerAddress := fmt.Sprintf(":%d", c.grpcPort)
	lis, err := net.Listen("tcp", workerServerAddress)
	if err != nil {
		log.Fatal().Err(err).Uint32("port", c.grpcPort).Msg("failed to listen GRPC port")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWorkerServer(grpcServer, c.api)

	grpcReady := make(chan struct{})

	g.Go(func() error {
		log.Info().Uint32("port", c.grpcPort).Msg("starting worker GRPC server")
		errCh := make(chan error, 1)
		go func() {
			errCh <- grpcServer.Serve(lis)
		}()

		close(grpcReady)

		select {
		case err := <-errCh:
			log.Error().Err(err).Msg("Worker gRPC server failed")
			return err
		case <-ctx.Done():
			log.Info().Msg("Shutting down worker gRPC server...")
			grpcServer.GracefulStop()
			log.Info().Msg("Worker gRPC server stopped gracefully.")
			return ctx.Err()
		}
	})

	g.Go(func() error {
		<-grpcReady

		if err := c.executor.JoinToCoordinator(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to register with coordinator")
		}

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping worker heartbeat/metrics routine...")
				if err := c.executor.LeaveCoordinator(context.Background()); err != nil {
					log.Error().Err(err).Msg("Failed to leave coordinator during shutdown")
				}
				log.Info().Msg("Worker heartbeat/metrics routine stopped.")
				return ctx.Err()
			case <-ticker.C:
				if err := c.executor.SendHeartbeat(ctx); err != nil {
					log.Error().Err(err).Msg("Failed to send heartbeat to coordinator")
					if err := c.executor.JoinToCoordinator(ctx); err != nil {
						log.Error().Err(err).Msg("Failed to re-register with coordinator")
					}
				}
				c.executor.ShipMetrics(ctx)
			}
		}
	})

	g.Go(func() error {
		log.Info().Msg("Starting worker flow queue consumption...")
		c.executor.ConsumeFlowQueue(ctx)
		log.Info().Msg("Worker flow queue consumption stopped.")
		return ctx.Err()
	})

	g.Go(func() error {
		log.Info().Msg("Starting worker log shipping...")
		c.executor.ShipLogs(ctx)
		log.Info().Msg("Worker log shipping stopped.")
		return ctx.Err()
	})

	log.Info().Msg("Worker running. Press Ctrl+C to stop.")
	if err := g.Wait(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		log.Error().Err(err).Msg("Worker encountered an error during run")
	} else {
		log.Info().Msg("Worker shutdown complete.")
	}
}
