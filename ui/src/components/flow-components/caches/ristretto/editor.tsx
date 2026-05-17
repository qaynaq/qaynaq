import { NumberField, TextField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  default_ttl: string;
  max_cost: number;
  num_counters: number;
}

export default function RistrettoEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <TextField
        label="Default TTL"
        description="The default TTL of each item."
        value={value.default_ttl}
        onChange={(v) => onChange({ ...value, default_ttl: v })}
        error={errors?.default_ttl}
      />
      <NumberField
        label="Max Cost"
        description="The maximum size of the cache in bytes."
        min={1}
        value={value.max_cost}
        onChange={(v) => onChange({ ...value, max_cost: v })}
        error={errors?.max_cost}
      />
      <NumberField
        label="Num Counters"
        description="The number of 4-bit access counters to keep for admission and eviction."
        min={1}
        value={value.num_counters}
        onChange={(v) => onChange({ ...value, num_counters: v })}
        error={errors?.num_counters}
      />
    </div>
  );
}
