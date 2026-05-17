import {
  TextField,
  NumberField,
  CheckboxField,
  SelectField,
  ArrayField,
} from "@/components/form-primitives";
import type { EditorProps } from "../../types";
import type { Config } from ".";

const EXCHANGE_TYPES = ["direct", "fanout", "topic", "x-custom"];

export default function Amqp09OutputEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  const setED = (next: Config["exchange_declare"]) =>
    set("exchange_declare", next);
  return (
    <div className="space-y-4">
      <ArrayField label="URLs" required value={value.urls} onChange={(v) => set("urls", v)} />
      <TextField label="Exchange" required value={value.exchange} onChange={(v) => set("exchange", v)} error={errors?.exchange} />
      <TextField label="Routing Key" value={value.key} onChange={(v) => set("key", v)} />
      <TextField label="Type" value={value.type} onChange={(v) => set("type", v)} />
      <TextField label="Content Type" value={value.content_type} onChange={(v) => set("content_type", v)} />
      <CheckboxField label="Persistent" description="Send messages with persistent delivery mode." checked={value.persistent} onChange={(c) => set("persistent", c)} />
      <NumberField label="Max In Flight" min={1} value={value.max_in_flight} onChange={(v) => set("max_in_flight", v)} />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Exchange Declare</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <CheckboxField label="Enabled" checked={value.exchange_declare.enabled} onChange={(c) => setED({ ...value.exchange_declare, enabled: c })} />
          <SelectField label="Type" size="sm" value={value.exchange_declare.type} onChange={(v) => setED({ ...value.exchange_declare, type: v as Config["exchange_declare"]["type"] })} options={EXCHANGE_TYPES} />
          <CheckboxField label="Durable" checked={value.exchange_declare.durable} onChange={(c) => setED({ ...value.exchange_declare, durable: c })} />
        </div>
      </section>
    </div>
  );
}
