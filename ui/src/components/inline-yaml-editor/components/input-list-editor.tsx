import React from "react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { EditorProps, ProcessorComponentSchema } from "../types";
import { InlineYamlEditor } from "../inline-yaml-editor";

interface InputListEditorProps extends EditorProps {
  value: any[];
  updateValue: (value: any[]) => void;
  availableInputs: ProcessorComponentSchema[];
}

export function InputListEditor({ 
  value, 
  updateValue, 
  availableInputs,
  availableProcessors = [],
  availableOutputs = [],
  previewMode = false 
}: InputListEditorProps) {
  const addInput = () => {
    const newInput = { componentId: "", component: "", configYaml: "" };
    updateValue([...value, newInput]);
  };

  const updateInput = (index: number, field: string, newValue: string) => {
    const updatedInputs = [...value];
    if (field === "componentId") {
      const input = availableInputs.find(i => i.id === newValue);
      if (input) {
        updatedInputs[index] = {
          ...updatedInputs[index],
          componentId: newValue,
          component: input.component,
          configYaml: ""
        };
      }
    } else {
      updatedInputs[index] = { ...updatedInputs[index], [field]: newValue };
    }
    updateValue(updatedInputs);
  };

  const removeInput = (index: number) => {
    updateValue(value.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-2">
      {value.map((input: any, index: number) => (
        <div key={index} className="border border-gray-600 rounded p-2 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-gray-300">Input {index + 1}</span>
            {!previewMode && (
              <Button
                onClick={() => removeInput(index)}
                size="sm"
                variant="ghost"
                className="h-6 w-6 p-0 text-red-400 hover:text-red-300"
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            )}
          </div>
          
          <div className="space-y-1">
            <label className="text-xs text-gray-400">Component:</label>
            <Select
              value={input.componentId || ""}
              onValueChange={(value) => updateInput(index, "componentId", value)}
              disabled={previewMode}
            >
              <SelectTrigger className="h-6 text-xs">
                <SelectValue placeholder="Select input..." />
              </SelectTrigger>
              <SelectContent>
                {availableInputs.map((availableInput) => (
                  <SelectItem key={availableInput.id} value={availableInput.id}>
                    {availableInput.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {input.componentId && (
            <div className="space-y-1">
              <label className="text-xs text-gray-400">Configuration:</label>
              <InlineYamlEditor
                schema={availableInputs.find(i => i.id === input.componentId)?.schema || {}}
                value={input.configYaml || ""}
                onChange={(yamlValue: string) => updateInput(index, "configYaml", yamlValue)}
                availableProcessors={availableProcessors}
                availableInputs={availableInputs}
                availableOutputs={availableOutputs}
                previewMode={previewMode}
              />
            </div>
          )}
        </div>
      ))}
      {!previewMode && (
        <Button
          onClick={addInput}
          size="sm"
          variant="ghost"
          className="h-6 text-sm bg-gray-700 hover:bg-gray-600"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add Input
        </Button>
      )}
    </div>
  );
} 