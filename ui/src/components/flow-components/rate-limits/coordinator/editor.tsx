import { NumberField, SelectField } from "@/components/form-primitives";
import { INTERVAL_OPTIONS } from ".";
import type { EditorProps } from "../../types";

interface Config {
  count: number;
  interval: (typeof INTERVAL_OPTIONS)[number];
  burst: number;
}

export default function CoordinatorEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <NumberField
        label="Count"
        description="Number of requests allowed per interval."
        required
        min={1}
        value={value.count}
        onChange={(v) => onChange({ ...value, count: v })}
        error={errors?.count}
      />
      <SelectField
        label="Interval"
        description="Time interval for rate limiting."
        required
        value={value.interval}
        onChange={(v) => onChange({ ...value, interval: v as Config["interval"] })}
        options={INTERVAL_OPTIONS as unknown as string[]}
        error={errors?.interval}
      />
      <NumberField
        label="Burst"
        description="Additional burst capacity for handling traffic spikes."
        min={0}
        value={value.burst}
        onChange={(v) => onChange({ ...value, burst: v })}
        error={errors?.burst}
      />
    </div>
  );
}
