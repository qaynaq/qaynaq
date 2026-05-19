import { CheckboxField } from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function XmlDocumentsScannerEditor({
  value,
  onChange,
}: ScannerEditorProps<Config>) {
  return (
    <div className="space-y-3">
      <CheckboxField
        label="Cast"
        description="Try to cast number and boolean string values into their native types. Off means every value stays a string."
        checked={value.cast}
        onChange={(c) => onChange({ ...value, cast: c })}
      />
    </div>
  );
}
