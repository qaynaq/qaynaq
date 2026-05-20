import { CheckboxField, SelectField } from "@/components/form-primitives";
import {
  HugotModelFields,
  type HugotModelConfig,
} from "../../shared/hugot-model-fields";
import type { EditorProps } from "../../types";
import {
  AGGREGATION_FUNCTIONS,
  type AggregationFunction,
  type Config,
} from ".";

export default function NlpClassifyTextEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setModel = (next: HugotModelConfig) => onChange({ ...value, ...next });
  return (
    <div className="space-y-4">
      <HugotModelFields value={value} onChange={setModel} errors={errors} />
      <SelectField
        label="Aggregation Function"
        description="How raw model logits are converted into probabilities."
        value={value.aggregation_function}
        onChange={(v) =>
          onChange({
            ...value,
            aggregation_function: v as AggregationFunction,
          })
        }
        options={AGGREGATION_FUNCTIONS as unknown as string[]}
      />
      <CheckboxField
        label="Multi-Label"
        description="Return a score for every label. When off, only the top label is emitted."
        checked={value.multi_label}
        onChange={(c) => onChange({ ...value, multi_label: c })}
      />
    </div>
  );
}
