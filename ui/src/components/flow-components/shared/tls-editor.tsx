import { CheckboxField, TextField } from "@/components/form-primitives";
import type { Tls } from "./tls";

interface Props {
  value: Tls;
  onChange: (next: Tls) => void;
  errors?: Record<string, string>;
  errorPathPrefix?: string;
}

export function TlsEditor({
  value,
  onChange,
  errors,
  errorPathPrefix = "tls",
}: Props) {
  const err = (k: keyof Tls) => errors?.[`${errorPathPrefix}.${k}`];
  return (
    <div className="space-y-3 pl-4 border-l-2 border-border">
      <CheckboxField
        label="Enabled"
        checked={value.enabled}
        onChange={(c) => onChange({ ...value, enabled: c })}
        error={err("enabled")}
      />
      <CheckboxField
        label="Skip Certificate Verification"
        description="Whether to skip server side certificate verification."
        checked={value.skip_cert_verify}
        onChange={(c) => onChange({ ...value, skip_cert_verify: c })}
      />
      <CheckboxField
        label="Enable Renegotiation"
        description="Whether to allow the remote server to repeatedly request renegotiation."
        checked={value.enable_renegotiation}
        onChange={(c) => onChange({ ...value, enable_renegotiation: c })}
      />
      <TextField
        label="Root CAs"
        size="sm"
        value={value.root_cas}
        onChange={(v) => onChange({ ...value, root_cas: v })}
      />
      <TextField
        label="Root CAs File"
        size="sm"
        value={value.root_cas_file}
        onChange={(v) => onChange({ ...value, root_cas_file: v })}
      />
    </div>
  );
}
