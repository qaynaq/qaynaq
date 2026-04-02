package shopify

import "github.com/warpstreamlabs/bento/public/service"

const (
	spfAction     = "action"
	spfLimit      = "limit"
	spfStatus     = "status"
	spfOrderID    = "order_id"
	spfProductID  = "product_id"
	spfCustomerID = "customer_id"
)

const (
	actionListOrders         = "list_orders"
	actionListProducts       = "list_products"
	actionListCustomers      = "list_customers"
	actionListInventoryItems = "list_inventory_items"
	actionGetOrder           = "get_order"
	actionGetProduct         = "get_product"
	actionGetCustomer        = "get_customer"
)

func ProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Services").
		Summary("Performs on-demand Shopify Admin API operations for MCP tools.").
		Description(`
This processor connects to the Shopify Admin API using a Custom App access token and
performs on-demand data retrieval operations. It is designed for use as an MCP tool processor,
where each incoming message triggers a single API call.

Authentication requires a Custom App (not a Private App, which was deprecated in January 2022):
- shop_name: Your Shopify store name (e.g., 'mystore' for mystore.myshopify.com)
- api_access_token: Your Custom App Admin API access token (starts with shpat_)

To create a Custom App:
1. Go to Settings > Apps > Develop apps in your Shopify admin
2. Click "Create an app" and name it
3. Configure Admin API scopes (read_orders, read_products, read_customers, read_inventory)
4. Install the app and copy the Admin API access token

For batch data ingestion, use the shopify input component instead.

Most action parameter fields support interpolation using the ` + "`${!this.field_name}`" + ` syntax,
allowing dynamic values from the incoming message.`).
		Field(service.NewStringField(sbfShopName).
			Description("Shopify store name (without .myshopify.com).")).
		Field(service.NewStringField(sbfAPIAccessToken).
			Description("Custom App Admin API access token (starts with shpat_).").
			Secret()).
		Field(service.NewStringField(sbfRateLimit).
			Description("Rate limit resource label for Shopify API requests. Uses shop name as the rate limit key.").
			Optional()).
		Field(service.NewStringEnumField(spfAction,
			actionListOrders,
			actionListProducts,
			actionListCustomers,
			actionListInventoryItems,
			actionGetOrder,
			actionGetProduct,
			actionGetCustomer,
		).Description("The Shopify operation to perform.")).
		Field(service.NewInterpolatedStringField(spfLimit).
			Description("Maximum number of items to return for list operations (1-250).").
			Default("50")).
		Field(service.NewInterpolatedStringField(spfStatus).
			Description("Order status filter: open, closed, cancelled, any. Used by list_orders.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(spfOrderID).
			Description("Order ID for get_order action.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(spfProductID).
			Description("Product ID for get_product action.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(spfCustomerID).
			Description("Customer ID for get_customer action.").
			Default("").
			Optional()).
		Version("1.1.0")
}
