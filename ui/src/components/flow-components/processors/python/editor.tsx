import { CodeField, ArrayField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  script: string;
  imports: string[];
}

export default function PythonEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <CodeField
        label="Script"
        description="The Python script to execute. The incoming message is available as 'this'; assign the transformed result to 'root'."
        required
        language="python"
        height="200px"
        value={value.script}
        onChange={(v) => onChange({ ...value, script: v })}
        error={errors?.script}
      />
      <ArrayField
        label="Imports"
        description="An optional list of Python modules to pre-import for the script."
        value={value.imports}
        onChange={(v) => onChange({ ...value, imports: v })}
      />
    </div>
  );
}
