import { useState, useEffect, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { InlineYamlEditor } from "@/components/inline-yaml-editor";
import { getComponentIcon } from "@/lib/component-catalog";

// Define FlowNodeData type locally since the file was deleted
export interface FlowNodeData {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
}

// Define a basic structure for ComponentSchema, assuming it will be passed from props
export interface ComponentSchema {
  id: string;
  name: string; 
  component: string; 
  type: "input" | "processor" | "output";
  schema?: any; // For YAML validation later
}

export interface AllComponentSchemas {
  input: ComponentSchema[];
  processor: ComponentSchema[];
  output: ComponentSchema[];
}

interface NodeConfigPanelProps {
  selectedNode: { id: string; data: FlowNodeData } | null;
  allComponentSchemas: AllComponentSchemas;
  onUpdateNode: (nodeId: string, data: FlowNodeData) => void;
  onDeleteNode: (nodeId: string) => void;
  lockedComponentId?: string;
}

export function NodeConfigPanel({
  selectedNode,
  allComponentSchemas,
  onUpdateNode,
  onDeleteNode,
  lockedComponentId,
}: NodeConfigPanelProps) {
  const [nodeData, setNodeData] = useState<FlowNodeData | null>(null);

  useEffect(() => {
    if (selectedNode) {
      const currentData = selectedNode.data as FlowNodeData;
      // Only update if the data has actually changed
      setNodeData(prev => {
        if (!prev || 
            prev.label !== currentData.label ||
            prev.componentId !== currentData.componentId ||
            prev.configYaml !== currentData.configYaml ||
            prev.component !== currentData.component) {
          return { ...currentData };
        }
        return prev;
      });
    } else {
      setNodeData(null);
    }
  }, [selectedNode?.id, selectedNode?.data.label, selectedNode?.data.componentId, selectedNode?.data.configYaml, selectedNode?.data.component]);

  const handleDebouncedUpdate = useCallback(
    (field: keyof FlowNodeData, value: any) => {
      if (selectedNode && nodeData) {
        const updatedData = {
            ...nodeData,
            [field]: value
        };
        setNodeData(updatedData);
        onUpdateNode(selectedNode.id, updatedData);
      }
    },
    [selectedNode, nodeData, onUpdateNode]
  );

  const getComponentSchema = (componentId: string, nodeType: "input" | "processor" | "output") => {
    const component = allComponentSchemas[nodeType]?.find(c => c.id === componentId);
    const schema = component?.schema || {};
    
    return schema;
  };

  if (!selectedNode || !nodeData) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Node Configuration</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">Select a node to configure it</p>
        </CardContent>
      </Card>
    );
  }

  if (lockedComponentId) {
    return (
      <Card className="w-full h-full flex flex-col">
        <CardHeader>
          <CardTitle>
            Configure{" "}
            {(nodeData.type.charAt(0).toUpperCase() + nodeData.type.slice(1))} Node
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col flex-1 p-4 min-h-0">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Label</Label>
              <Input value={nodeData.label} disabled />
            </div>
            <div className="space-y-2">
              <Label>Component</Label>
              <Input value={nodeData.component || nodeData.componentId || ""} disabled />
            </div>
            <p className="text-xs text-muted-foreground">
              Output is automatically configured for MCP Server input.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full h-full flex flex-col">
      <CardHeader>
        <CardTitle>
          Configure{" "}
          {(nodeData.type.charAt(0).toUpperCase() + nodeData.type.slice(1))} Node
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col flex-1 p-4 min-h-0">
        <div className="space-y-4 flex-shrink-0">
          <div className="space-y-2">
            <Label htmlFor="node-label">Label</Label>
            <Input
              id="node-label"
              value={nodeData.label}
              onChange={(e) => {
                const val = e.target.value;
                if (/^[a-z0-9_-]*$/.test(val)) {
                  if (selectedNode && nodeData) {
                      const updatedData = { ...nodeData, label: val };
                      setNodeData(updatedData);
                      onUpdateNode(selectedNode.id, updatedData);
                  }
                }
              }}
              placeholder="Node label (e.g., my_kafka_input)"
            />
          </div>

          <div className="space-y-2">
            <Label>Component</Label>
            {(() => {
              const Icon = getComponentIcon(nodeData.type, nodeData.componentId);
              return (
                <div className="flex items-center gap-2 rounded-md border border-border bg-muted/40 px-3 py-2">
                  <Icon className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm">
                    {nodeData.component || nodeData.componentId || "-"}
                  </span>
                </div>
              );
            })()}
            <p className="text-[10px] text-muted-foreground">
              Component is locked. Delete the node to pick a different one.
            </p>
          </div>
        </div>

        {nodeData.componentId && Object.keys(getComponentSchema(nodeData.componentId, nodeData.type).properties || {}).length > 0 && (
          <div className="flex flex-col flex-1 min-h-0 mt-4">
            <Label htmlFor="yaml-config" className="mb-2">Component Configuration</Label>
            <div className="flex-1 min-h-0 overflow-y-auto">
              <InlineYamlEditor
                schema={getComponentSchema(nodeData.componentId, nodeData.type)}
                value={nodeData.configYaml || ""}
                onChange={(yamlValue: string) => handleDebouncedUpdate("configYaml", yamlValue)}
                availableProcessors={allComponentSchemas.processor.map(p => ({
                  id: p.id,
                  name: p.name,
                  component: p.component,
                  type: p.type,
                  schema: p.schema
                }))}
                availableInputs={allComponentSchemas.input.map(i => ({
                  id: i.id,
                  name: i.name,
                  component: i.component,
                  type: i.type,
                  schema: i.schema
                }))}
                availableOutputs={allComponentSchemas.output.map(o => ({
                  id: o.id,
                  name: o.name,
                  component: o.component,
                  type: o.type,
                  schema: o.schema
                }))}
              />
            </div>
          </div>
        )}

        <Button
          variant="destructive"
          className="mt-4 flex-shrink-0"
          onClick={() => onDeleteNode(selectedNode.id)}
        >
          Delete Node
        </Button>
      </CardContent>
    </Card>
  );
}
