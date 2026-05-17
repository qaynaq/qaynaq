import {
  TextField,
  TextAreaField,
  CheckboxField,
} from "@/components/form-primitives";
import { ParameterListField } from "./parameter-list";
import type { EditorProps } from "../../types";

interface Param {
  name: string;
  type: string;
  required: boolean;
  description: string;
}

interface Annotations {
  title?: string;
  read_only_hint: boolean;
  destructive_hint: boolean;
  idempotent_hint: boolean;
  open_world_hint: boolean;
}

interface Config {
  name: string;
  description: string;
  input_schema: Param[];
  annotations: Annotations;
}

// Schema: https://modelcontextprotocol.io/specification/2025-11-25/schema#toolannotations
export default function McpToolEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const setAnnotation = (next: Annotations) => {
    const ro = next.read_only_hint;
    onChange({
      ...value,
      annotations: {
        ...next,
        destructive_hint: ro ? false : next.destructive_hint,
        idempotent_hint: ro ? false : next.idempotent_hint,
      },
    });
  };

  return (
    <div className="space-y-4">
      <TextField
        label="Tool Name"
        description="Unique identifier AI assistants use to call this tool."
        required
        value={value.name}
        onChange={(v) => onChange({ ...value, name: v })}
        error={errors?.name}
      />
      <TextAreaField
        label="Description"
        description="Helps AI assistants understand when and how to use this tool."
        required
        value={value.description}
        onChange={(v) => onChange({ ...value, description: v })}
        error={errors?.description}
      />
      <ParameterListField
        label="Input Parameters"
        value={value.input_schema}
        onChange={(v) => onChange({ ...value, input_schema: v })}
      />
      <section className="space-y-3 border-t pt-3">
        <div>
          <h4 className="text-sm font-medium">Annotations</h4>
          <p className="text-xs text-muted-foreground">
            Advisory behavioural hints. Per the MCP spec these are not a
            security boundary, but clients use them to decide whether to
            auto-approve calls.
          </p>
        </div>
        <TextField
          label="Display Title"
          size="sm"
          value={value.annotations.title ?? ""}
          onChange={(t) =>
            setAnnotation({ ...value.annotations, title: t || undefined })
          }
        />
        <CheckboxField
          label="Read-only"
          description="Tool does not modify its environment."
          checked={value.annotations.read_only_hint}
          onChange={(c) =>
            setAnnotation({ ...value.annotations, read_only_hint: c })
          }
        />
        <CheckboxField
          label="Destructive"
          description="May perform destructive updates (only meaningful when not read-only)."
          checked={value.annotations.destructive_hint}
          onChange={(c) =>
            setAnnotation({ ...value.annotations, destructive_hint: c })
          }
          disabled={value.annotations.read_only_hint}
        />
        <CheckboxField
          label="Idempotent"
          description="Repeated calls with the same arguments have no additional effect."
          checked={value.annotations.idempotent_hint}
          onChange={(c) =>
            setAnnotation({ ...value.annotations, idempotent_hint: c })
          }
          disabled={value.annotations.read_only_hint}
        />
        <CheckboxField
          label="Open world"
          description="Interacts with external entities (e.g. web search). Disable for closed-domain tools."
          checked={value.annotations.open_world_hint}
          onChange={(c) =>
            setAnnotation({ ...value.annotations, open_world_hint: c })
          }
        />
      </section>
    </div>
  );
}
