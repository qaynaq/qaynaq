import { TextField, SelectField } from "@/components/form-primitives";
import { TlsEditor } from "../../shared/tls-editor";
import { RetriesEditor } from "../../shared/retries-editor";
import type { Tls } from "../../shared/tls";
import type { Retries } from "../../shared/retries";
import type { EditorProps } from "../../types";

interface Config {
  url: string;
  kind: "simple" | "cluster" | "failover";
  master: string;
  tls: Tls;
  prefix: string;
  default_ttl: string;
  retries: Retries;
}

export default function RedisCacheEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  return (
    <div className="space-y-4">
      <TextField
        label="URL"
        description="The URL of the target Redis server. Database is optional and is supplied as the URL path."
        required
        value={value.url}
        onChange={(v) => onChange({ ...value, url: v })}
        error={errors?.url}
      />
      <SelectField
        label="Kind"
        description="Specifies a simple, cluster-aware, or failover-aware Redis client."
        value={value.kind}
        onChange={(v) =>
          onChange({ ...value, kind: v as Config["kind"] })
        }
        options={["simple", "cluster", "failover"]}
      />
      <TextField
        label="Master"
        description="Name of the redis master when kind is failover."
        value={value.master}
        onChange={(v) => onChange({ ...value, master: v })}
      />
      <TextField
        label="Prefix"
        description="An optional string to prefix item keys with in order to prevent collisions with similar services."
        value={value.prefix}
        onChange={(v) => onChange({ ...value, prefix: v })}
      />
      <TextField
        label="Default TTL"
        description="An optional default TTL to set for items, calculated from the moment the item is cached."
        value={value.default_ttl}
        onChange={(v) => onChange({ ...value, default_ttl: v })}
      />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">TLS</h4>
        <TlsEditor
          value={value.tls}
          onChange={(t) => onChange({ ...value, tls: t })}
          errors={errors}
          errorPathPrefix="tls"
        />
      </section>
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
