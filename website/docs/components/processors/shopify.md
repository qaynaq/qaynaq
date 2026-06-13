# Shopify

Performs on-demand Shopify Admin API operations for MCP tools - list and retrieve orders, products, customers, and inventory.

:::tip Template Available
Want to expose Shopify actions as MCP tools for AI assistants? Use the [Shopify template](/docs/guides/templates/shopify) to deploy all 7 tools in one step - no manual configuration needed.
:::

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Shop Name | string | - | Shopify store name (without .myshopify.com) |
| Admin API Access Token | secret | - | Custom App access token (starts with shpat_) |
| Action | select | - | The operation to perform (required) |
| Limit | string | `50` | Max items to return for list operations (1-250) |
| Status | string | - | Order status filter (open, closed, cancelled, any) |
| Order ID | string | - | Order ID for get_order |
| Product ID | string | - | Product ID for get_product |
| Customer ID | string | - | Customer ID for get_customer |
| Since ID | string | - | Return items after this ID for pagination |
| Rate Limit | string | - | Rate limit resource label |

## Authentication

This processor uses **Custom App** authentication (Private Apps were deprecated in January 2022).

### Creating a Custom App

1. In your Shopify admin, go to **Settings** > **Apps**
2. Click **Develop apps** (enable developer mode if prompted)
3. Click **Create an app** and give it a name
4. Go to **Configuration** > **Admin API integration**
5. Select the required scopes:
   - `read_orders` - for order operations
   - `read_products` - for product operations
   - `read_customers` - for customer operations
   - `read_inventory` - for inventory operations
6. Click **Install app** (the button activates once scopes are selected)
7. On the **API credentials** tab, reveal and copy the **Admin API access token** (starts with `shpat_`) immediately - it can only be revealed once at install time. If you lose it, uninstall and reinstall the app to generate a new one.

Store the access token as a secret in **Secrets**, then select it in the **Admin API Access Token** field when configuring the processor.

## Actions

### list_orders

List recent orders from the store.

| Parameter | Description |
|-----------|-------------|
| limit | Max orders to return (default 50, max 250) |
| status | Filter: open, closed, cancelled, any |
| since_id | Return orders after this ID for pagination |

### list_products

List products from the store.

| Parameter | Description |
|-----------|-------------|
| limit | Max products to return (default 50, max 250) |
| since_id | Return products after this ID for pagination |

### list_customers

List customers from the store.

| Parameter | Description |
|-----------|-------------|
| limit | Max customers to return (default 50, max 250) |
| since_id | Return customers after this ID for pagination |

### list_inventory_items

List inventory items from the store.

| Parameter | Description |
|-----------|-------------|
| limit | Max items to return (default 50, max 250) |
| since_id | Return items after this ID for pagination |

### get_order

Get a specific order by ID.

| Parameter | Description |
|-----------|-------------|
| order_id | The Shopify order ID (required) |

### get_product

Get a specific product by ID.

| Parameter | Description |
|-----------|-------------|
| product_id | The Shopify product ID (required) |

### get_customer

Get a specific customer by ID.

| Parameter | Description |
|-----------|-------------|
| customer_id | The Shopify customer ID (required) |

## Dynamic Fields

Most action parameter fields support interpolation using `${!this.field_name}` syntax, allowing dynamic values from the incoming message. This is how MCP tool parameters are passed to the processor at runtime.

## Processor vs Input

Qaynaq has two Shopify components:

- **Processor** (this page) - On-demand data retrieval, designed for MCP tools. Each message triggers a single API call and returns the result.
- **[Input](/docs/components/inputs/shopify)** - Batch data ingestion. Fetches all items of a resource type with pagination, designed for ETL pipelines.
