import {
  TextField,
  NumberField,
  KeyValueField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  default_ttl: string;
  compaction_interval: string;
  init_values: Record<string, string>;
  shards: number;
}

export default function MemoryCacheEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <TextField
        label="Default TTL"
        description="The default TTL of each item. After this period an item will be eligible for removal during the next compaction."
        value={value.default_ttl}
        onChange={(v) => onChange({ ...value, default_ttl: v })}
        error={errors?.default_ttl}
      />
      <TextField
        label="Compaction Interval"
        description="The period to wait between compactions. Empty string disables expiry."
        value={value.compaction_interval}
        onChange={(v) => onChange({ ...value, compaction_interval: v })}
        error={errors?.compaction_interval}
      />
      <NumberField
        label="Shards"
        description="A number of logical shards to spread keys across."
        min={1}
        value={value.shards}
        onChange={(v) => onChange({ ...value, shards: v })}
        error={errors?.shards}
      />
      <KeyValueField
        label="Init Values"
        description="A table of key/value pairs that should be present in the cache on initialization."
        value={value.init_values}
        onChange={(v) => onChange({ ...value, init_values: v })}
      />
    </div>
  );
}
