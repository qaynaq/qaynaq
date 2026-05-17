import { ProcessorCasesField } from "../../shared/processor-cases-field";
import type { EditorProps } from "../../types";
import type { ProcessorCase } from "../../shared/processor-cases-field";

interface Config {
  cases: ProcessorCase[];
}

export default function SwitchProcessorEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  return (
    <ProcessorCasesField
      label="Cases"
      description="A list of switch cases. The first matching case runs unless fallthrough is set."
      value={value.cases}
      onChange={(next) => onChange({ cases: next })}
    />
  );
}
