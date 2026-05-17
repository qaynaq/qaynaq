import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Plus, Trash2 } from "lucide-react";
import {
  TextField,
  CheckboxField,
} from "@/components/form-primitives";
import { ComponentListField } from "./component-list-field";
import type { ListItem } from "../utils/list-items";

export interface ProcessorCase {
  check: string;
  processors: ListItem[];
  fallthrough?: boolean;
}

interface Props {
  label: string;
  description?: string;
  value: ProcessorCase[];
  onChange: (next: ProcessorCase[]) => void;
}

export function ProcessorCasesField({
  label,
  description,
  value,
  onChange,
}: Props) {
  const update = (i: number, next: ProcessorCase) => {
    const arr = [...value];
    arr[i] = next;
    onChange(arr);
  };
  const remove = (i: number) =>
    onChange(value.filter((_, idx) => idx !== i));
  const add = () =>
    onChange([
      ...value,
      { check: "", processors: [], fallthrough: false },
    ]);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div>
          <Label>{label}</Label>
          {description && (
            <p className="text-xs text-muted-foreground">{description}</p>
          )}
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={add}
          className="h-7 text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add case
        </Button>
      </div>
      {value.length === 0 && (
        <p className="text-sm text-muted-foreground py-3 text-center border border-dashed rounded-lg">
          No cases defined.
        </p>
      )}
      {value.map((c, i) => (
        <div key={i} className="rounded-md border p-3 space-y-3">
          <div className="flex items-center justify-end">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => remove(i)}
              className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
          <TextField
            label="Check"
            description="A Bloblang query that gates this case. Empty matches everything."
            size="sm"
            value={c.check}
            onChange={(v) => update(i, { ...c, check: v })}
          />
          <CheckboxField
            label="Fallthrough"
            description="Whether matching messages should continue through subsequent cases."
            checked={!!c.fallthrough}
            onChange={(checked) => update(i, { ...c, fallthrough: checked })}
          />
          <ComponentListField
            label="Processors"
            category="processor"
            value={c.processors}
            onChange={(next) => update(i, { ...c, processors: next })}
          />
        </div>
      ))}
    </div>
  );
}
