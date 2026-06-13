---
sidebar_position: 4
---

# Shopify

Deploys up to 7 MCP tools for reading orders, products, customers, and inventory from your Shopify store.

## Getting Your Admin API Access Token

The Shopify template authenticates with a static Admin API access token from a custom app you create in your Shopify admin. Follow these steps once, then paste the token into the template wizard.

1. In your Shopify admin, go to **Settings** > **Apps** > **Develop apps**.
2. If this is your first custom app, click **Allow custom app development** and confirm.
3. Click **Create an app**, give it a name (e.g. `Qaynaq`), and click **Create app**.
4. Open the **Configuration** tab and click **Configure** under **Admin API integration**.
5. Select these four scopes, then click **Save**:

   | Scope | Enables |
   |-------|---------|
   | `read_orders` | Reading orders |
   | `read_products` | Reading products |
   | `read_customers` | Reading customers |
   | `read_inventory` | Reading inventory levels |

   These are all read-only. The template never writes to your store, so no `write_` scopes are needed.

6. Click **Install app**. The **Install** button only becomes active once you have selected at least one scope in step 5.
7. Installing takes you to the **API credentials** tab. Under **Admin API access token**, click **Reveal token once** and copy the token (it starts with `shpat_`) immediately.

:::warning
The token can be revealed **only once**, at install time. If you leave the page or come back to this tab later, the full token is no longer shown. Copy and save it in a secure place right away. The only way to recover from a lost token is to uninstall and reinstall the app, which generates a new one.
:::

:::note
The token is revoked if the app is uninstalled. If your flows start returning authentication errors, reinstall the custom app, copy the new token, and redeploy the template with **Override existing** enabled to update the stored secret.
:::

## Shared Configuration

| Field | Required | Description |
|-------|----------|-------------|
| Store Name | Yes | Your store name without `.myshopify.com` (e.g. `mystore`). |
| Admin API Access Token | Yes | The `shpat_` token from the steps above. Stored as a secret and referenced by name in the deployed flows, never inlined. |

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

:::note
By default `read_orders` returns orders from the last 60 days. Accessing older orders requires Shopify's protected customer data approval (the `read_all_orders` scope), which is granted automatically for a custom app on your own store but may need review in other setups.
:::
