import React from "react";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { EditorProps, ProcessorComponentSchema } from "../types";
import { NestedProcessor } from "./nested-processor";

interface ProcessorListEditorProps extends EditorProps {
  value: any[];
  updateValue: (value: any[]) => void;
  availableProcessors: ProcessorComponentSchema[];
}

export function ProcessorListEditor({ 
  value, 
  updateValue, 
  availableProcessors,
  availableInputs = [],
  availableOutputs = [],
  previewMode = false 
}: ProcessorListEditorProps) {
  const addProcessor = () => {
    updateValue([...value, { componentId: "", component: "", configYaml: "" }]);
  };

  const updateProcessor = (index: number, field: string, newValue: string) => {
    const newProcessors = [...value];
    const processor = { ...newProcessors[index] };
    
    if (field === 'componentId') {
      const selectedProc = availableProcessors.find(p => p.id === newValue);
      const oldComponentId = processor.componentId;
      
      processor.componentId = newValue;
      processor.component = selectedProc ? selectedProc.component : "";
      
      // Reset config when component changes (including switching between different components)
      if (oldComponentId !== newValue) {
        processor.configYaml = "";
      }
    } else {
      processor[field] = newValue;
    }
    
    newProcessors[index] = processor;
    updateValue(newProcessors);
  };

  const removeProcessor = (index: number) => {
    updateValue(value.filter((_, i) => i !== index));
  };

  if (value.length === 0) {
    return (
      <div className="flex items-center space-x-2">
        <span style={{ 
          fontFamily: 'monospace',
          fontSize: '13px',
          color: '#6b7280'
        }}>
          []
        </span>
        {!previewMode && (
          <Button
            onClick={addProcessor}
            size="sm"
            variant="ghost"
            className="h-6 text-sm bg-gray-700 hover:bg-gray-600"
          >
            <Plus className="h-3 w-3 mr-1" />
            Add Processor
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {value.map((processor, index) => (
        <NestedProcessor
          key={index}
          processor={processor}
          index={index}
          onUpdate={updateProcessor}
          onRemove={removeProcessor}
          availableProcessors={availableProcessors}
          availableInputs={availableInputs}
          availableOutputs={availableOutputs}
          previewMode={previewMode}
        />
      ))}
      {!previewMode && (
        <Button
          onClick={addProcessor}
          size="sm"
          variant="ghost"
          className="h-6 text-sm bg-gray-700 hover:bg-gray-600"
        >
          <Plus className="h-3 w-3 mr-1" />
          Add Processor
        </Button>
      )}
    </div>
  );
} 