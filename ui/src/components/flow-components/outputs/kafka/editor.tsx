import {
  TextField,
  NumberField,
  CheckboxField,
  SelectField,
  ArrayField,
} from "@/components/form-primitives";
import { TlsEditor } from "../../shared/tls-editor";
import type { EditorProps } from "../../types";
import type { Config } from ".";

const SASL = ["none", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"];
const COMPRESSION = ["none", "gzip", "snappy", "lz4", "zstd"];

type Sasl = Config["sasl"];

export default function KafkaOutputEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <ArrayField label="Addresses" value={value.addresses} onChange={(v) => set("addresses", v)} />
      <TextField label="Topic" value={value.topic} onChange={(v) => set("topic", v)} />
      <TextField label="Key" description="Optional key for each message (supports interpolation)." value={value.key} onChange={(v) => set("key", v)} />
      <TextField label="Client ID" value={value.client_id} onChange={(v) => set("client_id", v)} />
      <NumberField label="Max In Flight" min={1} value={value.max_in_flight} onChange={(v) => set("max_in_flight", v)} />
      <CheckboxField label="Ack Replicas" description="Wait for all replicas to acknowledge messages." checked={value.ack_replicas} onChange={(c) => set("ack_replicas", c)} />
      <SelectField label="Compression" value={value.compression} onChange={(v) => set("compression", v as Config["compression"])} options={COMPRESSION} />
      <NumberField label="Max Message Bytes" min={1} value={value.max_message_bytes} onChange={(v) => set("max_message_bytes", v)} />
      <TextField label="Target Version" value={value.target_version} onChange={(v) => set("target_version", v)} />
      <TextField label="Timeout" value={value.timeout} onChange={(v) => set("timeout", v)} />

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Metadata</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <ArrayField label="Include Prefixes" size="sm" value={value.metadata.include_prefixes} onChange={(v) => set("metadata", { ...value.metadata, include_prefixes: v })} />
          <ArrayField label="Include Patterns" size="sm" value={value.metadata.include_patterns} onChange={(v) => set("metadata", { ...value.metadata, include_patterns: v })} />
        </div>
      </section>

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">SASL</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <CheckboxField label="Enabled" checked={value.sasl.enabled} onChange={(c) => set("sasl", { ...value.sasl, enabled: c })} />
          <SelectField label="Mechanism" size="sm" value={value.sasl.mechanism} onChange={(v) => set("sasl", { ...value.sasl, mechanism: v as Sasl["mechanism"] })} options={SASL} />
          <TextField label="Username" size="sm" value={value.sasl.username} onChange={(v) => set("sasl", { ...value.sasl, username: v })} />
          <TextField label="Password" size="sm" type="password" value={value.sasl.password} onChange={(v) => set("sasl", { ...value.sasl, password: v })} />
        </div>
      </section>

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">TLS</h4>
        <TlsEditor value={value.tls} onChange={(v) => set("tls", v)} errors={errors} errorPathPrefix="tls" />
      </section>

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Batching</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <NumberField label="Count" size="sm" min={0} value={value.batching.count} onChange={(v) => set("batching", { ...value.batching, count: v })} />
          <NumberField label="Byte Size" size="sm" min={0} value={value.batching.byte_size} onChange={(v) => set("batching", { ...value.batching, byte_size: v })} />
          <TextField label="Period" size="sm" value={value.batching.period} onChange={(v) => set("batching", { ...value.batching, period: v })} />
          <NumberField label="Jitter" size="sm" min={0} step={0.1} value={value.batching.jitter} onChange={(v) => set("batching", { ...value.batching, jitter: v })} />
          <TextField label="Check" size="sm" value={value.batching.check} onChange={(v) => set("batching", { ...value.batching, check: v })} />
        </div>
      </section>
    </div>
  );
}
