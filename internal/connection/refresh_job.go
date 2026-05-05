package connection

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// RefreshJob proactively rotates access tokens before they expire so
// workers' caches don't have to.
type RefreshJob struct {
	manager  *Manager
	interval time.Duration
	// Must exceed interval so no token can slip past expiry between ticks.
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
