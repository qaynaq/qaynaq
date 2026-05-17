import { TextField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  directory: string;
}

export default function FileCacheEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <TextField
      label="Directory"
      description="The directory within which to store items as files."
      required
      value={value.directory}
      onChange={(v) => onChange({ ...value, directory: v })}
      error={errors?.directory}
    />
  );
}
