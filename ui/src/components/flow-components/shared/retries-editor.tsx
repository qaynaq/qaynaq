import { TextField } from "@/components/form-primitives";
import type { Retries } from "./retries";

interface Props {
  value: Retries;
  onChange: (next: Retries) => void;
  errors?: Record<string, string>;
  errorPathPrefix?: string;
}

export function RetriesEditor({
  value,
  onChange,
  errors,
  errorPathPrefix = "retries",
}: Props) {
  const err = (k: keyof Retries) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <TextField
        label="Initial Interval"
        size="sm"
        value={value.initial_interval}
        onChange={(v) => onChange({ ...value, initial_interval: v })}
        error={err("initial_interval")}
      />
      <TextField
        label="Max Interval"
        size="sm"
        value={value.max_interval}
        onChange={(v) => onChange({ ...value, max_interval: v })}
        error={err("max_interval")}
      />
      <TextField
        label="Max Elapsed Time"
        size="sm"
        value={value.max_elapsed_time}
        onChange={(v) => onChange({ ...value, max_elapsed_time: v })}
        error={err("max_elapsed_time")}
      />
    </div>
  );
}
