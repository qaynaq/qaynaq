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

interface OutputListEditorProps extends EditorProps {
  value: any[];
  updateValue: (value: any[]) => void;
  availableOutputs: ProcessorComponentSchema[];
}

export function OutputListEditor({ 
  value, 
  updateValue, 
  availableOutputs,
  availableProcessors = [],
  availableInputs = [],
  previewMode = false 
}: OutputListEditorProps) {
  const addOutput = () => {
    const newOutput = { componentId: "", component: "", configYaml: "" };
    updateValue([...value, newOutput]);
  };

  const updateOutput = (index: number, field: string, newValue: string) => {
    const updatedOutputs = [...value];
    if (field === "componentId") {
      const output = availableOutputs.find(o => o.id === newValue);
      if (output) {
        updatedOutputs[index] = {
          ...updatedOutputs[index],
          componentId: newValue,
          component: output.component,
          configYaml: ""
        };
      }
    } else {
      updatedOutputs[index] = { ...updatedOutputs[index], [field]: newValue };
    }
    updateValue(updatedOutputs);
  };

  const removeOutput = (index: number) => {
    updateValue(value.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-2">
      {value.map((output: any, index: number) => (
        <div key={index} className="border border-gray-600 rounded p-2 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-gray-300">Output {index + 1}</span>
            {!previewMode && (
              <Button
                onClick={() => removeOutput(index)}
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
              value={output.componentId || ""}
              onValueChange={(value) => updateOutput(index, "componentId", value)}
              disabled={previewMode}
            >
              <SelectTrigger className="h-6 text-xs">
                <SelectValue placeholder="Select output..." />
              </SelectTrigger>
              <SelectContent>
                {availableOutputs.map((availableOutput) => (
                  <SelectItem key={availableOutput.id} value={availableOutput.id}>
                    {availableOutput.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {output.componentId && (
            <div className="space-y-1">
              <label className="text-xs text-gray-400">Configuration:</label>
              <InlineYamlEditor
                schema={availableOutputs.find(o => o.id === output.componentId)?.schema || {}}
                value={output.configYaml || ""}
                onChange={(yamlValue: string) => updateOutput(index, "configYaml", yamlValue)}
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
          onClick={addOutput}
          size="sm"
          variant="ghost"
          className="h-6 text-sm bg-gray-700 hover:bg-gray-600"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add Output
        </Button>
      )}
    </div>
  );
} 