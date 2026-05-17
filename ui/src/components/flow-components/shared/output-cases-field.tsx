import { Suspense } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Plus, Trash2 } from "lucide-react";
import { listComponents, getComponent } from "../registry";
import { newListItem, type ListItem } from "../utils/list-items";
import {
  TextField,
  CheckboxField,
} from "@/components/form-primitives";

export interface OutputCase {
  check: string;
  output: ListItem;
  continue?: boolean;
}

interface Props {
  label: string;
  description?: string;
  value: OutputCase[];
  onChange: (next: OutputCase[]) => void;
}

export function OutputCasesField({
  label,
  description,
  value,
  onChange,
}: Props) {
  const choices = listComponents("output");

  const update = (i: number, next: OutputCase) => {
    const arr = [...value];
    arr[i] = next;
    onChange(arr);
  };
  const remove = (i: number) =>
    onChange(value.filter((_, idx) => idx !== i));
  const add = () => {
    const first = choices[0];
    if (!first) return;
    onChange([
      ...value,
      { check: "", output: newListItem("output", first.id), continue: false },
    ]);
  };

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
          disabled={choices.length === 0}
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
      {value.map((c, i) => {
        const component = getComponent("output", c.output.componentId);
        const Editor = component?.Editor;
        return (
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
              description="A Bloblang query evaluated for each message. Empty matches everything."
              size="sm"
              value={c.check}
              onChange={(v) => update(i, { ...c, check: v })}
            />
            <CheckboxField
              label="Continue"
              description="Whether matching messages should also be tested against subsequent cases."
              checked={!!c.continue}
              onChange={(checked) => update(i, { ...c, continue: checked })}
            />
            <div className="space-y-2">
              <Label className="text-xs font-normal text-muted-foreground">Output</Label>
              <Select
                value={c.output.componentId}
                onValueChange={(id) =>
                  update(i, { ...c, output: newListItem("output", id) })
                }
              >
                <SelectTrigger className="h-8 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {choices.map((co) => (
                    <SelectItem key={co.id} value={co.id}>
                      {co.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {Editor && (
                <div className="pl-3 border-l-2 border-border">
                  <Suspense
                    fallback={
                      <p className="text-xs text-muted-foreground">Loading...</p>
                    }
                  >
                    <Editor
                      value={c.output.config as never}
                      onChange={(next: unknown) =>
                        update(i, {
                          ...c,
                          output: { ...c.output, config: next },
                        })
                      }
                    />
                  </Suspense>
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
