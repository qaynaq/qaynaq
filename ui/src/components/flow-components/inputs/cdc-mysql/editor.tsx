import {
  TextField,
  NumberField,
  SelectField,
  ArrayField,
  CheckboxField,
  ConnectionPickerField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  host: string;
  port: number;
  user: string;
  password: string;
  server_id: string;
  position_cache: string;
  position_cache_key: string;
  position_mode: "gtid" | "file";
  flavor: "mysql" | "mariadb";
  cache_save_interval: string;
  include_tables: string[];
  exclude_tables: string[];
  use_schema_cache: boolean;
  schema_cache_ttl: string;
  max_batch_size: number;
  max_pending_checkpoints: number;
  retry_initial_interval: string;
  retry_max_interval: string;
  retry_multiplier: number;
}

export default function CdcMysqlEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <TextField label="Host" required value={value.host} onChange={(v) => set("host", v)} error={errors?.host} />
      <NumberField label="Port" min={1} value={value.port} onChange={(v) => set("port", v)} />
      <TextField label="User" required value={value.user} onChange={(v) => set("user", v)} error={errors?.user} />
      <TextField label="Password" required type="password" value={value.password} onChange={(v) => set("password", v)} error={errors?.password} />
      <TextField label="Server ID" description="Unique server ID for this binlog consumer." value={value.server_id} onChange={(v) => set("server_id", v)} />
      <ConnectionPickerField
        label="Position Cache"
        description="Cache resource to use for position tracking."
        required
        source="caches"
        value={value.position_cache}
        onChange={(v) => set("position_cache", v)}
      />
      <TextField label="Position Cache Key" required value={value.position_cache_key} onChange={(v) => set("position_cache_key", v)} error={errors?.position_cache_key} />
      <SelectField label="Position Mode" value={value.position_mode} onChange={(v) => set("position_mode", v as Config["position_mode"])} options={["gtid", "file"]} />
      <SelectField label="Flavor" value={value.flavor} onChange={(v) => set("flavor", v as Config["flavor"])} options={["mysql", "mariadb"]} />
      <TextField label="Cache Save Interval" description="Interval for saving binlog position to cache." value={value.cache_save_interval} onChange={(v) => set("cache_save_interval", v)} />
      <ArrayField label="Include Tables" description="List of tables to monitor in schema.table format. Empty = all." value={value.include_tables} onChange={(v) => set("include_tables", v)} />
      <ArrayField label="Exclude Tables" value={value.exclude_tables} onChange={(v) => set("exclude_tables", v)} />
      <CheckboxField label="Use Schema Cache" checked={value.use_schema_cache} onChange={(c) => set("use_schema_cache", c)} />
      <TextField label="Schema Cache TTL" value={value.schema_cache_ttl} onChange={(v) => set("schema_cache_ttl", v)} />
      <NumberField label="Max Batch Size" min={1} value={value.max_batch_size} onChange={(v) => set("max_batch_size", v)} />
      <NumberField label="Max Pending Checkpoints" min={1} value={value.max_pending_checkpoints} onChange={(v) => set("max_pending_checkpoints", v)} />
      <TextField label="Retry Initial Interval" value={value.retry_initial_interval} onChange={(v) => set("retry_initial_interval", v)} />
      <TextField label="Retry Max Interval" value={value.retry_max_interval} onChange={(v) => set("retry_max_interval", v)} />
      <NumberField label="Retry Multiplier" min={1} step={0.1} value={value.retry_multiplier} onChange={(v) => set("retry_multiplier", v)} />
    </div>
  );
}
