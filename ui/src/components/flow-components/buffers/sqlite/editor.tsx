import { TextField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  path: string;
}

export default function SqliteBufferEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <TextField
      label="Path"
      description="The path of the database file, which will be created if it does not already exist."
      required
      value={value.path}
      onChange={(v) => onChange({ ...value, path: v })}
      error={errors?.path}
    />
  );
}
