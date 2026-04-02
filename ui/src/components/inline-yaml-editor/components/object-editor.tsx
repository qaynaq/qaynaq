import React from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { EditorProps, FieldSchema } from "../types";
import { getDefaultValue } from "../utils/defaults";
import { KeyValueEditor } from "./key-value-editor";
import { NestedPropertyInput } from "./nested-property-input";

interface ObjectEditorProps extends EditorProps {
  value: Record<string, any>;
  updateValue: (value: Record<string, any>) => void;
  fieldSchema: FieldSchema;
}

export function ObjectEditor({
  value,
  updateValue,
  fieldSchema,
  previewMode = false
}: ObjectEditorProps) {
  if (fieldSchema.properties) {
    const updateNestedValue = (propKey: string, propValue: any) => {
      const newValue = { ...value, [propKey]: propValue };
      updateValue(newValue);
    };

    return (
      <div className="space-y-3 pl-4 border-l-2 border-border">
        {Object.entries(fieldSchema.properties).map(([propKey, propSchema]) => {
          const propValue = value[propKey] ?? getDefaultValue(propSchema);
          const isEnabled = value.hasOwnProperty(propKey);

          if (previewMode && !isEnabled) return null;

          return (
            <div key={propKey} className="space-y-1.5">
              <div className="flex items-center gap-2">
                {!previewMode && (
                  <Checkbox
                    checked={isEnabled}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        updateNestedValue(propKey, getDefaultValue(propSchema));
                      } else {
                        const newValue = { ...value };
                        delete newValue[propKey];
                        updateValue(newValue);
                      }
                    }}
                    className="h-4 w-4"
                  />
                )}
                <Label className={`text-sm ${isEnabled ? "text-foreground" : "text-muted-foreground"}`}>
                  {propSchema.title || propKey}
                </Label>
              </div>
              {isEnabled && (
                <div className={!previewMode ? "pl-6" : ""}>
                  <NestedPropertyInput
                    propKey={propKey}
                    propSchema={propSchema}
                    propValue={propValue}
                    updateValue={(val: any) => updateNestedValue(propKey, val)}
                    previewMode={previewMode}
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>
    );
  }

  return <KeyValueEditor value={value} updateValue={updateValue} previewMode={previewMode} />;
}
