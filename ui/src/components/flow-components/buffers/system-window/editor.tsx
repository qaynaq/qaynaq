import { TextAreaField, TextField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  timestamp_mapping: string;
  size: string;
  slide: string;
  offset: string;
  allowed_lateness: string;
}

export default function SystemWindowEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <TextAreaField
        label="Timestamp Mapping"
        description="A Bloblang mapping that provides the timestamp to use for allocating messages to windows."
        value={value.timestamp_mapping}
        onChange={(v) => onChange({ ...value, timestamp_mapping: v })}
        error={errors?.timestamp_mapping}
      />
      <TextField
        label="Size"
        description="A duration string describing the size of each window (e.g., 30s, 10m, 1h)."
        required
        value={value.size}
        onChange={(v) => onChange({ ...value, size: v })}
        error={errors?.size}
      />
      <TextField
        label="Slide"
        description="An optional duration for sliding windows. Must be smaller than the size (e.g., 30s, 10m)."
        value={value.slide}
        onChange={(v) => onChange({ ...value, slide: v })}
        error={errors?.slide}
      />
      <TextField
        label="Offset"
        description="An optional duration to offset the beginning of each window (e.g., -6h, 30m)."
        value={value.offset}
        onChange={(v) => onChange({ ...value, offset: v })}
        error={errors?.offset}
      />
      <TextField
        label="Allowed Lateness"
        description="Length of time to wait after a window ends before flushing it (e.g., 10s, 1m)."
        value={value.allowed_lateness}
        onChange={(v) => onChange({ ...value, allowed_lateness: v })}
        error={errors?.allowed_lateness}
      />
    </div>
  );
}
