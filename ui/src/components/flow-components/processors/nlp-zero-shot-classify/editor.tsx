import {
  ArrayField,
  CheckboxField,
  TextField,
} from "@/components/form-primitives";
import {
  HugotModelFields,
  type HugotModelConfig,
} from "../../shared/hugot-model-fields";
import type { EditorProps } from "../../types";
import type { Config } from ".";

export default function NlpZeroShotClassifyEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setModel = (next: HugotModelConfig) => onChange({ ...value, ...next });
  return (
    <div className="space-y-4">
      <HugotModelFields value={value} onChange={setModel} errors={errors} />
      <ArrayField
        label="Labels"
        description="Candidate labels the input will be classified against. At least one is required."
        required
        value={value.labels}
        onChange={(v) => onChange({ ...value, labels: v })}
        error={errors?.labels}
        placeholder="positive"
      />
      <CheckboxField
        label="Multi-Label"
        description="Score each label independently. Off means scores sum to 1 across labels."
        checked={value.multi_label}
        onChange={(c) => onChange({ ...value, multi_label: c })}
      />
      <TextField
        label="Hypothesis Template"
        description="Template used to turn each label into an NLI hypothesis. Must contain {} where the label is inserted."
        required
        value={value.hypothesis_template}
        onChange={(v) => onChange({ ...value, hypothesis_template: v })}
        error={errors?.hypothesis_template}
        placeholder="This example is {}."
      />
    </div>
  );
}
