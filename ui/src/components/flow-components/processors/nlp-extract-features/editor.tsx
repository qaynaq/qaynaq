import { CheckboxField } from "@/components/form-primitives";
import {
  HugotModelFields,
  type HugotModelConfig,
} from "../../shared/hugot-model-fields";
import type { EditorProps } from "../../types";
import type { Config } from ".";

export default function NlpExtractFeaturesEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setModel = (next: HugotModelConfig) => onChange({ ...value, ...next });
  return (
    <div className="space-y-4">
      <HugotModelFields value={value} onChange={setModel} errors={errors} />
      <CheckboxField
        label="Normalization"
        description="L2-normalize the output vector. Required by most cosine-similarity vector stores."
        checked={value.normalization}
        onChange={(c) => onChange({ ...value, normalization: c })}
      />
    </div>
  );
}
