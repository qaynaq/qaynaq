import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export interface McpAnnotationsValue {
  title?: string;
  read_only_hint?: boolean;
  destructive_hint?: boolean;
  idempotent_hint?: boolean;
  open_world_hint?: boolean;
}

// Schema: https://modelcontextprotocol.io/specification/2025-11-25/schema#toolannotations
const SPEC_DEFAULTS = {
  read_only_hint: false,
  destructive_hint: true,
  idempotent_hint: false,
  open_world_hint: true,
} as const;

interface Props {
  value: McpAnnotationsValue | undefined;
  updateValue: (value: McpAnnotationsValue | undefined) => void;
  previewMode?: boolean;
}

export function McpAnnotationsEditor({ value, updateValue, previewMode = false }: Props) {
  const current: McpAnnotationsValue = value ?? {};

  const readOnly = current.read_only_hint ?? SPEC_DEFAULTS.read_only_hint;
  const openWorld = current.open_world_hint ?? SPEC_DEFAULTS.open_world_hint;
  const destructive = readOnly ? false : (current.destructive_hint ?? false);
  const idempotent = readOnly ? false : (current.idempotent_hint ?? false);

  const setHint = (key: keyof typeof SPEC_DEFAULTS, next: boolean) => {
    const merged: McpAnnotationsValue = { ...current, [key]: next };
    if (key === "read_only_hint") {
      merged.destructive_hint = false;
      merged.idempotent_hint = false;
    }
    emit(merged);
  };

  const setTitle = (next: string) => {
    const merged: McpAnnotationsValue = { ...current };
    const trimmed = next.trim();
    if (trimmed === "") {
      delete merged.title;
    } else {
      merged.title = next;
    }
    emit(merged);
  };

  const emit = (next: McpAnnotationsValue) => {
    updateValue(Object.keys(next).length === 0 ? undefined : next);
  };

  return (
    <div className="space-y-3">
      <div className="space-y-1">
        <Label className="text-xs font-normal text-muted-foreground">Display Title</Label>
        <Input
          value={current.title ?? ""}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Human-readable title (e.g. Customer Lookup)"
          className="h-8 text-sm"
          disabled={previewMode}
        />
      </div>

      <HintRow
        label="Read-only"
        description="Tool does not modify its environment"
        checked={readOnly}
        onChange={(next) => setHint("read_only_hint", next)}
        previewMode={previewMode}
      />

      <HintRow
        label="Destructive"
        description="May perform destructive updates (only meaningful when not read-only)"
        checked={destructive}
        onChange={(next) => setHint("destructive_hint", next)}
        previewMode={previewMode}
        disabled={readOnly}
      />

      <HintRow
        label="Idempotent"
        description="Repeated calls with the same arguments have no additional effect"
        checked={idempotent}
        onChange={(next) => setHint("idempotent_hint", next)}
        previewMode={previewMode}
        disabled={readOnly}
      />

      <HintRow
        label="Open world"
        description="Interacts with external entities (e.g. web search). Disable for closed-domain tools"
        checked={openWorld}
        onChange={(next) => setHint("open_world_hint", next)}
        previewMode={previewMode}
      />
    </div>
  );
}

function HintRow({
  label,
  description,
  checked,
  onChange,
  previewMode,
  disabled,
}: {
  label: string;
  description: string;
  checked: boolean;
  onChange: (next: boolean) => void;
  previewMode: boolean;
  disabled?: boolean;
}) {
  const isDisabled = previewMode || disabled;
  return (
    <label
      className={`flex items-start gap-2 text-sm ${isDisabled ? "cursor-default" : "cursor-pointer"}`}
    >
      <Checkbox
        checked={checked}
        onCheckedChange={(next) => onChange(!!next)}
        disabled={isDisabled}
        className="mt-0.5 h-3.5 w-3.5"
      />
      <span>
        <span className={`font-medium ${disabled ? "text-muted-foreground" : ""}`}>{label}</span>
        <span className="block text-xs text-muted-foreground">{description}</span>
      </span>
    </label>
  );
}
