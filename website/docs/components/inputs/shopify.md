# Shopify

Fetches data from Shopify stores via the Admin API. This is the batch input component for ETL pipelines. For on-demand MCP tool access, see the [Shopify processor](/docs/components/processors/shopify).

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Shop Name | string | - | Shopify store name (without .myshopify.com) |
| API Key | string | - | Shopify API key (required for Private Apps, optional for Custom Apps) |
| Admin API Access Token | secret | - | Shopify Admin API access token |
| Resource | string | `products` | Resource type to fetch |
| Limit | integer | `50` | Results per page (max 250) |
| API Version | string | - | Shopify API version (e.g., 2024-01) |
| Rate Limit | string | - | Rate limit resource label |
| Cache | string | - | Cache resource for position tracking |

## Authentication

This input supports both **Custom App** and legacy **Private App** credentials.

**Custom App** (recommended - Private Apps were deprecated in January 2022):
1. Go to **Settings** > **Apps** > **Develop apps** in your Shopify admin
2. Create an app with the required read scopes
3. Install and copy the Admin API access token
4. Set `api_key` to empty and use only `api_access_token`

**Private App** (legacy):
- Provide both `api_key` and `api_access_token`

## Available Resources

`products`, `orders`, `customers`, `inventory_items`, `locations`
