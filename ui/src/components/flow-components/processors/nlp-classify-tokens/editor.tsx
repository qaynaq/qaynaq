import { ArrayField, SelectField } from "@/components/form-primitives";
import {
  HugotModelFields,
  type HugotModelConfig,
} from "../../shared/hugot-model-fields";
import type { EditorProps } from "../../types";
import {
  AGGREGATION_STRATEGIES,
  type AggregationStrategy,
  type Config,
} from ".";

export default function NlpClassifyTokensEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setModel = (next: HugotModelConfig) => onChange({ ...value, ...next });
  return (
    <div className="space-y-4">
      <HugotModelFields value={value} onChange={setModel} errors={errors} />
      <SelectField
        label="Aggregation Strategy"
        description="SIMPLE merges adjacent tokens that share the same entity label. NONE returns every individual token."
        value={value.aggregation_strategy}
        onChange={(v) =>
          onChange({
            ...value,
            aggregation_strategy: v as AggregationStrategy,
          })
        }
        options={AGGREGATION_STRATEGIES as unknown as string[]}
      />
      <ArrayField
        label="Ignore Labels"
        description="Labels to drop from the output. Common values are 'O' (outside, the BIO scheme's non-entity tag) and 'MISC'."
        value={value.ignore_labels}
        onChange={(v) => onChange({ ...value, ignore_labels: v })}
        placeholder="O"
      />
    </div>
  );
}
