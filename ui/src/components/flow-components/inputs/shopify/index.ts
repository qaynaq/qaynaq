import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const SHOPIFY_RESOURCES = [
  "products",
  "orders",
  "customers",
  "inventory_items",
  "locations",
] as const;

const configSchema = z.object({
  shop_name: z.string().min(1, "Required"),
  api_key: z.string().min(1, "Required"),
  api_access_token: z.string().min(1, "Required"),
  shop_resource: z.enum(SHOPIFY_RESOURCES),
  limit: z.number().int().min(1).max(250),
  api_version: z.string(),
  cache_resource: z.string(),
  rate_limit: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  shop_name: "",
  api_key: "",
  api_access_token: "",
  shop_resource: "products",
  limit: 50,
  api_version: "",
  cache_resource: "",
  rate_limit: "",
};

const component: FlowComponent<Config> = {
  id: "shopify",
  name: "Shopify",
  category: "input",
  description: "Fetches Shopify resources via the Admin API.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
