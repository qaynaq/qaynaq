package shopify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	goshopify "github.com/bold-commerce/go-shopify/v4"
	"github.com/warpstreamlabs/bento/public/service"
)

const rateLimitKeyContextKey = "rate_limit_key"

func init() {
	err := service.RegisterBatchInput(
		"shopify", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			input, err := NewFromConfig(conf, mgr)
			if err != nil {
				return nil, err
			}
			return service.AutoRetryNacksBatched(input), nil
		})
	if err != nil {
		panic(err)
	}
}

type Input struct {
	shopName       string
	apiKey         string
	apiAccessToken string
	shopResource   string
	limit          int
	apiVersion     string

	client    *goshopify.Client
	connected bool
	connMutex sync.RWMutex

	rateLimitLabel string

	finished     bool
	pendingBatch []any
	stateMutex   sync.Mutex

	cacheName     string
	cacheKey      string
	lastUpdatedAt *time.Time
	latestAckAt   *time.Time

	mgr    *service.Resources
	logger *service.Logger
}

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Input, error) {
	shopName, err := conf.FieldString(sbfShopName)
	if err != nil {
		return nil, err
	}

	apiKey, err := conf.FieldString(sbfAPIKey)
	if err != nil {
		return nil, err
	}

	apiAccessToken, err := conf.FieldString(sbfAPIAccessToken)
	if err != nil {
		return nil, err
	}

	shopResource := "products"
	if conf.Contains(sbfShopResource) {
		shopResource, err = conf.FieldString(sbfShopResource)
		if err != nil {
			return nil, err
		}
	}

	limit := 50
	if conf.Contains(sbfLimit) {
		limit, err = conf.FieldInt(sbfLimit)
		if err != nil {
			return nil, err
		}
	}

	if limit < 1 || limit > 250 {
		return nil, fmt.Errorf("limit must be between 1 and 250, got %d", limit)
	}

	var apiVersion string
	if conf.Contains(sbfAPIVersion) {
		apiVersion, err = conf.FieldString(sbfAPIVersion)
		if err != nil {
			return nil, err
		}
	}

	var cacheName string
	if conf.Contains(sbfCache) {
		cacheName, err = conf.FieldString(sbfCache)
		if err != nil {
			return nil, err
		}
	}

	var cacheKey string
	if cacheName != "" {
		cacheKey = fmt.Sprintf("shopify:%s:%s:last_updated_at", shopName, shopResource)
	}

	var rateLimitLabel string
	if conf.Contains(sbfRateLimit) {
		rateLimitLabel, err = conf.FieldString(sbfRateLimit)
		if err != nil {
			return nil, err
		}
	}

	return &Input{
		shopName:       shopName,
		apiKey:         apiKey,
		apiAccessToken: apiAccessToken,
		shopResource:   shopResource,
		limit:          limit,
		apiVersion:     apiVersion,
		rateLimitLabel: rateLimitLabel,
		finished:       false,
		cacheName:      cacheName,
		cacheKey:       cacheKey,
		lastUpdatedAt:  nil,
		latestAckAt:    nil,
		mgr:            mgr,
		logger:         mgr.Logger(),
	}, nil
}

func (s *Input) Connect(ctx context.Context) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.connected {
		return nil
	}

	app := goshopify.App{
		ApiKey:   s.apiKey,
		Password: s.apiAccessToken,
	}

	var opts []goshopify.Option
	if s.apiVersion != "" {
		opts = append(opts, goshopify.WithVersion(s.apiVersion))
	}
	opts = append(opts, goshopify.WithRetry(3))

	client, err := goshopify.NewClient(app, s.shopName, "", opts...)
	if err != nil {
		return fmt.Errorf("failed to create Shopify client: %w", err)
	}

	s.client = client
	s.connected = true
	s.logger.Infof("Connected to Shopify store: %s.myshopify.com", s.shopName)
	s.logger.Infof("Resource type: %s", s.shopResource)
	if s.rateLimitLabel != "" {
		s.logger.Infof("Using rate limit: %s with key: %s", s.rateLimitLabel, s.shopName)
	}

	if s.cacheName != "" {
		s.logger.Infof("Using cache '%s' for position tracking", s.cacheName)
		if err := s.loadPosition(ctx); err != nil {
			s.logger.Warnf("Failed to load position from cache: %v", err)
		}
	}

	return nil
}

func (s *Input) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	s.connMutex.RLock()
	connected := s.connected
	s.connMutex.RUnlock()

	if !connected {
		return nil, nil, service.ErrNotConnected
	}

	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	// If finished, no more data to fetch
	if s.finished {
		return nil, nil, service.ErrEndOfInput
	}

	var items []any

	if s.pendingBatch != nil {
		s.logger.Debugf("Returning pending batch with %d items", len(s.pendingBatch))
		items = s.pendingBatch
	} else {
		var err error
		items, err = s.fetchBatch(ctx)
		if err != nil {
			return nil, nil, err
		}
		s.pendingBatch = items
	}

	batch := make(service.MessageBatch, len(items))
	for i, item := range items {
		msg := service.NewMessage(nil)

		// Convert struct to map[string]any for field access on mapping
		itemMap, err := structToMap(item)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert item to map: %w", err)
		}

		msg.SetStructured(map[string]any{
			"resource": s.shopResource,
			"data":     itemMap,
		})
		batch[i] = msg
	}

	ackFunc := func(ctx context.Context, err error) error {
		if err != nil {
			s.logger.Errorf("Batch not acknowledged: %v", err)
			// Keep pending batch and return error for AutoRetryNacksBatched to retry with backoff
			return err
		}

		// Clear pending batch on successful acknowledgment
		s.stateMutex.Lock()
		s.pendingBatch = nil
		s.stateMutex.Unlock()

		if s.cacheName != "" && len(items) > 0 {
			lastItem := items[len(items)-1]
			if updatedAt := s.extractUpdatedAt(lastItem); updatedAt != nil {
				s.latestAckAt = updatedAt
				adjustedTime := updatedAt.Add(1 * time.Nanosecond)
				s.savePosition(context.Background(), adjustedTime)
			}
		}

		// Mark as finished only after successful acknowledgment
		if len(items) < s.limit {
			s.stateMutex.Lock()
			s.finished = true
			s.stateMutex.Unlock()
		}

		return nil
	}

	return batch, ackFunc, nil
}

func structToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Input) fetchBatch(ctx context.Context) ([]any, error) {
	s.logger.Debugf("Fetching next batch of %s from Shopify", s.shopResource)

	if s.rateLimitLabel != "" {
		if err := s.checkRateLimit(ctx); err != nil {
			return nil, fmt.Errorf("rate limit check failed: %w", err)
		}
	}

	var items []any
	var err error

	switch s.shopResource {
	case "products":
		items, err = s.fetchProductsPage(ctx)
	case "orders":
		items, err = s.fetchOrdersPage(ctx)
	case "customers":
		items, err = s.fetchCustomersPage(ctx)
	case "inventory_items":
		items, err = s.fetchInventoryItemsPage(ctx)
	case "locations":
		items, err = s.fetchLocationsPage(ctx)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", s.shopResource)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", s.shopResource, err)
	}

	s.logger.Infof("Fetched %d %s", len(items), s.shopResource)

	return items, nil
}

func (s *Input) checkRateLimit(ctx context.Context) error {
	var waitDuration time.Duration
	var accessErr error

	keyCtx := context.WithValue(ctx, rateLimitKeyContextKey, s.shopName)

	err := s.mgr.AccessRateLimit(ctx, s.rateLimitLabel, func(rl service.RateLimit) {
		waitDuration, accessErr = rl.Access(keyCtx)
	})

	if err != nil {
		return fmt.Errorf("failed to access rate limit: %w", err)
	}

	if accessErr != nil {
		return fmt.Errorf("rate limit access error: %w", accessErr)
	}

	if waitDuration > 0 {
		s.logger.Warnf("Rate limit exceeded, waiting %v before retry", waitDuration)
		time.Sleep(waitDuration)
		return s.checkRateLimit(ctx)
	}

	return nil
}

func (s *Input) fetchProductsPage(ctx context.Context) ([]any, error) {
	options := &goshopify.ListOptions{
		Limit: s.limit,
	}
	if s.lastUpdatedAt != nil {
		options.UpdatedAtMin = *s.lastUpdatedAt
	}

	products, err := s.client.Product.List(ctx, options)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(products))
	for i, p := range products {
		items[i] = p
	}
	return items, nil
}

func (s *Input) fetchOrdersPage(ctx context.Context) ([]any, error) {
	options := &goshopify.OrderListOptions{
		ListOptions: goshopify.ListOptions{
			Limit: s.limit,
		},
	}
	if s.lastUpdatedAt != nil {
		options.ListOptions.UpdatedAtMin = *s.lastUpdatedAt
	}

	orders, err := s.client.Order.List(ctx, options)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(orders))
	for i, o := range orders {
		items[i] = o
	}
	return items, nil
}

func (s *Input) fetchCustomersPage(ctx context.Context) ([]any, error) {
	options := &goshopify.ListOptions{
		Limit: s.limit,
	}
	if s.lastUpdatedAt != nil {
		options.UpdatedAtMin = *s.lastUpdatedAt
	}

	customers, err := s.client.Customer.List(ctx, options)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(customers))
	for i, c := range customers {
		items[i] = c
	}
	return items, nil
}

func (s *Input) fetchInventoryItemsPage(ctx context.Context) ([]any, error) {
	options := &goshopify.ListOptions{
		Limit: s.limit,
	}
	if s.lastUpdatedAt != nil {
		options.UpdatedAtMin = *s.lastUpdatedAt
	}

	inventoryItems, err := s.client.InventoryItem.List(ctx, options)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(inventoryItems))
	for i, item := range inventoryItems {
		items[i] = item
	}
	return items, nil
}

func (s *Input) fetchLocationsPage(ctx context.Context) ([]any, error) {
	options := &goshopify.ListOptions{
		Limit: s.limit,
	}
	if s.lastUpdatedAt != nil {
		options.UpdatedAtMin = *s.lastUpdatedAt
	}

	locations, err := s.client.Location.List(ctx, options)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(locations))
	for i, l := range locations {
		items[i] = l
	}
	return items, nil
}

func (s *Input) extractUpdatedAt(item any) *time.Time {
	// Try to extract updated_at from common Shopify resource types
	if m, ok := item.(map[string]any); ok {
		if updatedAtStr, ok := m["updated_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				return &t
			}
		}
	}

	// Use type assertion for typed objects
	switch v := item.(type) {
	case goshopify.Product:
		if v.UpdatedAt != nil && !v.UpdatedAt.IsZero() {
			return v.UpdatedAt
		}
	case goshopify.Order:
		if v.UpdatedAt != nil && !v.UpdatedAt.IsZero() {
			return v.UpdatedAt
		}
	case goshopify.Customer:
		if v.UpdatedAt != nil && !v.UpdatedAt.IsZero() {
			return v.UpdatedAt
		}
	case goshopify.InventoryItem:
		if v.UpdatedAt != nil && !v.UpdatedAt.IsZero() {
			return v.UpdatedAt
		}
	case goshopify.Location:
		if !v.UpdatedAt.IsZero() {
			updatedAt := v.UpdatedAt
			return &updatedAt
		}
	}

	return nil
}

func (s *Input) Close(ctx context.Context) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	s.connected = false
	s.logger.Info("Shopify input closed")

	return nil
}

func (s *Input) loadPosition(ctx context.Context) error {
	if s.cacheName == "" {
		return nil
	}

	var data []byte
	var getErr error

	err := s.mgr.AccessCache(ctx, s.cacheName, func(c service.Cache) {
		data, getErr = c.Get(ctx, s.cacheKey)
	})

	if err != nil {
		return fmt.Errorf("failed to access cache: %w", err)
	}

	if getErr != nil {
		if getErr == service.ErrKeyNotFound {
			s.logger.Debug("No previous position found in cache, starting from beginning")
			return nil
		}
		return fmt.Errorf("failed to get position from cache: %w", getErr)
	}

	lastUpdatedAtStr := string(data)
	if lastUpdatedAtStr == "" {
		return fmt.Errorf("cached position is empty")
	}

	parsedTime, err := time.Parse(time.RFC3339, lastUpdatedAtStr)
	if err != nil {
		return fmt.Errorf("failed to parse cached timestamp: %w", err)
	}

	s.lastUpdatedAt = &parsedTime
	s.latestAckAt = &parsedTime
	s.logger.Infof("Resuming from last position: updated_at_min=%s", lastUpdatedAtStr)

	return nil
}

func (s *Input) savePosition(ctx context.Context, updatedAt time.Time) {
	if s.cacheName == "" {
		return
	}

	updatedAtStr := updatedAt.Format(time.RFC3339)
	data := []byte(updatedAtStr)

	err := s.mgr.AccessCache(ctx, s.cacheName, func(c service.Cache) {
		if err := c.Set(ctx, s.cacheKey, data, nil); err != nil {
			s.logger.Errorf("Failed to save position to cache: %v", err)
		} else {
			s.logger.Debugf("Saved position to cache: %s", updatedAtStr)
		}
	})

	if err != nil {
		s.logger.Errorf("Failed to access cache: %v", err)
	}
}
