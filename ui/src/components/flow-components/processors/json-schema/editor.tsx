import {
  ConnectionPickerField,
  CodeField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  schema_path: string;
  schema: string;
}

export default function JsonSchemaEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <ConnectionPickerField
        label="Schema File"
        description="Select a file from the file manager containing the JSON schema."
        source="files"
        value={value.schema_path}
        onChange={(v) => onChange({ ...value, schema_path: v })}
      />
      <CodeField
        label="Inline Schema"
        description="A JSON schema to validate messages against. Used if no schema file is selected."
        language="json"
        value={value.schema}
        onChange={(v) => onChange({ ...value, schema: v })}
      />
    </div>
  );
}
