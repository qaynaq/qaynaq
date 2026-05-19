import {
  CheckboxField,
  NumberField,
  TextField,
} from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function LinesScannerEditor({
  value,
  onChange,
  errors,
}: ScannerEditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-3">
      <TextField
        label="Custom Delimiter"
        description="Use a custom string to mark line endings instead of a single newline. Leave empty for the default."
        size="sm"
        value={value.custom_delimiter}
        onChange={(v) => set("custom_delimiter", v)}
      />
      <NumberField
        label="Max Buffer Size"
        description="Maximum size in bytes that a single line can reach before producing an error."
        size="sm"
        min={1}
        value={value.max_buffer_size}
        onChange={(v) => set("max_buffer_size", v)}
        error={errors?.max_buffer_size}
      />
      <CheckboxField
        label="Omit Empty"
        description="Skip empty lines instead of emitting empty messages."
        checked={value.omit_empty}
        onChange={(c) => set("omit_empty", c)}
      />
    </div>
  );
}
