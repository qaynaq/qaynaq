import { useEffect, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  fetchCaches,
  fetchConnections,
  fetchSecrets,
  fetchRateLimits,
  fetchFiles,
} from "@/lib/api";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

export type PickerSource =
  | "caches"
  | "connections"
  | "secrets"
  | "rate_limits"
  | "files";

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: string;
  onChange: (next: string) => void;
  source: PickerSource;
  disabled?: boolean;
}

const SECRET_RE = /^\$\{(.+)\}$/;
const CONN_RE = /^\$\{QAYNAQ_CONN_(.+)\}$/;
const FILE_PREFIX = "qaynaq://";

function unwrap(raw: string, source: PickerSource): string {
  if (!raw) return "";
  if (source === "secrets") return raw.replace(SECRET_RE, "$1");
  if (source === "connections") return raw.replace(CONN_RE, "$1");
  if (source === "files") return raw.startsWith(FILE_PREFIX) ? raw.slice(FILE_PREFIX.length) : raw;
  return raw;
}

function wrap(name: string, source: PickerSource): string {
  if (!name) return "";
  if (source === "secrets") return `\${${name}}`;
  if (source === "connections") return `\${QAYNAQ_CONN_${name}}`;
  if (source === "files") return `${FILE_PREFIX}${name}`;
  return name;
}

async function loadOptions(source: PickerSource): Promise<string[]> {
  if (source === "caches") return (await fetchCaches()).map((c) => c.label);
  if (source === "secrets") return (await fetchSecrets()).map((s) => s.key);
  if (source === "connections") return (await fetchConnections()).map((c) => c.name);
  if (source === "rate_limits") return (await fetchRateLimits()).map((r) => r.label);
  if (source === "files") return (await fetchFiles()).map((f) => f.key);
  return [];
}

export function ConnectionPickerField({
  value,
  onChange,
  source,
  disabled,
  size,
  ...wrapper
}: Props) {
  const [options, setOptions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    loadOptions(source)
      .then((opts) => {
        if (!cancelled) setOptions(opts);
      })
      .catch(() => {
        if (!cancelled) setOptions([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [source]);

  if (loading) {
    return (
      <FieldWrapper size={size} {...wrapper}>
        <div className="text-sm text-muted-foreground">Loading...</div>
      </FieldWrapper>
    );
  }

  const displayValue = unwrap(value, source);

  return (
    <FieldWrapper size={size} {...wrapper}>
      <Select
        value={displayValue}
        onValueChange={(v) => onChange(wrap(v, source))}
        disabled={disabled}
      >
        <SelectTrigger className={size === "sm" ? "h-8 text-sm" : "h-9 text-sm"}>
          <SelectValue
            placeholder={
              options.length === 0 ? `No ${source.replace("_", " ")} available` : "Select..."
            }
          />
        </SelectTrigger>
        <SelectContent>
          {options.length === 0 ? (
            <div className="px-2 py-1.5 text-sm text-muted-foreground">
              No {source.replace("_", " ")} available
            </div>
          ) : (
            options.map((opt) => (
              <SelectItem key={opt} value={opt}>
                {opt}
              </SelectItem>
            ))
          )}
        </SelectContent>
      </Select>
    </FieldWrapper>
  );
}
