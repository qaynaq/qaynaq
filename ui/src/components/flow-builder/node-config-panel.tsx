import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { getComponentIcon } from "@/lib/component-catalog";
import { getComponent } from "@/components/flow-components/registry";
import { flattenZodErrors } from "@/components/flow-components/utils/errors";
import type { ComponentCategory } from "@/components/flow-components/types";

export interface FlowNodeData {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
}

interface NodeConfigPanelProps {
  selectedNode: { id: string; data: FlowNodeData } | null;
  onUpdateNode: (nodeId: string, data: FlowNodeData) => void;
  onDeleteNode: (nodeId: string) => void;
  lockedComponentId?: string;
}

export function NodeConfigPanel({
  selectedNode,
  onUpdateNode,
  onDeleteNode,
  lockedComponentId,
}: NodeConfigPanelProps) {
  const [nodeData, setNodeData] = useState<FlowNodeData | null>(null);

  useEffect(() => {
    if (selectedNode) {
      setNodeData(selectedNode.data);
    } else {
      setNodeData(null);
    }
  }, [selectedNode?.id, selectedNode?.data]);

  const update = useCallback(
    (next: FlowNodeData) => {
      if (!selectedNode) return;
      setNodeData(next);
      onUpdateNode(selectedNode.id, next);
    },
    [selectedNode, onUpdateNode],
  );

  const handleConfigChange = useCallback(
    (yamlValue: string) => {
      if (!nodeData) return;
      update({ ...nodeData, configYaml: yamlValue });
    },
    [nodeData, update],
  );

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
            {nodeData.type.charAt(0).toUpperCase() + nodeData.type.slice(1)}{" "}
            Node
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
              <Input
                value={nodeData.component || nodeData.componentId || ""}
                disabled
              />
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
          {nodeData.type.charAt(0).toUpperCase() + nodeData.type.slice(1)} Node
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
                  update({ ...nodeData, label: val });
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

        {nodeData.componentId && (
          <div className="flex flex-col flex-1 min-h-0 mt-4">
            <Label className="mb-2">Component Configuration</Label>
            <div className="flex-1 min-h-0 overflow-y-auto">
              <ConfigEditor
                category={nodeData.type as ComponentCategory}
                componentId={nodeData.componentId}
                configYaml={nodeData.configYaml ?? ""}
                onChange={handleConfigChange}
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

interface ConfigEditorProps {
  category: ComponentCategory;
  componentId: string;
  configYaml: string;
  onChange: (yaml: string) => void;
}

function ConfigEditor({
  category,
  componentId,
  configYaml,
  onChange,
}: ConfigEditorProps) {
  const component = getComponent(category, componentId);

  const parsed = useMemo(() => {
    if (!component) return null;
    try {
      return component.parse(configYaml);
    } catch {
      return component.defaultConfig;
    }
  }, [component, configYaml]);

  const errors = useMemo(() => {
    if (!component || parsed === null) return undefined;
    const result = component.configSchema.safeParse(parsed);
    return result.success ? undefined : flattenZodErrors(result.error);
  }, [component, parsed]);

  if (!component) {
    return (
      <p className="text-sm text-destructive">
        Unknown component: {category}:{componentId}
      </p>
    );
  }

  const Editor = component.Editor;

  const handleChange = (next: unknown) => {
    const yaml = component.serialize(next as never);
    onChange(yaml);
  };

  return (
    <Suspense
      fallback={
        <p className="text-sm text-muted-foreground">Loading editor...</p>
      }
    >
      <Editor
        value={parsed as never}
        onChange={handleChange}
        errors={errors}
      />
    </Suspense>
  );
}
