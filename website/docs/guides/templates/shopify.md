---
sidebar_position: 4
---

# Shopify

Deploys up to 7 MCP tools for reading orders, products, customers, and inventory from your Shopify store.

## Shared Configuration

| Field | Required | Description |
|-------|----------|-------------|
| Store Name | Yes | Your store name without `.myshopify.com` (e.g. `mystore`). |
| Admin API Access Token | Yes | Stored as a secret and referenced by name in the deployed flows. Create the token in your Shopify admin under **Settings** > **Apps and sales channels** > **Develop apps** > your app > **API credentials**. |

The access token needs the `read_orders`, `read_products`, `read_customers`, and `read_inventory` scopes.

## Included Tools

| Tool | Description |
|------|-------------|
| `shopify_list_orders` | List recent orders with optional status filter and pagination |
| `shopify_list_products` | List products with pagination |
| `shopify_list_customers` | List customers with pagination |
| `shopify_list_inventory_items` | List inventory items with pagination |
| `shopify_get_order` | Get a specific order by ID |
| `shopify_get_product` | Get a specific product by ID |
| `shopify_get_customer` | Get a specific customer by ID |
