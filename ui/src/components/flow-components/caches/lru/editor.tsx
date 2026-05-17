import { NumberField, KeyValueField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  cap: number;
  init_values: Record<string, string>;
}

export default function LruEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <NumberField
        label="Capacity"
        description="The maximum number of items to store in the cache."
        required
        min={1}
        value={value.cap}
        onChange={(v) => onChange({ ...value, cap: v })}
        error={errors?.cap}
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
