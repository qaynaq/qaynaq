import { TextField, CodeField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  name: string;
  args_mapping: string;
}

export default function CommandEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <TextField
        label="Name"
        description="The name of the command to execute. Supports interpolation functions."
        required
        value={value.name}
        onChange={(v) => onChange({ ...value, name: v })}
        error={errors?.name}
      />
      <CodeField
        label="Args Mapping"
        description='A Bloblang mapping that should evaluate to an array of strings to use as command arguments (e.g. [ "-c", this.script_path ]).'
        value={value.args_mapping}
        onChange={(v) => onChange({ ...value, args_mapping: v })}
        error={errors?.args_mapping}
      />
    </div>
  );
}
