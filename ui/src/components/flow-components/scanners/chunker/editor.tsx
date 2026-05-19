import { NumberField } from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function ChunkerScannerEditor({
  value,
  onChange,
  errors,
}: ScannerEditorProps<Config>) {
  return (
    <div className="space-y-3">
      <NumberField
        label="Size"
        description="Number of bytes per chunk."
        required
        size="sm"
        min={1}
        value={value.size}
        onChange={(v) => onChange({ ...value, size: v })}
        error={errors?.size}
      />
    </div>
  );
}
