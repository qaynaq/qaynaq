import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
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

interface OutputCasesEditorProps extends EditorProps {
  value: any[];
  updateValue: (value: any[]) => void;
  availableOutputs: ProcessorComponentSchema[];
  internalValue?: any[];
  setInternalValue?: (value: any[]) => void;
  isInternalUpdateRef?: React.MutableRefObject<boolean>;
}

export function OutputCasesEditor({ 
  value, 
  updateValue, 
  availableOutputs,
  availableProcessors = [],
  availableInputs = [],
  internalValue,
  setInternalValue,
  isInternalUpdateRef,
  previewMode = false 
}: OutputCasesEditorProps) {
  const displayValue = internalValue && internalValue.length > 0 ? internalValue : value;

  const addCase = () => {
    const newCase = { check: "", output: { componentId: "", component: "", configYaml: "" }, continue: false };
    const updatedCases = [...displayValue, newCase];
    
    if (setInternalValue) {
      setInternalValue(updatedCases);
    }
    
    if (isInternalUpdateRef) {
      isInternalUpdateRef.current = true;
    }
    updateValue(updatedCases);
  };

  const updateCase = (index: number, field: string, newValue: any) => {
    const updatedCases = [...displayValue];
    if (field === "check" || field === "continue") {
      updatedCases[index] = { ...updatedCases[index], [field]: newValue };
    } else if (field.startsWith("output.")) {
      const outputField = field.replace("output.", "");
      updatedCases[index] = {
        ...updatedCases[index],
        output: { ...updatedCases[index].output, [outputField]: newValue }
      };
    }
    
    if (setInternalValue) {
      setInternalValue(updatedCases);
    }
    
    if (isInternalUpdateRef) {
      isInternalUpdateRef.current = true;
    }
    updateValue(updatedCases);
  };

  const removeCase = (index: number) => {
    const updatedCases = displayValue.filter((_, i) => i !== index);
    
    if (setInternalValue) {
      setInternalValue(updatedCases);
    }
    
    if (isInternalUpdateRef) {
      isInternalUpdateRef.current = true;
    }
    updateValue(updatedCases);
  };

      return (
      <div className="space-y-2">
        {displayValue.map((caseItem: any, index: number) => (
        <div key={index} className="border border-gray-600 rounded p-2 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-gray-300">Case {index + 1}</span>
            {!previewMode && (
              <Button
                onClick={() => removeCase(index)}
                size="sm"
                variant="ghost"
                className="h-6 w-6 p-0 text-red-400 hover:text-red-300"
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            )}
          </div>
          
          <div className="space-y-1">
            <label className="text-xs text-gray-400">Check (Bloblang query):</label>
            <Input
              value={caseItem.check || ""}
              onChange={(e) => updateCase(index, "check", e.target.value)}
              placeholder="e.g., this.type == 'foo'"
              className="h-6 text-xs"
              disabled={previewMode}
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs text-gray-400">Output:</label>
            <Select
              value={caseItem.output?.componentId || ""}
                             onValueChange={(selectedValue) => {
                 const output = availableOutputs.find(o => o.id === selectedValue);
                 if (output) {
                   const updatedCases = [...displayValue];
                   updatedCases[index] = {
                     ...updatedCases[index],
                     output: {
                       componentId: selectedValue,
                       component: output.component,
                       configYaml: ""
                     }
                   };
                   
                   if (setInternalValue) {
                     setInternalValue(updatedCases);
                   }
                   
                   if (isInternalUpdateRef) {
                     isInternalUpdateRef.current = true;
                   }
                   updateValue(updatedCases);
                 }
               }}
              disabled={previewMode}
            >
              <SelectTrigger className="h-6 text-xs">
                <SelectValue placeholder="Select output..." />
              </SelectTrigger>
              <SelectContent>
                {availableOutputs.map((output) => (
                  <SelectItem key={output.id} value={output.id}>
                    {output.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {caseItem.output?.componentId && (
            <div className="space-y-1">
              <label className="text-xs text-gray-400">Output Configuration:</label>
              <InlineYamlEditor
                schema={availableOutputs.find(o => o.id === caseItem.output?.componentId)?.schema || {}}
                value={caseItem.output?.configYaml || ""}
                onChange={(yamlValue: string) => updateCase(index, "output.configYaml", yamlValue)}
                availableProcessors={availableProcessors}
                availableInputs={availableInputs}
                availableOutputs={availableOutputs}
                previewMode={previewMode}
              />
            </div>
          )}

          <div className="flex items-center space-x-2">
            <Switch
              checked={caseItem.continue || false}
              onCheckedChange={(checked) => updateCase(index, "continue", checked)}
              className="scale-75"
              disabled={previewMode}
            />
            <label className="text-xs text-gray-400">Continue to next case</label>
          </div>
        </div>
      ))}
      {!previewMode && (
        <Button
          onClick={addCase}
          size="sm"
          variant="ghost"
          className="h-6 text-sm bg-gray-700 hover:bg-gray-600"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add Case
        </Button>
      )}
    </div>
  );
} 