import {
  NumberField,
  TextField,
} from "@/components/form-primitives";
import type { Batching } from "./batching";

interface Props {
  value: Batching;
  onChange: (next: Batching) => void;
  errors?: Record<string, string>;
  errorPathPrefix?: string;
}

export function BatchingEditor({
  value,
  onChange,
  errors,
  errorPathPrefix = "batching",
}: Props) {
  const err = (k: keyof Batching) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <NumberField
        label="Count"
        description="Number of messages at which the batch should be flushed. 0 disables count-based batching."
        size="sm"
        min={0}
        value={value.count}
        onChange={(v) => onChange({ ...value, count: v })}
        error={err("count")}
      />
      <NumberField
        label="Byte Size"
        description="Bytes at which the batch should be flushed. 0 disables size-based batching."
        size="sm"
        min={0}
        value={value.byte_size}
        onChange={(v) => onChange({ ...value, byte_size: v })}
        error={err("byte_size")}
      />
      <TextField
        label="Period"
        description="A period in which an incomplete batch should be flushed regardless of its size (e.g. 1s, 30ms)."
        size="sm"
        value={value.period}
        onChange={(v) => onChange({ ...value, period: v })}
        error={err("period")}
      />
      <NumberField
        label="Jitter"
        description="A non-negative factor that adds random delay to batch flush intervals (0-1)."
        size="sm"
        min={0}
        step={0.1}
        value={value.jitter}
        onChange={(v) => onChange({ ...value, jitter: v })}
        error={err("jitter")}
      />
      <TextField
        label="Check"
        description="A Bloblang query that should return a boolean indicating whether a message should end a batch."
        size="sm"
        value={value.check}
        onChange={(v) => onChange({ ...value, check: v })}
        error={err("check")}
      />
    </div>
  );
}
