import {
  TextField,
  NumberField,
  SelectField,
  ConnectionPickerField,
} from "@/components/form-primitives";
import { SHOPIFY_RESOURCES } from ".";
import type { EditorProps } from "../../types";

interface Config {
  shop_name: string;
  api_key: string;
  api_access_token: string;
  shop_resource: (typeof SHOPIFY_RESOURCES)[number];
  limit: number;
  api_version: string;
  cache_resource: string;
  rate_limit: string;
}

export default function ShopifyEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <TextField label="Shop Name" description="Shopify store name (without .myshopify.com)." required value={value.shop_name} onChange={(v) => set("shop_name", v)} error={errors?.shop_name} />
      <TextField label="API Key" required value={value.api_key} onChange={(v) => set("api_key", v)} error={errors?.api_key} />
      <TextField label="API Access Token" type="password" required value={value.api_access_token} onChange={(v) => set("api_access_token", v)} error={errors?.api_access_token} />
      <SelectField label="Shop Resource" value={value.shop_resource} onChange={(v) => set("shop_resource", v as Config["shop_resource"])} options={SHOPIFY_RESOURCES as unknown as string[]} />
      <NumberField label="Limit" description="Max items per API request (max 250)." min={1} max={250} value={value.limit} onChange={(v) => set("limit", v)} />
      <TextField label="API Version" description="Shopify API version (e.g. 2024-01). Empty = default." value={value.api_version} onChange={(v) => set("api_version", v)} />
      <ConnectionPickerField label="Cache Resource" description="Optional cache for the last updated_at timestamp." source="caches" value={value.cache_resource} onChange={(v) => set("cache_resource", v)} />
      <ConnectionPickerField label="Rate Limit" description="Rate limit resource for Shopify API requests." source="rate_limits" value={value.rate_limit} onChange={(v) => set("rate_limit", v)} />
    </div>
  );
}
