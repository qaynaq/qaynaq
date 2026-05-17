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
import type { ComponentCategory } from "../types";

interface Props {
  label: string;
  description?: string;
  category: ComponentCategory;
  value: ListItem[];
  onChange: (next: ListItem[]) => void;
  errors?: Record<string, string>;
  errorPathPrefix?: string;
}

export function ComponentListField({
  label,
  description,
  category,
  value,
  onChange,
  errors,
  errorPathPrefix,
}: Props) {
  const choices = listComponents(category);

  const updateItem = (i: number, next: ListItem) => {
    const arr = [...value];
    arr[i] = next;
    onChange(arr);
  };
  const removeItem = (i: number) =>
    onChange(value.filter((_, idx) => idx !== i));
  const addItem = () => {
    const first = choices[0];
    if (!first) return;
    onChange([...value, newListItem(category, first.id)]);
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
          onClick={addItem}
          disabled={choices.length === 0}
          className="h-7 text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add
        </Button>
      </div>
      {value.length === 0 && (
        <p className="text-sm text-muted-foreground py-3 text-center border border-dashed rounded-lg">
          None configured.
        </p>
      )}
      {value.map((item, i) => {
        const component = getComponent(category, item.componentId);
        const Editor = component?.Editor;
        const itemErrors: Record<string, string> | undefined = errors
          ? Object.fromEntries(
              Object.entries(errors)
                .filter(([k]) =>
                  k.startsWith(`${errorPathPrefix ?? ""}${i}.`),
                )
                .map(([k, v]) => [
                  k.slice(`${errorPathPrefix ?? ""}${i}.`.length),
                  v,
                ]),
            )
          : undefined;
        return (
          <div
            key={i}
            className="rounded-md border p-3 space-y-3"
          >
            <div className="flex items-center justify-between gap-2">
              <Select
                value={item.componentId}
                onValueChange={(id) =>
                  updateItem(i, newListItem(category, id))
                }
              >
                <SelectTrigger className="h-8 text-sm w-auto min-w-[160px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {choices.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => removeItem(i)}
                className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
            {Editor && (
              <div className="pl-3 border-l-2 border-border">
                <Suspense
                  fallback={
                    <p className="text-xs text-muted-foreground">Loading...</p>
                  }
                >
                  <Editor
                    value={item.config as never}
                    onChange={(next: unknown) =>
                      updateItem(i, { ...item, config: next })
                    }
                    errors={itemErrors}
                  />
                </Suspense>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
