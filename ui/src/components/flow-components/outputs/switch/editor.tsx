import { CheckboxField } from "@/components/form-primitives";
import {
  OutputCasesField,
  type OutputCase,
} from "../../shared/output-cases-field";
import type { EditorProps } from "../../types";

interface Config {
  retry_until_success: boolean;
  strict_mode: boolean;
  cases: OutputCase[];
}

export default function SwitchOutputEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <CheckboxField
        label="Retry Until Success"
        description="If a selected output fails to send a message, reattempt indefinitely."
        checked={value.retry_until_success}
        onChange={(c) => onChange({ ...value, retry_until_success: c })}
      />
      <CheckboxField
        label="Strict Mode"
        description="Report an error if no condition is met."
        checked={value.strict_mode}
        onChange={(c) => onChange({ ...value, strict_mode: c })}
      />
      <OutputCasesField
        label="Cases"
        description="A list of switch cases, outlining outputs that can be routed to."
        value={value.cases}
        onChange={(next) => onChange({ ...value, cases: next })}
      />
    </div>
  );
}
