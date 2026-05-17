import { CodeField } from "@/components/form-primitives";
import { ComponentListField } from "../../shared/component-list-field";
import type { EditorProps } from "../../types";
import type { ListItem } from "../../utils/list-items";

interface Config {
  request_map: string;
  processors: ListItem[];
  result_map: string;
}

export default function BranchEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <CodeField
        label="Request Map"
        description="A Bloblang mapping that transforms the message before sending it to the branch processors. Empty = original message is used."
        value={value.request_map}
        onChange={(v) => onChange({ ...value, request_map: v })}
      />
      <ComponentListField
        label="Processors"
        description="A list of processors to execute on the mapped request."
        category="processor"
        value={value.processors}
        onChange={(next) => onChange({ ...value, processors: next })}
      />
      <CodeField
        label="Result Map"
        description="A Bloblang mapping that merges the branch result back into the original message. In this mapping 'this' = branch result, 'root' = original message."
        value={value.result_map}
        onChange={(v) => onChange({ ...value, result_map: v })}
      />
    </div>
  );
}
