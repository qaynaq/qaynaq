package shopify

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goshopify "github.com/bold-commerce/go-shopify/v4"
	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	err := service.RegisterProcessor(
		"shopify", ProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewProcessorFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type Processor struct {
	shopName       string
	apiAccessToken string
	rateLimitLabel string
	action         string

	limit      *service.InterpolatedString
	status     *service.InterpolatedString
	orderID    *service.InterpolatedString
	productID  *service.InterpolatedString
	customerID *service.InterpolatedString

	client *goshopify.Client
	mgr    *service.Resources
	logger *service.Logger
}

func NewProcessorFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	shopName, err := conf.FieldString(sbfShopName)
	if err != nil {
		return nil, err
	}

	apiAccessToken, err := conf.FieldString(sbfAPIAccessToken)
	if err != nil {
		return nil, err
	}

	action, err := conf.FieldString(spfAction)
	if err != nil {
		return nil, err
	}

	p := &Processor{
		shopName:       shopName,
		apiAccessToken: apiAccessToken,
		action:         action,
		mgr:            mgr,
		logger:         mgr.Logger(),
	}

	if conf.Contains(sbfRateLimit) {
		if p.rateLimitLabel, err = conf.FieldString(sbfRateLimit); err != nil {
			return nil, err
		}
	}

	if p.limit, err = conf.FieldInterpolatedString(spfLimit); err != nil {
		return nil, err
	}

	if conf.Contains(spfStatus) {
		if p.status, err = conf.FieldInterpolatedString(spfStatus); err != nil {
			return nil, err
		}
	}

	if conf.Contains(spfOrderID) {
		if p.orderID, err = conf.FieldInterpolatedString(spfOrderID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(spfProductID) {
		if p.productID, err = conf.FieldInterpolatedString(spfProductID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(spfCustomerID) {
		if p.customerID, err = conf.FieldInterpolatedString(spfCustomerID); err != nil {
			return nil, err
		}
	}

	// Create Shopify client with empty ApiKey (Custom App uses only access token)
	app := goshopify.App{
		ApiKey:   "",
		Password: apiAccessToken,
	}
	client, err := goshopify.NewClient(app, shopName, "", goshopify.WithRetry(3))
	if err != nil {
		return nil, fmt.Errorf("failed to create Shopify client: %w", err)
	}
	p.client = client

	p.logger.Infof("Shopify processor configured for store: %s.myshopify.com", shopName)

	return p, nil
}

type resolvedProcessorFields struct {
	limit      int
	status     string
	orderID    uint64
	productID  uint64
	customerID uint64
}

func (p *Processor) resolveFields(msg *service.Message) (*resolvedProcessorFields, error) {
	r := &resolvedProcessorFields{}

	limitStr, err := p.limit.TryString(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to interpolate limit: %w", err)
	}
	r.limit, err = strconv.Atoi(limitStr)
	if err != nil || r.limit < 1 || r.limit > 250 {
		r.limit = 50
	}

	if p.status != nil {
		if r.status, err = p.status.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate status: %w", err)
		}
	}

	if p.orderID != nil {
		s, _ := p.orderID.TryString(msg)
		if s != "" {
			r.orderID, _ = strconv.ParseUint(s, 10, 64)
		}
	}

	if p.productID != nil {
		s, _ := p.productID.TryString(msg)
		if s != "" {
			r.productID, _ = strconv.ParseUint(s, 10, 64)
		}
	}

	if p.customerID != nil {
		s, _ := p.customerID.TryString(msg)
		if s != "" {
			r.customerID, _ = strconv.ParseUint(s, 10, 64)
		}
	}

	return r, nil
}

func (p *Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	fields, err := p.resolveFields(msg)
	if err != nil {
		return nil, classifyShopifyError(err)
	}

	if p.rateLimitLabel != "" {
		if err := p.checkRateLimit(ctx); err != nil {
			return nil, classifyShopifyError(fmt.Errorf("rate limit check failed: %w", err))
		}
	}

	var result map[string]any

	switch p.action {
	case actionListOrders:
		result, err = p.listOrders(ctx, fields)
	case actionListProducts:
		result, err = p.listProducts(ctx, fields)
	case actionListCustomers:
		result, err = p.listCustomers(ctx, fields)
	case actionListInventoryItems:
		result, err = p.listInventoryItems(ctx, fields)
	case actionGetOrder:
		result, err = p.getOrder(ctx, fields)
	case actionGetProduct:
		result, err = p.getProduct(ctx, fields)
	case actionGetCustomer:
		result, err = p.getCustomer(ctx, fields)
	default:
		err = fmt.Errorf("unsupported action: %s", p.action)
	}

	if err != nil {
		return nil, classifyShopifyError(err)
	}

	outMsg := msg.Copy()
	outMsg.SetStructured(result)
	return service.MessageBatch{outMsg}, nil
}

func (p *Processor) Close(ctx context.Context) error {
	return nil
}

func (p *Processor) checkRateLimit(ctx context.Context) error {
	var waitDuration time.Duration
	var accessErr error

	keyCtx := context.WithValue(ctx, rateLimitKeyContextKey, p.shopName)

	err := p.mgr.AccessRateLimit(ctx, p.rateLimitLabel, func(rl service.RateLimit) {
		waitDuration, accessErr = rl.Access(keyCtx)
	})

	if err != nil {
		return fmt.Errorf("failed to access rate limit: %w", err)
	}

	if accessErr != nil {
		return fmt.Errorf("rate limit access error: %w", accessErr)
	}

	if waitDuration > 0 {
		p.logger.Warnf("Rate limit exceeded, waiting %v before retry", waitDuration)
		time.Sleep(waitDuration)
		return p.checkRateLimit(ctx)
	}

	return nil
}

func (p *Processor) listOrders(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	options := &goshopify.OrderListOptions{
		ListOptions: goshopify.ListOptions{
			Limit: f.limit,
		},
	}
	if f.status != "" {
		options.Status = goshopify.OrderStatus(f.status)
	}

	orders, err := p.client.Order.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	items := make([]map[string]any, len(orders))
	for i, o := range orders {
		m, err := structToMap(o)
		if err != nil {
			return nil, fmt.Errorf("failed to convert order: %w", err)
		}
		items[i] = m
	}

	return map[string]any{
		"orders": items,
		"count":  len(items),
	}, nil
}

func (p *Processor) listProducts(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	options := &goshopify.ListOptions{
		Limit: f.limit,
	}

	products, err := p.client.Product.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	items := make([]map[string]any, len(products))
	for i, prod := range products {
		m, err := structToMap(prod)
		if err != nil {
			return nil, fmt.Errorf("failed to convert product: %w", err)
		}
		items[i] = m
	}

	return map[string]any{
		"products": items,
		"count":    len(items),
	}, nil
}

func (p *Processor) listCustomers(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	options := &goshopify.ListOptions{
		Limit: f.limit,
	}

	customers, err := p.client.Customer.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	items := make([]map[string]any, len(customers))
	for i, c := range customers {
		m, err := structToMap(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert customer: %w", err)
		}
		items[i] = m
	}

	return map[string]any{
		"customers": items,
		"count":     len(items),
	}, nil
}

func (p *Processor) listInventoryItems(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	options := &goshopify.ListOptions{
		Limit: f.limit,
	}

	inventoryItems, err := p.client.InventoryItem.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory items: %w", err)
	}

	items := make([]map[string]any, len(inventoryItems))
	for i, item := range inventoryItems {
		m, err := structToMap(item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert inventory item: %w", err)
		}
		items[i] = m
	}

	return map[string]any{
		"inventory_items": items,
		"count":           len(items),
	}, nil
}

func (p *Processor) getOrder(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	if f.orderID == 0 {
		return nil, fmt.Errorf("order_id is required for get_order action")
	}

	order, err := p.client.Order.Get(ctx, f.orderID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get order %d: %w", f.orderID, err)
	}

	result, err := structToMap(*order)
	if err != nil {
		return nil, fmt.Errorf("failed to convert order: %w", err)
	}

	return result, nil
}

func (p *Processor) getProduct(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	if f.productID == 0 {
		return nil, fmt.Errorf("product_id is required for get_product action")
	}

	product, err := p.client.Product.Get(ctx, f.productID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get product %d: %w", f.productID, err)
	}

	result, err := structToMap(*product)
	if err != nil {
		return nil, fmt.Errorf("failed to convert product: %w", err)
	}

	return result, nil
}

func (p *Processor) getCustomer(ctx context.Context, f *resolvedProcessorFields) (map[string]any, error) {
	if f.customerID == 0 {
		return nil, fmt.Errorf("customer_id is required for get_customer action")
	}

	customer, err := p.client.Customer.Get(ctx, f.customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer %d: %w", f.customerID, err)
	}

	result, err := structToMap(*customer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert customer: %w", err)
	}

	return result, nil
}

func classifyShopifyError(err error) error {
	msg := err.Error()

	if contains(msg, "is required", "must be set", "failed to interpolate", "invalid", "unsupported action") {
		return fmt.Errorf("[400] %s", msg)
	}

	if contains(msg, "401", "Unauthorized", "authentication") {
		return fmt.Errorf("[401] %s", msg)
	}

	if contains(msg, "404", "Not Found") {
		return fmt.Errorf("[404] %s", msg)
	}

	if contains(msg, "429", "rate limit", "exceeded") {
		return fmt.Errorf("[429] %s", msg)
	}

	return fmt.Errorf("[500] %s", msg)
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
