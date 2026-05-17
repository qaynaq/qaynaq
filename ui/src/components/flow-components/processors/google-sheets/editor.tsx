import {
  TextField,
  CheckboxField,
  SelectField,
  ConnectionPickerField,
} from "@/components/form-primitives";
import { SHEETS_ACTIONS, SHEETS_PASTE, SHEETS_SORT } from ".";
import type { EditorProps } from "../../types";

interface Config {
  service_account_json: string;
  oauth_connection: string;
  delegate_to: string;
  action: (typeof SHEETS_ACTIONS)[number];
  spreadsheet_id: string;
  sheet_name: string;
  title: string;
  range: string;
  row_number: string;
  end_row_number: string;
  column_name: string;
  lookup_value: string;
  values: string;
  rows: string;
  max_results: string;
  new_name: string;
  destination_spreadsheet_id: string;
  source_range: string;
  destination_range: string;
  paste_type: (typeof SHEETS_PASTE)[number];
  sort_column_index: string;
  sort_order: (typeof SHEETS_SORT)[number];
  bold: boolean;
  italic: boolean;
  strikethrough: boolean;
  background_color: string;
  foreground_color: string;
  number_format: string;
  frozen_rows: string;
  frozen_columns: string;
  hidden: boolean;
  sheet_position: string;
  validation_type: string;
  validation_values: string;
  condition_type: string;
  condition_value: string;
  condition_background_color: string;
  include_headers: boolean;
}

export default function GoogleSheetsEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <ConnectionPickerField label="Service Account JSON" source="secrets" value={value.service_account_json} onChange={(v) => set("service_account_json", v)} />
      <ConnectionPickerField label="OAuth Connection" source="connections" value={value.oauth_connection} onChange={(v) => set("oauth_connection", v)} />
      <TextField label="Delegate To" value={value.delegate_to} onChange={(v) => set("delegate_to", v)} />
      <SelectField label="Action" required value={value.action} onChange={(v) => set("action", v as Config["action"])} options={SHEETS_ACTIONS as unknown as string[]} />
      <TextField label="Spreadsheet ID" value={value.spreadsheet_id} onChange={(v) => set("spreadsheet_id", v)} />
      <TextField label="Sheet Name" value={value.sheet_name} onChange={(v) => set("sheet_name", v)} />
      <TextField label="Title" value={value.title} onChange={(v) => set("title", v)} />
      <TextField label="Range" description="A1 notation (e.g. A1:D10)." value={value.range} onChange={(v) => set("range", v)} />
      <TextField label="Row Number" value={value.row_number} onChange={(v) => set("row_number", v)} />
      <TextField label="End Row Number" value={value.end_row_number} onChange={(v) => set("end_row_number", v)} />
      <TextField label="Column Name" value={value.column_name} onChange={(v) => set("column_name", v)} />
      <TextField label="Lookup Value" value={value.lookup_value} onChange={(v) => set("lookup_value", v)} />
      <TextField label="Values" description="JSON array for a single row." value={value.values} onChange={(v) => set("values", v)} />
      <TextField label="Rows" description="JSON array of arrays for multiple rows." value={value.rows} onChange={(v) => set("rows", v)} />
      <TextField label="Max Results" value={value.max_results} onChange={(v) => set("max_results", v)} />
      <TextField label="New Name" value={value.new_name} onChange={(v) => set("new_name", v)} />
      <TextField label="Destination Spreadsheet ID" value={value.destination_spreadsheet_id} onChange={(v) => set("destination_spreadsheet_id", v)} />
      <TextField label="Source Range" value={value.source_range} onChange={(v) => set("source_range", v)} />
      <TextField label="Destination Range" value={value.destination_range} onChange={(v) => set("destination_range", v)} />
      <SelectField label="Paste Type" value={value.paste_type} onChange={(v) => set("paste_type", v as Config["paste_type"])} options={SHEETS_PASTE as unknown as string[]} />
      <TextField label="Sort Column Index" value={value.sort_column_index} onChange={(v) => set("sort_column_index", v)} />
      <SelectField label="Sort Order" value={value.sort_order} onChange={(v) => set("sort_order", v as Config["sort_order"])} options={SHEETS_SORT as unknown as string[]} />
      <CheckboxField label="Bold" checked={value.bold} onChange={(c) => set("bold", c)} />
      <CheckboxField label="Italic" checked={value.italic} onChange={(c) => set("italic", c)} />
      <CheckboxField label="Strikethrough" checked={value.strikethrough} onChange={(c) => set("strikethrough", c)} />
      <TextField label="Background Color" value={value.background_color} onChange={(v) => set("background_color", v)} />
      <TextField label="Foreground Color" value={value.foreground_color} onChange={(v) => set("foreground_color", v)} />
      <TextField label="Number Format" value={value.number_format} onChange={(v) => set("number_format", v)} />
      <TextField label="Frozen Rows" value={value.frozen_rows} onChange={(v) => set("frozen_rows", v)} />
      <TextField label="Frozen Columns" value={value.frozen_columns} onChange={(v) => set("frozen_columns", v)} />
      <CheckboxField label="Hidden" checked={value.hidden} onChange={(c) => set("hidden", c)} />
      <TextField label="Sheet Position" value={value.sheet_position} onChange={(v) => set("sheet_position", v)} />
      <TextField label="Validation Type" value={value.validation_type} onChange={(v) => set("validation_type", v)} />
      <TextField label="Validation Values" value={value.validation_values} onChange={(v) => set("validation_values", v)} />
      <TextField label="Condition Type" value={value.condition_type} onChange={(v) => set("condition_type", v)} />
      <TextField label="Condition Value" value={value.condition_value} onChange={(v) => set("condition_value", v)} />
      <TextField label="Condition Background Color" value={value.condition_background_color} onChange={(v) => set("condition_background_color", v)} />
      <CheckboxField label="Include Headers" checked={value.include_headers} onChange={(c) => set("include_headers", c)} />
    </div>
  );
}
