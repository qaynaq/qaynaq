import {
  NumberField,
  CheckboxField,
  TextField,
  TextAreaField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  limit: number;
  spillover: boolean;
  batch_policy: {
    enabled: boolean;
    count: number;
    byte_size: number;
    period: string;
    jitter: number;
    check: string;
    processors: string;
  };
}

export default function MemoryBufferEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setBP = (next: Config["batch_policy"]) =>
    onChange({ ...value, batch_policy: next });

  return (
    <div className="space-y-4">
      <NumberField
        label="Limit"
        description="The maximum buffer size (in bytes) to allow before applying backpressure upstream."
        required
        min={1}
        value={value.limit}
        onChange={(v) => onChange({ ...value, limit: v })}
        error={errors?.limit}
      />
      <CheckboxField
        label="Spillover"
        description="Drop incoming messages that would exceed the buffer limit."
        checked={value.spillover}
        onChange={(c) => onChange({ ...value, spillover: c })}
      />
      <section className="space-y-3 border-t pt-3">
        <h4 className="text-sm font-medium">Batch Policy</h4>
        <CheckboxField
          label="Enabled"
          checked={value.batch_policy.enabled}
          onChange={(c) => setBP({ ...value.batch_policy, enabled: c })}
        />
        <NumberField
          label="Count"
          description="Number of messages at which the batch should be flushed. 0 disables count-based batching."
          size="sm"
          min={0}
          value={value.batch_policy.count}
          onChange={(v) => setBP({ ...value.batch_policy, count: v })}
        />
        <NumberField
          label="Byte Size"
          description="Bytes at which the batch should be flushed. 0 disables size-based batching."
          size="sm"
          min={0}
          value={value.batch_policy.byte_size}
          onChange={(v) => setBP({ ...value.batch_policy, byte_size: v })}
        />
        <TextField
          label="Period"
          description="A period in which an incomplete batch should be flushed regardless of its size."
          size="sm"
          value={value.batch_policy.period}
          onChange={(v) => setBP({ ...value.batch_policy, period: v })}
        />
        <NumberField
          label="Jitter"
          description="A non-negative factor that adds random delay to batch flush intervals (0-1)."
          size="sm"
          min={0}
          step={0.1}
          value={value.batch_policy.jitter}
          onChange={(v) => setBP({ ...value.batch_policy, jitter: v })}
        />
        <TextAreaField
          label="Check"
          description="A Bloblang query that should return a boolean indicating whether a message should end a batch."
          rows={2}
          value={value.batch_policy.check}
          onChange={(v) => setBP({ ...value.batch_policy, check: v })}
        />
        <TextAreaField
          label="Processors"
          description="A list of processors to apply to a batch as it is flushed (YAML array)."
          rows={2}
          value={value.batch_policy.processors}
          onChange={(v) => setBP({ ...value.batch_policy, processors: v })}
        />
      </section>
    </div>
  );
}
