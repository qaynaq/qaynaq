import {
  TextField,
  NumberField,
  SelectField,
  CheckboxField,
  ArrayField,
  KeyValueField,
  CodeField,
  ConnectionPickerField,
} from "@/components/form-primitives";
import {
  BasicAuthEditor,
  OAuthEditor,
  OAuth2Editor,
  JwtEditor,
} from "../../shared/auth-editors";
import type {
  BasicAuth,
  OAuth,
  OAuth2,
  Jwt,
} from "../../shared/auth";
import type { EditorProps } from "../../types";

interface Meta {
  include_prefixes: string[];
  include_patterns: string[];
}
interface Config {
  url: string;
  verb: "GET" | "POST" | "PUT" | "DELETE" | "PATCH" | "HEAD";
  oauth_connection: string;
  headers: Record<string, string>;
  metadata: Meta;
  oauth: OAuth;
  oauth2: OAuth2;
  basic_auth: BasicAuth;
  jwt: Jwt;
  extract_headers: Meta;
  rate_limit: string;
  timeout: string;
  retry_period: string;
  max_retry_backoff: string;
  retries: number;
  backoff_on: number[];
  drop_on: number[];
  successful_on: number[];
  proxy_url: string;
  payload: string;
  batch_as_multipart: boolean;
  parallel: boolean;
}

export default function HttpProcessorEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <TextField label="URL" required value={value.url} onChange={(v) => set("url", v)} error={errors?.url} />
      <SelectField label="HTTP Verb" value={value.verb} onChange={(v) => set("verb", v as Config["verb"])} options={["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"]} />
      <ConnectionPickerField label="OAuth Connection" source="connections" value={value.oauth_connection} onChange={(v) => set("oauth_connection", v)} />
      <KeyValueField label="Headers" value={value.headers} onChange={(v) => set("headers", v)} />
      <CodeField label="Payload" description="Optional Bloblang mapping to create the request payload. Empty = raw message payload." value={value.payload} onChange={(v) => set("payload", v)} />
      <TextField label="Rate Limit" value={value.rate_limit} onChange={(v) => set("rate_limit", v)} />
      <TextField label="Timeout" value={value.timeout} onChange={(v) => set("timeout", v)} />
      <TextField label="Retry Period" value={value.retry_period} onChange={(v) => set("retry_period", v)} />
      <TextField label="Max Retry Backoff" value={value.max_retry_backoff} onChange={(v) => set("max_retry_backoff", v)} />
      <NumberField label="Retries" min={0} value={value.retries} onChange={(v) => set("retries", v)} />
      <ArrayField<number> label="Backoff On" itemType="number" value={value.backoff_on} onChange={(v) => set("backoff_on", v)} />
      <ArrayField<number> label="Drop On" itemType="number" value={value.drop_on} onChange={(v) => set("drop_on", v)} />
      <ArrayField<number> label="Successful On" itemType="number" value={value.successful_on} onChange={(v) => set("successful_on", v)} />
      <TextField label="Proxy URL" value={value.proxy_url} onChange={(v) => set("proxy_url", v)} />
      <CheckboxField label="Batch as Multipart" checked={value.batch_as_multipart} onChange={(c) => set("batch_as_multipart", c)} />
      <CheckboxField label="Parallel" description="When processing batched messages, send in parallel." checked={value.parallel} onChange={(c) => set("parallel", c)} />

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Metadata</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <ArrayField label="Include Prefixes" size="sm" value={value.metadata.include_prefixes} onChange={(v) => set("metadata", { ...value.metadata, include_prefixes: v })} />
          <ArrayField label="Include Patterns" size="sm" value={value.metadata.include_patterns} onChange={(v) => set("metadata", { ...value.metadata, include_patterns: v })} />
        </div>
      </section>

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Basic Auth</h4>
        <BasicAuthEditor value={value.basic_auth} onChange={(v) => set("basic_auth", v)} errors={errors} errorPathPrefix="basic_auth" />
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">OAuth</h4>
        <OAuthEditor value={value.oauth} onChange={(v) => set("oauth", v)} errors={errors} errorPathPrefix="oauth" />
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">OAuth2</h4>
        <OAuth2Editor value={value.oauth2} onChange={(v) => set("oauth2", v)} errors={errors} errorPathPrefix="oauth2" />
      </section>
      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">JWT</h4>
        <JwtEditor value={value.jwt} onChange={(v) => set("jwt", v)} errors={errors} errorPathPrefix="jwt" />
      </section>

      <section className="space-y-2 border-t pt-3">
        <h4 className="text-sm font-medium">Extract Headers</h4>
        <div className="space-y-2 pl-4 border-l-2 border-border">
          <ArrayField label="Include Prefixes" size="sm" value={value.extract_headers.include_prefixes} onChange={(v) => set("extract_headers", { ...value.extract_headers, include_prefixes: v })} />
          <ArrayField label="Include Patterns" size="sm" value={value.extract_headers.include_patterns} onChange={(v) => set("extract_headers", { ...value.extract_headers, include_patterns: v })} />
        </div>
      </section>
    </div>
  );
}
