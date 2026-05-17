import {
  TextField,
  NumberField,
  CheckboxField,
  SelectField,
  ArrayField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";
import type { Config } from ".";

const SASL = ["none", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"];
type Sasl = Config["sasl"];

export default function KafkaInputEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <ArrayField label="Addresses" description="Broker addresses." value={value.addresses} onChange={(v) => set("addresses", v)} />
      <ArrayField label="Topics" value={value.topics} onChange={(v) => set("topics", v)} />
      <TextField label="Consumer Group" value={value.consumer_group} onChange={(v) => set("consumer_group", v)} error={errors?.consumer_group} />
      <TextField label="Client ID" value={value.client_id} onChange={(v) => set("client_id", v)} />
      <TextField label="Rack ID" value={value.rack_id} onChange={(v) => set("rack_id", v)} />
      <CheckboxField label="Start From Oldest" checked={value.start_from_oldest} onChange={(c) => set("start_from_oldest", c)} />
      <NumberField label="Checkpoint Limit" min={1} value={value.checkpoint_limit} onChange={(v) => set("checkpoint_limit", v)} />
      <TextField label="Commit Period" value={value.commit_period} onChange={(v) => set("commit_period", v)} />
      <TextField label="Max Processing Period" value={value.max_processing_period} onChange={(v) => set("max_processing_period", v)} />
      <TextField label="Target Version" value={value.target_version} onChange={(v) => set("target_version", v)} />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">SASL</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <SelectField label="Mechanism" size="sm" value={value.sasl.mechanism} onChange={(v) => set("sasl", { ...value.sasl, mechanism: v as Sasl["mechanism"] })} options={SASL} />
          <TextField label="Username" size="sm" value={value.sasl.user} onChange={(v) => set("sasl", { ...value.sasl, user: v })} />
          <TextField label="Password" size="sm" type="password" value={value.sasl.password} onChange={(v) => set("sasl", { ...value.sasl, password: v })} />
        </div>
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Batching</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <NumberField label="Count" size="sm" min={0} value={value.batching.count} onChange={(v) => set("batching", { ...value.batching, count: v })} />
          <NumberField label="Byte Size" size="sm" min={0} value={value.batching.byte_size} onChange={(v) => set("batching", { ...value.batching, byte_size: v })} />
          <TextField label="Period" size="sm" value={value.batching.period} onChange={(v) => set("batching", { ...value.batching, period: v })} />
          <TextField label="Check" size="sm" value={value.batching.check} onChange={(v) => set("batching", { ...value.batching, check: v })} />
        </div>
      </section>
    </div>
  );
}
