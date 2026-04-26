import React from "react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FieldSchema } from "../types";
import { TextInputField } from "./text-input-field";
import { LazyCodeEditorField, LazyArrayEditor } from "./lazy-components";
import { Suspense } from "react";
import { KeyValueEditor } from "./key-value-editor";

interface NestedPropertyInputProps {
  propKey: string;
  propSchema: FieldSchema;
  propValue: any;
  updateValue: (value: any) => void;
  previewMode?: boolean;
}

export function NestedPropertyInput({
  propKey,
  propSchema,
  propValue,
  updateValue,
  previewMode = false,
}: NestedPropertyInputProps) {
  switch (propSchema.type) {
    case "input":
      return (
        <TextInputField
          value={propValue || ""}
          onChange={updateValue}
          previewMode={previewMode}
          placeholder={propSchema.description}
        />
      );

    case "number":
      return (
        <Input
          type="number"
          value={propValue || 0}
          onChange={(e) => updateValue(Number(e.target.value))}
          className="h-8 text-sm"
          disabled={previewMode}
        />
      );

    case "bool":
      return (
        <div className="flex items-center gap-2">
          <Switch
            checked={propValue || false}
            onCheckedChange={updateValue}
            disabled={previewMode}
          />
          <span className="text-sm text-muted-foreground">
            {propValue ? "Enabled" : "Disabled"}
          </span>
        </div>
      );

    case "select":
      return (
        <Select
          value={propValue || ""}
          onValueChange={updateValue}
          disabled={previewMode}
        >
          <SelectTrigger className="h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {propSchema.options?.map((option) => (
              <SelectItem key={option} value={option}>
                {option}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );

    case "array":
      return (
        <Suspense
          fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
        >
          <LazyArrayEditor
            value={propValue || []}
            updateValue={updateValue}
            previewMode={previewMode}
          />
        </Suspense>
      );

    case "key_value":
      return (
        <KeyValueEditor
          value={propValue || {}}
          updateValue={updateValue}
          previewMode={previewMode}
        />
      );

    case "code":
      return (
        <Suspense
          fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
        >
          <LazyCodeEditorField
            value={propValue || ""}
            onChange={updateValue}
            previewMode={previewMode}
            language={propSchema.language}
          />
        </Suspense>
      );

    default:
      return (
        <TextInputField
          value={String(propValue || "")}
          onChange={updateValue}
          previewMode={previewMode}
          placeholder={propSchema.description}
        />
      );
  }
}
