package shopify

import "github.com/warpstreamlabs/bento/public/service"

const (
	sbfShopName       = "shop_name"
	sbfAPIKey         = "api_key"
	sbfAPIAccessToken = "api_access_token"
	sbfShopResource   = "shop_resource"
	sbfLimit          = "limit"
	sbfAPIVersion     = "api_version"
	sbfCache          = "cache_resource"
	sbfRateLimit      = "rate_limit"
)

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Services").
		Summary("Fetches all data from Shopify API with pagination.").
		Description(`
This input connects to Shopify and retrieves all data from your store using the Shopify Admin API.

The component requires authentication credentials for a Custom App (or legacy Private App):
- shop_name: Your Shopify store name (e.g., 'mystore' for mystore.myshopify.com)
- api_key: Your Shopify API key
- api_access_token: Your Shopify API Access Token

The component fetches all items of the specified resource type using pagination. It will iterate
through all pages until all data is retrieved, then close with ErrEndOfInput.

Position tracking: When a cache resource is configured, the component stores the last successfully
processed item's updated_at timestamp. On restart, it uses updated_at_min to fetch only items that
were created or updated after that timestamp. This ensures you get both new entities and updates to
existing entities.

For scheduled/periodic fetching, combine this input with the 'generate' input:
https://warpstreamlabs.github.io/bento/docs/components/inputs/generate

Supported resources:
- products: Fetch products from your store
- orders: Fetch orders from your store
- customers: Fetch customers from your store
- inventory_items: Fetch inventory items
- locations: Fetch locations

Each message contains the raw resource data from the Shopify API.`).
		Field(service.NewStringField(sbfShopName).
			Description("Shopify store name (without .myshopify.com).")).
		Field(service.NewStringField(sbfAPIKey).
			Description("Shopify API key for authentication. Required for Private Apps, optional for Custom Apps.")).
		Field(service.NewStringField(sbfAPIAccessToken).
			Description("Shopify Admin API Access Token for authentication.").
			Secret()).
		Field(service.NewStringField(sbfShopResource).
			Description("The Shopify resource type to fetch (e.g., products, orders, customers, inventory_items, locations).").
			Default("products")).
		Field(service.NewIntField(sbfLimit).
			Description("Maximum number of items to fetch per API request (max 250).").
			Default(50)).
		Field(service.NewStringField(sbfAPIVersion).
			Description("Shopify API version to use (e.g., '2024-01'). If not specified, uses the default version.").
			Optional()).
		Field(service.NewStringField(sbfCache).
			Description("Optional cache resource name for storing the last updated_at timestamp. When configured, resumes fetching from items updated after that timestamp, ensuring both new and updated entities are captured.").
			Optional()).
		Field(service.NewStringField(sbfRateLimit).
			Description("Rate limit resource to use for Shopify API requests. Uses shop name as the rate limit key.").
			Optional()).
		Version("1.0.0").
		Example("Fetch all products from Shopify with rate limiting",
			`Fetch all products with coordinator-based rate limiting`,
			`
rate_limit_resources:
  - label: internal
    coordinator:
      count: 2
      interval: "1s"
      burst: 5

input:
  shopify:
    shop_name: mystore
    api_key: your_api_key_here
    api_access_token: your_access_token_here
    shop_resource: products
    limit: 50
    rate_limit: internal
`,
		).
		Example("Fetch all products from Shopify",
			`Fetch all products with pagination`,
			`
input:
  shopify:
    shop_name: mystore
    api_key: your_api_key_here
    api_access_token: your_access_token_here
    shop_resource: products
    limit: 50
`,
		).
		Example("Scheduled fetch with generate input",
			`Fetch all orders every 5 minutes using generate`,
			`
input:
  generate:
    interval: 5m
    mapping: 'root = {}'
  processors:
    - branch:
        request_map: 'root = deleted()'
        processors:
          - shopify:
              shop_name: ${SHOPIFY_SHOP_NAME}
              api_key: ${SHOPIFY_API_KEY}
              api_access_token: ${SHOPIFY_ACCESS_TOKEN}
              shop_resource: orders
              limit: 100
`,
		).
		Example("Fetch customers with API version",
			`Fetch all customers using a specific API version`,
			`
input:
  shopify:
    shop_name: mystore
    api_key: ${SHOPIFY_API_KEY}
    api_access_token: ${SHOPIFY_ACCESS_TOKEN}
    shop_resource: customers
    limit: 250
    api_version: "2024-01"
`,
		).
		Example("Fetch with position tracking",
			`Use cache to track updated_at and resume on restart to get new and updated items`,
			`
cache_resources:
  - label: shopify_position
    file:
      directory: /tmp/bento/cache

input:
  shopify:
    shop_name: ${SHOPIFY_SHOP_NAME}
    api_key: ${SHOPIFY_API_KEY}
    api_access_token: ${SHOPIFY_ACCESS_TOKEN}
    shop_resource: products
    limit: 100
    cache_resource: shopify_position
`,
		)
}
