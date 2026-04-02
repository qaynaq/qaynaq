import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { EditorProps } from "../types";

export interface PropertyItem {
  name: string;
  type: string;
  required?: boolean;
  description?: string;
}

const PROPERTY_TYPES = ["string", "number", "boolean", "array", "object"];

interface PropertyListEditorProps extends EditorProps {
  value: PropertyItem[];
  updateValue: (value: PropertyItem[]) => void;
}

export function PropertyListEditor({
  value,
  updateValue,
  previewMode = false,
}: PropertyListEditorProps) {
  const items: PropertyItem[] = Array.isArray(value) ? value : [];

  const addItem = () => {
    updateValue([...items, { name: "", type: "string", required: false, description: "" }]);
  };

  const updateItem = (index: number, field: keyof PropertyItem, newValue: any) => {
    const updated = [...items];
    updated[index] = { ...updated[index], [field]: newValue };
    updateValue(updated);
  };

  const removeItem = (index: number) => {
    updateValue(items.filter((_, i) => i !== index));
  };

  if (items.length === 0) {
    return (
      <div>
        {!previewMode ? (
          <Button
            variant="outline"
            size="sm"
            onClick={addItem}
            className="h-7 text-xs"
          >
            <Plus className="h-3 w-3 mr-1" />
            Add parameter
          </Button>
        ) : (
          <span className="text-sm text-muted-foreground">No parameters</span>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {items.map((item, index) => (
        <div
          key={index}
          className="flex items-start gap-3 p-3 rounded-lg border bg-muted/30"
        >
          <div className="flex-1 space-y-2">
            <div className="grid grid-cols-[1fr_120px] gap-2">
              <Input
                value={item.name}
                onChange={(e) => updateItem(index, "name", e.target.value)}
                placeholder="Parameter name"
                className="h-8 text-sm"
                disabled={previewMode}
              />
              <Select
                value={item.type || "string"}
                onValueChange={(val) => updateItem(index, "type", val)}
                disabled={previewMode}
              >
                <SelectTrigger className="h-8 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PROPERTY_TYPES.map((t) => (
                    <SelectItem key={t} value={t}>
                      {t}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Input
              value={item.description || ""}
              onChange={(e) => updateItem(index, "description", e.target.value)}
              placeholder="Description"
              className="h-8 text-sm"
              disabled={previewMode}
            />
            <div className="flex items-center gap-2">
              <Checkbox
                checked={item.required || false}
                onCheckedChange={(checked) => updateItem(index, "required", !!checked)}
                className="h-4 w-4"
                disabled={previewMode}
              />
              <span className="text-xs text-muted-foreground">Required</span>
            </div>
          </div>
          {!previewMode && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => removeItem(index)}
              className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          )}
        </div>
      ))}
      {!previewMode && (
        <Button
          variant="outline"
          size="sm"
          onClick={addItem}
          className="h-7 text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add parameter
        </Button>
      )}
    </div>
  );
}
