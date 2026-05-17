import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { Label } from "@/components/ui/label";

interface Param {
  name: string;
  type: string;
  required: boolean;
  description: string;
}

const TYPES = ["string", "number", "boolean", "array", "object"];

interface Props {
  label: string;
  value: Param[];
  onChange: (next: Param[]) => void;
}

export function ParameterListField({ label, value, onChange }: Props) {
  const update = (i: number, key: keyof Param, v: string | boolean) => {
    const next = [...value];
    next[i] = { ...next[i], [key]: v };
    onChange(next);
  };
  const add = () =>
    onChange([
      ...value,
      { name: "", type: "string", required: false, description: "" },
    ]);
  const remove = (i: number) => onChange(value.filter((_, idx) => idx !== i));

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <Label>{label}</Label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={add}
          className="h-7 text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add Parameter
        </Button>
      </div>
      {value.length === 0 && (
        <p className="text-sm text-muted-foreground py-4 text-center border border-dashed rounded-lg">
          No parameters defined.
        </p>
      )}
      {value.map((p, i) => (
        <div
          key={i}
          className="flex items-start gap-3 p-3 rounded-lg border bg-muted/30"
        >
          <div className="flex-1 space-y-2">
            <div className="grid grid-cols-[1fr_120px] gap-2">
              <Input
                value={p.name}
                onChange={(e) => update(i, "name", e.target.value)}
                placeholder="Parameter name"
                className="h-8 text-sm"
              />
              <Select
                value={p.type}
                onValueChange={(v) => update(i, "type", v)}
              >
                <SelectTrigger className="h-8 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TYPES.map((t) => (
                    <SelectItem key={t} value={t}>
                      {t}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Input
              value={p.description}
              onChange={(e) => update(i, "description", e.target.value)}
              placeholder="Description"
              className="h-8 text-sm"
            />
            <div className="flex items-center gap-2">
              <Checkbox
                checked={p.required}
                onCheckedChange={(c) => update(i, "required", !!c)}
                className="h-3.5 w-3.5"
              />
              <span className="text-xs text-muted-foreground">Required</span>
            </div>
          </div>
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
      ))}
    </div>
  );
}
