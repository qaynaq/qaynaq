import {
  ArrayField,
  CheckboxField,
  NumberField,
  TextField,
} from "@/components/form-primitives";
import type { ScannerEditorProps } from "../types";
import type { Config } from ".";

export default function CsvScannerEditor({
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
        description="Override the default comma. Use a single character or a string."
        size="sm"
        value={value.custom_delimiter}
        onChange={(v) => set("custom_delimiter", v)}
      />
      <CheckboxField
        label="Parse Header Row"
        description="Use the first row as field names. When off, rows are emitted as arrays."
        checked={value.parse_header_row}
        onChange={(c) => set("parse_header_row", c)}
      />
      <CheckboxField
        label="Lazy Quotes"
        description="Allow quotes inside unquoted fields and unescaped quotes inside quoted fields."
        checked={value.lazy_quotes}
        onChange={(c) => set("lazy_quotes", c)}
      />
      <CheckboxField
        label="Continue On Error"
        description="If a row fails to parse, emit an error message and keep going instead of stopping."
        checked={value.continue_on_error}
        onChange={(c) => set("continue_on_error", c)}
      />
      <ArrayField
        label="Expected Headers"
        description="Optional list of header names to assert. The scanner errors if the file's header row does not match. Requires Parse Header Row."
        size="sm"
        value={value.expected_headers}
        onChange={(v) => set("expected_headers", v)}
        error={errors?.expected_headers}
      />
      <NumberField
        label="Expected Number Of Fields"
        description="Optional. Assert that every row has exactly this many fields. 0 disables the check."
        size="sm"
        min={0}
        value={value.expected_number_of_fields}
        onChange={(v) => set("expected_number_of_fields", v)}
      />
    </div>
  );
}
