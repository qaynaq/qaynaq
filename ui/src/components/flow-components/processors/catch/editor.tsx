import { ComponentListField } from "../../shared/component-list-field";
import type { EditorProps } from "../../types";
import type { ListItem } from "../../utils/list-items";

interface Config {
  processors: ListItem[];
}

export default function CatchEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  return (
    <ComponentListField
      label="Processors"
      description="A list of processors to apply when a message fails processing."
      category="processor"
      value={value.processors}
      onChange={(next) => onChange({ processors: next })}
    />
  );
}
