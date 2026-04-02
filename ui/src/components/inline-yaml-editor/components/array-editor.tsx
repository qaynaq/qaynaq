import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Trash2 } from "lucide-react";
import { EditorProps } from "../types";

interface ArrayEditorProps extends EditorProps {
  value: any[];
  updateValue: (value: any[]) => void;
}

export function ArrayEditor({
  value,
  updateValue,
  previewMode = false,
}: ArrayEditorProps) {
  const addItem = () => {
    updateValue([...value, ""]);
  };

  const updateItem = (index: number, newValue: string) => {
    const newArray = [...value];
    newArray[index] = newValue;
    updateValue(newArray);
  };

  const removeItem = (index: number) => {
    const newArray = value.filter((_, i) => i !== index);
    updateValue(newArray);
  };

  if (value.length === 0) {
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
            Add item
          </Button>
        ) : (
          <span className="text-sm text-muted-foreground">No items</span>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {value.map((item, index) => (
        <div key={index} className="flex items-center gap-2">
          <Input
            value={item}
            onChange={(e) => updateItem(index, e.target.value)}
            placeholder={`Item ${index + 1}`}
            className="h-8 text-sm flex-1"
            disabled={previewMode}
          />
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
          Add item
        </Button>
      )}
    </div>
  );
}
