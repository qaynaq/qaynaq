import {
  CheckboxField,
  TextField,
  ArrayField,
  KeyValueField,
} from "@/components/form-primitives";
import type { BasicAuth, OAuth, OAuth2, Jwt } from "./auth";

interface Common<T> {
  value: T;
  onChange: (next: T) => void;
  errors?: Record<string, string>;
  errorPathPrefix?: string;
}

export function BasicAuthEditor({
  value,
  onChange,
  errors,
  errorPathPrefix = "basic_auth",
}: Common<BasicAuth>) {
  const err = (k: keyof BasicAuth) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <CheckboxField
        label="Enabled"
        checked={value.enabled}
        onChange={(c) => onChange({ ...value, enabled: c })}
      />
      <TextField
        label="Username"
        size="sm"
        value={value.username}
        onChange={(v) => onChange({ ...value, username: v })}
        error={err("username")}
      />
      <TextField
        label="Password"
        size="sm"
        type="password"
        value={value.password}
        onChange={(v) => onChange({ ...value, password: v })}
        error={err("password")}
      />
    </div>
  );
}

export function OAuthEditor({
  value,
  onChange,
  errors,
  errorPathPrefix = "oauth",
}: Common<OAuth>) {
  const err = (k: keyof OAuth) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <CheckboxField
        label="Enabled"
        checked={value.enabled}
        onChange={(c) => onChange({ ...value, enabled: c })}
      />
      <TextField
        label="Consumer Key"
        size="sm"
        value={value.consumer_key}
        onChange={(v) => onChange({ ...value, consumer_key: v })}
        error={err("consumer_key")}
      />
      <TextField
        label="Consumer Secret"
        size="sm"
        type="password"
        value={value.consumer_secret}
        onChange={(v) => onChange({ ...value, consumer_secret: v })}
        error={err("consumer_secret")}
      />
      <TextField
        label="Access Token"
        size="sm"
        value={value.access_token}
        onChange={(v) => onChange({ ...value, access_token: v })}
      />
      <TextField
        label="Access Token Secret"
        size="sm"
        type="password"
        value={value.access_token_secret}
        onChange={(v) => onChange({ ...value, access_token_secret: v })}
      />
    </div>
  );
}

export function OAuth2Editor({
  value,
  onChange,
  errors,
  errorPathPrefix = "oauth2",
}: Common<OAuth2>) {
  const err = (k: keyof OAuth2) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <CheckboxField
        label="Enabled"
        checked={value.enabled}
        onChange={(c) => onChange({ ...value, enabled: c })}
      />
      <TextField
        label="Client Key"
        size="sm"
        value={value.client_key}
        onChange={(v) => onChange({ ...value, client_key: v })}
        error={err("client_key")}
      />
      <TextField
        label="Client Secret"
        size="sm"
        type="password"
        value={value.client_secret}
        onChange={(v) => onChange({ ...value, client_secret: v })}
      />
      <TextField
        label="Token URL"
        size="sm"
        value={value.token_url}
        onChange={(v) => onChange({ ...value, token_url: v })}
      />
      <ArrayField
        label="Scopes"
        size="sm"
        value={value.scopes}
        onChange={(v) => onChange({ ...value, scopes: v })}
      />
      <KeyValueField
        label="Endpoint Parameters"
        size="sm"
        value={value.endpoint_params}
        onChange={(v) => onChange({ ...value, endpoint_params: v })}
      />
    </div>
  );
}

export function JwtEditor({
  value,
  onChange,
  errors,
  errorPathPrefix = "jwt",
}: Common<Jwt>) {
  const err = (k: keyof Jwt) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <CheckboxField
        label="Enabled"
        checked={value.enabled}
        onChange={(c) => onChange({ ...value, enabled: c })}
      />
      <TextField
        label="Private Key File"
        size="sm"
        value={value.private_key_file}
        onChange={(v) => onChange({ ...value, private_key_file: v })}
        error={err("private_key_file")}
      />
      <TextField
        label="Signing Method"
        size="sm"
        value={value.signing_method}
        onChange={(v) => onChange({ ...value, signing_method: v })}
      />
      <KeyValueField
        label="Claims"
        size="sm"
        value={value.claims}
        onChange={(v) => onChange({ ...value, claims: v })}
      />
      <KeyValueField
        label="Headers"
        size="sm"
        value={value.headers}
        onChange={(v) => onChange({ ...value, headers: v })}
      />
    </div>
  );
}
