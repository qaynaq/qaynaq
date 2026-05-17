import {
  CodeField,
  TextField,
  NumberField,
  CheckboxField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  mapping: string;
  interval: string;
  count: number;
  batch_size: number;
  auto_replay_nacks: boolean;
}

export default function GenerateEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <CodeField
        label="Mapping"
        description="A Bloblang mapping to use for generating messages."
        required
        value={value.mapping}
        onChange={(v) => onChange({ ...value, mapping: v })}
        error={errors?.mapping}
      />
      <TextField
        label="Interval"
        description="Time interval at which messages should be generated (e.g. 1s, 1m, @every 1s)."
        value={value.interval}
        onChange={(v) => onChange({ ...value, interval: v })}
        error={errors?.interval}
      />
      <NumberField
        label="Count"
        description="Optional number of messages to generate; the input shuts down after. 0 means unlimited."
        min={0}
        value={value.count}
        onChange={(v) => onChange({ ...value, count: v })}
        error={errors?.count}
      />
      <NumberField
        label="Batch Size"
        description="Number of generated messages accumulated into each batch."
        min={1}
        value={value.batch_size}
        onChange={(v) => onChange({ ...value, batch_size: v })}
        error={errors?.batch_size}
      />
      <CheckboxField
        label="Auto Replay Nacks"
        description="Whether messages rejected at the output level should be automatically replayed."
        checked={value.auto_replay_nacks}
        onChange={(c) => onChange({ ...value, auto_replay_nacks: c })}
      />
    </div>
  );
}
