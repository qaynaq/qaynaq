import { CheckboxField } from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function AvroScannerEditor({
  value,
  onChange,
}: ScannerEditorProps<Config>) {
  return (
    <div className="space-y-3">
      <CheckboxField
        label="Avro Raw JSON"
        description="Decode into standard JSON instead of Avro JSON encoding (where union types are wrapped in objects)."
        checked={value.avro_raw_json}
        onChange={(c) => onChange({ ...value, avro_raw_json: c })}
      />
    </div>
  );
}
