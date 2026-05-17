import { TextField, ArrayField } from "@/components/form-primitives";
import { RetriesEditor } from "../../shared/retries-editor";
import type { Retries } from "../../shared/retries";
import type { EditorProps } from "../../types";

interface Config {
  addresses: string[];
  prefix: string;
  default_ttl: string;
  retries: Retries;
}

export default function MemcachedEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <ArrayField
        label="Addresses"
        description="A list of addresses of memcached servers to use."
        required
        value={value.addresses}
        onChange={(v) => onChange({ ...value, addresses: v })}
      />
      <TextField
        label="Prefix"
        description="An optional string to prefix item keys with."
        value={value.prefix}
        onChange={(v) => onChange({ ...value, prefix: v })}
      />
      <TextField
        label="Default TTL"
        description="A default TTL to set for items."
        value={value.default_ttl}
        onChange={(v) => onChange({ ...value, default_ttl: v })}
      />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Retries</h4>
        <RetriesEditor
          value={value.retries}
          onChange={(r) => onChange({ ...value, retries: r })}
          errors={errors}
          errorPathPrefix="retries"
        />
      </section>
    </div>
  );
}
