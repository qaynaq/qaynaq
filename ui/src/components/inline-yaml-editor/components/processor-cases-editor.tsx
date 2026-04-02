import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Plus, Trash2 } from "lucide-react";
import { EditorProps, ProcessorComponentSchema } from "../types";
import { LazyProcessorListEditor } from "./lazy-components";
import { Suspense } from "react";

interface ProcessorCasesEditorProps extends EditorProps {
  value: any[];
  updateValue: (value: any[]) => void;
  availableProcessors: ProcessorComponentSchema[];
}

export function ProcessorCasesEditor({ 
  value, 
  updateValue, 
  availableProcessors,
  availableInputs = [],
  availableOutputs = [],
  previewMode = false 
}: ProcessorCasesEditorProps) {
  const addCase = () => {
    const newCase = { check: "", processors: [], fallthrough: false };
    updateValue([...value, newCase]);
  };

  const updateCase = (index: number, field: string, newValue: any) => {
    const updatedCases = [...value];
    updatedCases[index] = { ...updatedCases[index], [field]: newValue };
    updateValue(updatedCases);
  };

  const removeCase = (index: number) => {
    updateValue(value.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-2">
      {value.map((caseItem: any, index: number) => (
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
            <label className="text-xs text-gray-400">Check (Bloblang query - leave empty for default case):</label>
            <Input
              value={caseItem.check || ""}
              onChange={(e) => updateCase(index, "check", e.target.value)}
              placeholder="e.g., this.type == 'foo' (empty = default case)"
              className="h-6 text-xs"
              disabled={previewMode}
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs text-gray-400">Processors:</label>
            <Suspense fallback={<div className="text-xs text-gray-400">Loading...</div>}>
              <LazyProcessorListEditor 
                value={caseItem.processors || []} 
                updateValue={(processors: any[]) => updateCase(index, "processors", processors)} 
                availableProcessors={availableProcessors}
                availableInputs={availableInputs}
                availableOutputs={availableOutputs}
                previewMode={previewMode}
              />
            </Suspense>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={caseItem.fallthrough || false}
              onCheckedChange={(checked) => updateCase(index, "fallthrough", checked)}
              className="scale-75"
              disabled={previewMode}
            />
            <label className="text-xs text-gray-400">Fallthrough (continue to next case)</label>
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