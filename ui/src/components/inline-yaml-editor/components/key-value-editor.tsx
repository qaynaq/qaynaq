import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Trash2 } from "lucide-react";
import { EditorProps } from "../types";

interface KeyValueEditorProps extends EditorProps {
  value: Record<string, any>;
  updateValue: (value: Record<string, any>) => void;
}

export function KeyValueEditor({
  value,
  updateValue,
  previewMode = false,
}: KeyValueEditorProps) {
  const pairs = Object.entries(value);

  const addPair = () => {
    updateValue({ ...value, "": "" });
  };

  const updatePairKey = (oldKey: string, newKey: string) => {
    const newObj = { ...value };
    if (oldKey !== newKey) {
      const val = newObj[oldKey] || "";
      delete newObj[oldKey];
      if (newKey.trim()) {
        newObj[newKey] = val;
      }
    }
    updateValue(newObj);
  };

  const updatePairValue = (key: string, newValue: string) => {
    const newObj = { ...value };
    newObj[key] = newValue;
    updateValue(newObj);
  };

  const removePair = (key: string) => {
    const newObj = { ...value };
    delete newObj[key];
    updateValue(newObj);
  };

  if (pairs.length === 0) {
    return (
      <div>
        {!previewMode ? (
          <Button
            variant="outline"
            size="sm"
            onClick={addPair}
            className="h-7 text-xs"
          >
            <Plus className="h-3 w-3 mr-1" />
            Add entry
          </Button>
        ) : (
          <span className="text-sm text-muted-foreground">No entries</span>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {pairs.map(([key, val], index) => (
        <div key={index} className="flex items-center gap-2">
          <Input
            value={key}
            onChange={(e) => updatePairKey(key, e.target.value)}
            placeholder="Key"
            className="h-8 text-sm flex-1"
            disabled={previewMode}
          />
          <Input
            value={val}
            onChange={(e) => updatePairValue(key, e.target.value)}
            placeholder="Value"
            className="h-8 text-sm flex-1"
            disabled={previewMode}
          />
          {!previewMode && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => removePair(key)}
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
          onClick={addPair}
          className="h-7 text-xs"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add entry
        </Button>
      )}
    </div>
  );
}
