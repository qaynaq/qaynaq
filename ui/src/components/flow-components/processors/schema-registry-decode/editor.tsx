import { TextField, CheckboxField } from "@/components/form-primitives";
import {
  BasicAuthEditor,
  OAuthEditor,
  JwtEditor,
} from "../../shared/auth-editors";
import { TlsEditor } from "../../shared/tls-editor";
import type { BasicAuth, OAuth, Jwt } from "../../shared/auth";
import type { Tls } from "../../shared/tls";
import type { EditorProps } from "../../types";

interface Config {
  url: string;
  avro_raw_json: boolean;
  avro_nested_schemas: boolean;
  oauth: OAuth;
  basic_auth: BasicAuth;
  jwt: Jwt;
  tls: Tls;
}

export default function SchemaRegistryDecodeEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <TextField label="URL" description="The base URL of the schema registry service." required value={value.url} onChange={(v) => set("url", v)} error={errors?.url} />
      <CheckboxField label="Avro Raw JSON" description="Decode Avro messages as normal JSON rather than Avro JSON." checked={value.avro_raw_json} onChange={(c) => set("avro_raw_json", c)} />
      <CheckboxField label="Avro Nested Schemas" description="Resolve nested Avro schema references." checked={value.avro_nested_schemas} onChange={(c) => set("avro_nested_schemas", c)} />
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Basic Auth</h4>
        <BasicAuthEditor value={value.basic_auth} onChange={(v) => set("basic_auth", v)} errors={errors} errorPathPrefix="basic_auth" />
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">OAuth</h4>
        <OAuthEditor value={value.oauth} onChange={(v) => set("oauth", v)} errors={errors} errorPathPrefix="oauth" />
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">JWT</h4>
        <JwtEditor value={value.jwt} onChange={(v) => set("jwt", v)} errors={errors} errorPathPrefix="jwt" />
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">TLS</h4>
        <TlsEditor value={value.tls} onChange={(v) => set("tls", v)} errors={errors} errorPathPrefix="tls" />
      </section>
    </div>
  );
}
