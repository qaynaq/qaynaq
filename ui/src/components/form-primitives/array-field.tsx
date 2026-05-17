import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Plus, Trash2 } from "lucide-react";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

interface Props<T extends string | number> extends Omit<FieldWrapperProps, "children"> {
  value: T[];
  onChange: (next: T[]) => void;
  itemType?: "string" | "number";
  placeholder?: string;
  disabled?: boolean;
}

export function ArrayField<T extends string | number = string>({
  value,
  onChange,
  itemType = "string",
  placeholder,
  disabled,
  ...wrapper
}: Props<T>) {
  const updateItem = (i: number, raw: string) => {
    const next = [...value];
    next[i] = (itemType === "number" ? (Number(raw) as T) : (raw as T));
    onChange(next);
  };
  const removeItem = (i: number) => onChange(value.filter((_, idx) => idx !== i));
  const addItem = () =>
    onChange([...value, (itemType === "number" ? (0 as T) : ("" as T))]);

  return (
    <FieldWrapper {...wrapper}>
      <div className="space-y-2">
        {value.map((v, i) => (
          <div key={i} className="flex items-center gap-2">
            <Input
              type={itemType === "number" ? "number" : "text"}
              value={String(v ?? "")}
              onChange={(e) => updateItem(i, e.target.value)}
              placeholder={placeholder}
              disabled={disabled}
              className="h-8 text-sm"
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => removeItem(i)}
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
          onClick={addItem}
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
