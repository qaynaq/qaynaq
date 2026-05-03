package connection

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// RefreshJob proactively refreshes connection access tokens before they expire.
// Running this in coordinator means workers' cached tokens stay fresh without
// the workers ever needing to refresh themselves.
type RefreshJob struct {
	manager  *Manager
	interval time.Duration
	// proactiveWindow is how far in advance of expiry the job refreshes a token.
	// Should comfortably exceed `interval` so a token can't slip past expiry
	// between two job runs.
	proactiveWindow time.Duration
}

func NewRefreshJob(manager *Manager) *RefreshJob {
	return &RefreshJob{
		manager:         manager,
		interval:        5 * time.Minute,
		proactiveWindow: 15 * time.Minute,
	}
}

func (j *RefreshJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Connection refresh job stopped")
			return
		case <-ticker.C:
			j.runOnce(ctx)
		}
	}
}

func (j *RefreshJob) runOnce(ctx context.Context) {
	threshold := time.Now().Add(j.proactiveWindow)
	conns, err := j.manager.connRepo.ListExpiringBefore(threshold)
	if err != nil {
		log.Error().Err(err).Msg("Refresh job: failed to list expiring connections")
		return
	}

	if len(conns) == 0 {
		return
	}

	log.Debug().Int("count", len(conns)).Msg("Refresh job: refreshing connections")

	for _, conn := range conns {
		if ctx.Err() != nil {
			return
		}
		if err := j.manager.RefreshIfExpiring(ctx, conn.Name); err != nil {
			log.Warn().Err(err).Str("connection", conn.Name).Msg("Refresh job: failed to refresh connection")
		}
	}
}
