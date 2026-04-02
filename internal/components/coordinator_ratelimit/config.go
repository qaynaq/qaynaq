package coordinator_ratelimit

import "github.com/warpstreamlabs/bento/public/service"

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Summary("Coordinator-based distributed rate limiter").
		Description("A rate limiter that uses the coordinator for distributed rate limiting across all workers. This ensures rate limits are enforced globally across the entire worker pool.")
}
