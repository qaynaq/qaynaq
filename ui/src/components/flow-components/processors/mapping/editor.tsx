import { CodeField } from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  mapping: string;
}

export default function MappingEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <CodeField
      label="Mapping"
      description="A Bloblang mapping to apply to messages."
      required
      value={value.mapping}
      onChange={(v) => onChange({ mapping: v })}
      error={errors?.mapping}
    />
  );
}
