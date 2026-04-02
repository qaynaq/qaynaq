import { param, type McpToolTemplate, type TemplatePack } from "./types";

function shopifyTool(
  id: string,
  name: string,
  description: string,
  action: string,
  parameters: ReturnType<typeof param>[],
  configOverrides: Record<string, string | number | boolean> = {},
): McpToolTemplate {
  return {
    id,
    name,
    description,
    parameters,
    processor: {
      component: "shopify",
      config: { action, ...configOverrides },
    },
  };
}

export const shopifyPack: TemplatePack = {
  id: "shopify",
  name: "Shopify",
  description: "MCP tools for accessing your Shopify store data",
  sharedConfig: [
    {
      key: "shop_name",
      title: "Store Name",
      type: "input",
      required: true,
      placeholder: "mystore",
      description:
        "Your store name (without .myshopify.com)",
    },
    {
      key: "api_access_token",
      title: "Admin API Access Token",
      type: "input",
      required: true,
      secret: true,
      placeholder: "shpat_xxxxxxxxxxxxx",
      description:
        "From Settings > Apps > Develop apps > your app > API credentials",
    },
  ],
  templates: [
    shopifyTool(
      "shopify_list_orders",
      "shopify_list_orders",
      "List recent orders from your Shopify store",
      "list_orders",
      [
        param("limit", "Max orders to return (default 50, max 250)"),
        param("status", "Filter: open, closed, cancelled, any"),
      ],
      {
        limit: '${!this.limit.or("50")}',
        status: '${!this.status.or("")}',
      },
    ),
    shopifyTool(
      "shopify_list_products",
      "shopify_list_products",
      "List products from your Shopify store",
      "list_products",
      [param("limit", "Max products to return (default 50, max 250)")],
      { limit: '${!this.limit.or("50")}' },
    ),
    shopifyTool(
      "shopify_list_customers",
      "shopify_list_customers",
      "List customers from your Shopify store",
      "list_customers",
      [param("limit", "Max customers to return (default 50, max 250)")],
      { limit: '${!this.limit.or("50")}' },
    ),
    shopifyTool(
      "shopify_list_inventory_items",
      "shopify_list_inventory_items",
      "List inventory items from your Shopify store",
      "list_inventory_items",
      [param("limit", "Max items to return (default 50, max 250)")],
      { limit: '${!this.limit.or("50")}' },
    ),
    shopifyTool(
      "shopify_get_order",
      "shopify_get_order",
      "Get a specific order by ID from your Shopify store",
      "get_order",
      [param("order_id", "The Shopify order ID", true)],
      { order_id: "${!this.order_id}" },
    ),
    shopifyTool(
      "shopify_get_product",
      "shopify_get_product",
      "Get a specific product by ID from your Shopify store",
      "get_product",
      [param("product_id", "The Shopify product ID", true)],
      { product_id: "${!this.product_id}" },
    ),
    shopifyTool(
      "shopify_get_customer",
      "shopify_get_customer",
      "Get a specific customer by ID from your Shopify store",
      "get_customer",
      [param("customer_id", "The Shopify customer ID", true)],
      { customer_id: "${!this.customer_id}" },
    ),
  ],
};
