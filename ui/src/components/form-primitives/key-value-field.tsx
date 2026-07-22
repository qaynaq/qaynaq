import { useEffect, useRef, useState } from "react";
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

interface Row {
  id: number;
  key: string;
  value: string;
}

const toRecord = (rows: Row[]) =>
  Object.fromEntries(rows.map((r) => [r.key, r.value]));

const sameRecord = (a: Record<string, string>, b: Record<string, string>) =>
  Object.keys(a).length === Object.keys(b).length &&
  Object.keys(a).every((k) => b[k] === a[k]);

export function KeyValueField({
  value,
  onChange,
  disabled,
  keyPlaceholder = "Key",
  valuePlaceholder = "Value",
  ...wrapper
}: Props) {
  const idRef = useRef(0);
  const toRows = (record: Record<string, string>): Row[] =>
    Object.entries(record).map(([k, v]) => ({ id: ++idRef.current, key: k, value: v }));

  const [rows, setRows] = useState(() => toRows(value));

  // Parents round-trip changes through YAML, which drops ""-valued entries;
  // rebuild rows only on genuine external changes, not echoes of our onChange.
  useEffect(() => {
    setRows((current) => {
      const full = toRecord(current);
      const committed = Object.fromEntries(
        Object.entries(full).filter(([, v]) => v !== ""),
      );
      return sameRecord(value, full) || sameRecord(value, committed)
        ? current
        : toRows(value);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  const commit = (next: Row[]) => {
    setRows(next);
    onChange(toRecord(next));
  };

  const update = (id: number, patch: Partial<Row>) =>
    commit(rows.map((r) => (r.id === id ? { ...r, ...patch } : r)));
  const remove = (id: number) => commit(rows.filter((r) => r.id !== id));
  const add = () => commit([...rows, { id: ++idRef.current, key: "", value: "" }]);

  return (
    <FieldWrapper {...wrapper}>
      <div className="space-y-2">
        {rows.map((row) => (
          <div key={row.id} className="flex items-center gap-2">
            <Input
              value={row.key}
              onChange={(e) => update(row.id, { key: e.target.value })}
              placeholder={keyPlaceholder}
              className="h-8 text-sm flex-1"
              disabled={disabled}
            />
            <Input
              value={row.value}
              onChange={(e) => update(row.id, { value: e.target.value })}
              placeholder={valuePlaceholder}
              className="h-8 text-sm flex-1"
              disabled={disabled}
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => remove(row.id)}
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
