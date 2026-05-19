import { NumberField, TextField } from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function ReMatchScannerEditor({
  value,
  onChange,
  errors,
}: ScannerEditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-3">
      <TextField
        label="Pattern"
        description="Regular expression. Each match marks the start of a new message."
        required
        size="sm"
        value={value.pattern}
        onChange={(v) => set("pattern", v)}
        error={errors?.pattern}
        placeholder={String.raw`(?m)^\d\d:\d\d:\d\d`}
      />
      <NumberField
        label="Max Buffer Size"
        description="Maximum size in bytes that a single message can reach before producing an error."
        size="sm"
        min={1}
        value={value.max_buffer_size}
        onChange={(v) => set("max_buffer_size", v)}
      />
    </div>
  );
}
