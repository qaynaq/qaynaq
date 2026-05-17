import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Plus, Trash2 } from "lucide-react";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: Record<string, string>;
  onChange: (next: Record<string, string>) => void;
  disabled?: boolean;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
}

export function KeyValueField({
  value,
  onChange,
  disabled,
  keyPlaceholder = "Key",
  valuePlaceholder = "Value",
  ...wrapper
}: Props) {
  const entries = Object.entries(value);

  const updateKey = (oldKey: string, newKey: string) => {
    if (newKey === oldKey) return;
    const next: Record<string, string> = {};
    for (const [k, v] of entries) next[k === oldKey ? newKey : k] = v;
    onChange(next);
  };
  const updateValue = (k: string, v: string) => {
    onChange({ ...value, [k]: v });
  };
  const remove = (k: string) => {
    const next = { ...value };
    delete next[k];
    onChange(next);
  };
  const add = () => {
    let i = 1;
    let key = "key";
    while (key in value) {
      key = `key_${i++}`;
    }
    onChange({ ...value, [key]: "" });
  };

  return (
    <FieldWrapper {...wrapper}>
      <div className="space-y-2">
        {entries.map(([k, v]) => (
          <div key={k} className="flex items-center gap-2">
            <Input
              value={k}
              onChange={(e) => updateKey(k, e.target.value)}
              placeholder={keyPlaceholder}
              className="h-8 text-sm flex-1"
              disabled={disabled}
            />
            <Input
              value={v}
              onChange={(e) => updateValue(k, e.target.value)}
              placeholder={valuePlaceholder}
              className="h-8 text-sm flex-1"
              disabled={disabled}
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => remove(k)}
              disabled={disabled}
              className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        ))}
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={add}
          disabled={disabled}
          className="h-7 text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add
        </Button>
      </div>
    </FieldWrapper>
  );
}
