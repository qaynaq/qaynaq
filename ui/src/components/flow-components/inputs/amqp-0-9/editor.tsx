import {
  TextField,
  NumberField,
  CheckboxField,
  ArrayField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";

interface Config {
  urls: string[];
  queue: string;
  consumer_tag: string;
  auto_ack: boolean;
  prefetch_count: number;
  nack_reject_patterns: string[];
  queue_declare: { enabled: boolean; durable: boolean; auto_delete: boolean };
}

export default function Amqp09Editor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setQD = (next: Config["queue_declare"]) =>
    onChange({ ...value, queue_declare: next });
  return (
    <div className="space-y-4">
      <ArrayField label="URLs" description="A list of URLs; the first to connect is used." required value={value.urls} onChange={(v) => onChange({ ...value, urls: v })} />
      <TextField label="Queue" required value={value.queue} onChange={(v) => onChange({ ...value, queue: v })} error={errors?.queue} />
      <TextField label="Consumer Tag" value={value.consumer_tag} onChange={(v) => onChange({ ...value, consumer_tag: v })} />
      <CheckboxField label="Auto Ack" description="Acknowledge on receipt, skipping downstream ack." checked={value.auto_ack} onChange={(c) => onChange({ ...value, auto_ack: c })} />
      <NumberField label="Prefetch Count" min={0} value={value.prefetch_count} onChange={(v) => onChange({ ...value, prefetch_count: v })} />
      <ArrayField label="Nack Reject Patterns" description="Regex patterns; matching failures are dropped instead of requeued." value={value.nack_reject_patterns} onChange={(v) => onChange({ ...value, nack_reject_patterns: v })} />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Queue Declare</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <CheckboxField label="Enabled" checked={value.queue_declare.enabled} onChange={(c) => setQD({ ...value.queue_declare, enabled: c })} />
          <CheckboxField label="Durable" checked={value.queue_declare.durable} onChange={(c) => setQD({ ...value.queue_declare, durable: c })} />
          <CheckboxField label="Auto Delete" checked={value.queue_declare.auto_delete} onChange={(c) => setQD({ ...value.queue_declare, auto_delete: c })} />
        </div>
      </section>
    </div>
  );
}
