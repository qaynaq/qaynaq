import {
  NumberField,
  TextField,
} from "@/components/form-primitives";
import { ComponentListField } from "../../shared/component-list-field";
import type { EditorProps } from "../../types";
import type { ListItem } from "../../utils/list-items";

interface Batching {
  count: number;
  byte_size: number;
  period: string;
  jitter: number;
  check: string;
  processors: ListItem[];
}

interface Config {
  copies: number;
  inputs: ListItem[];
  batching: Batching;
}

export default function BrokerInputEditor({
  value,
  onChange,
}: EditorProps<Config>) {
  const setB = (next: Batching) => onChange({ ...value, batching: next });
  return (
    <div className="space-y-4">
      <NumberField
        label="Copies"
        description="The number of copies of each configured input to spawn."
        min={1}
        value={value.copies}
        onChange={(v) => onChange({ ...value, copies: v })}
      />
      <ComponentListField
        label="Inputs"
        description="A list of child inputs to broker."
        category="input"
        value={value.inputs}
        onChange={(next) => onChange({ ...value, inputs: next })}
      />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Batching</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <NumberField label="Count" size="sm" min={0} value={value.batching.count} onChange={(v) => setB({ ...value.batching, count: v })} />
          <NumberField label="Byte Size" size="sm" min={0} value={value.batching.byte_size} onChange={(v) => setB({ ...value.batching, byte_size: v })} />
          <TextField label="Period" size="sm" value={value.batching.period} onChange={(v) => setB({ ...value.batching, period: v })} />
          <NumberField label="Jitter" size="sm" min={0} step={0.1} value={value.batching.jitter} onChange={(v) => setB({ ...value.batching, jitter: v })} />
          <TextField label="Check" size="sm" value={value.batching.check} onChange={(v) => setB({ ...value.batching, check: v })} />
          <ComponentListField
            label="Processors"
            category="processor"
            value={value.batching.processors}
            onChange={(next) => setB({ ...value.batching, processors: next })}
          />
        </div>
      </section>
    </div>
  );
}
