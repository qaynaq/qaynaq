package ratelimiter

import (
	"fmt"
	"math"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

type Engine struct {
	rateLimitRepo      persistence.RateLimitRepository
	rateLimitStateRepo persistence.RateLimitStateRepository
	mu                 sync.Mutex
}

func NewEngine(
	rateLimitRepo persistence.RateLimitRepository,
	rateLimitStateRepo persistence.RateLimitStateRepository,
) *Engine {
	return &Engine{
		rateLimitRepo:      rateLimitRepo,
		rateLimitStateRepo: rateLimitStateRepo,
	}
}

type CheckResult struct {
	Allowed      bool
	RetryAfterMs int64
	Remaining    int64
	Limit        int64
	ResetAt      int64
}

func (e *Engine) Check(label, key string, cost int64) (*CheckResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	rateLimit, err := e.rateLimitRepo.FindByLabel(label)
	if err != nil {
		return nil, fmt.Errorf("failed to find rate limit: %w", err)
	}
	if rateLimit == nil {
		return nil, fmt.Errorf("rate limit not found: %s", label)
	}

	var config map[string]any
	if err := yaml.Unmarshal(rateLimit.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rate limit config: %w", err)
	}

	count, ok := config["count"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid or missing count in rate limit config")
	}

	interval, ok := config["interval"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing interval in rate limit config")
	}

	burst := int64(0)
	if burstVal, ok := config["burst"].(int); ok {
		burst = int64(burstVal)
	}

	intervalDuration, err := parseDuration(interval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval: %w", err)
	}

	maxTokens := float64(count)
	if burst > 0 {
		maxTokens = float64(int64(count) + burst)
	}

	state, err := e.rateLimitStateRepo.GetOrCreate(label, key, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create state: %w", err)
	}

	now := time.Now()
	elapsed := now.Sub(state.LastRefillAt)

	tokensToAdd := (float64(count) / intervalDuration.Seconds()) * elapsed.Seconds()
	state.Tokens = math.Min(state.Tokens+tokensToAdd, maxTokens)
	state.LastRefillAt = now

	requestCost := float64(cost)
	if requestCost == 0 {
		requestCost = 1
	}

	allowed := state.Tokens >= requestCost

	var retryAfterMs int64
	if !allowed {
		tokensNeeded := requestCost - state.Tokens
		refillRate := float64(count) / intervalDuration.Seconds()
		secondsToWait := tokensNeeded / refillRate
		retryAfterMs = int64(math.Ceil(secondsToWait * 1000))
	} else {
		state.Tokens -= requestCost
	}

	if err := e.rateLimitStateRepo.Update(state); err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}

	resetAt := now.Add(intervalDuration).Unix()
	remaining := int64(math.Floor(state.Tokens))

	return &CheckResult{
		Allowed:      allowed,
		RetryAfterMs: retryAfterMs,
		Remaining:    remaining,
		Limit:        int64(maxTokens),
		ResetAt:      resetAt,
	}, nil
}

func (e *Engine) Cleanup(olderThan time.Duration) error {
	return e.rateLimitStateRepo.DeleteOlderThan(olderThan)
}

func parseDuration(interval string) (time.Duration, error) {
	if len(interval) < 2 {
		return 0, fmt.Errorf("invalid interval format: %s", interval)
	}

	unit := interval[len(interval)-1:]
	valueStr := interval[:len(interval)-1]

	var value int64
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return 0, fmt.Errorf("invalid interval value: %s", interval)
	}

	switch unit {
	case "s":
		return time.Duration(value) * time.Second, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported interval unit: %s", unit)
	}
}
