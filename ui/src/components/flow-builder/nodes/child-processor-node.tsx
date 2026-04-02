import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Workflow, GripVertical } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const ChildProcessorNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const componentName = d.componentId || "Processor";
  const hasCustomLabel = d.label && d.label !== componentName;
  const configPreview = d.configYaml?.trim()
    ? d.configYaml.trim().split("\n")[0].slice(0, 30)
    : "";
  const orderNum = typeof d.childIndex === "number" ? d.childIndex + 1 : undefined;
  const isDragging = (d as any).isDragging === true;

  return (
    <div
      className={`relative rounded-md border bg-card text-card-foreground w-[160px] border-l-[3px] border-l-blue-400 cursor-grab active:cursor-grabbing transition-all duration-150 ${
        isDragging
          ? "shadow-lg ring-2 ring-primary/50 scale-105 opacity-90 z-50"
          : "shadow-sm"
      } ${selected && !isDragging ? "ring-2 ring-primary" : ""}`}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="!w-0 !h-0 !border-0 !bg-transparent !min-w-0 !min-h-0"
        isConnectable={false}
      />

      {orderNum !== undefined && (
        <span className={`absolute -top-2 -right-2 flex items-center justify-center w-5 h-5 rounded-full border text-[9px] font-semibold z-10 transition-colors duration-150 ${
          isDragging ? "bg-primary text-primary-foreground border-primary" : "bg-muted text-muted-foreground"
        }`}>
          {orderNum}
        </span>
      )}

      <div className="px-2 py-1.5">
        <div className="flex items-center gap-1 mb-0.5">
          <GripVertical className={`h-3 w-3 flex-shrink-0 transition-colors ${isDragging ? "text-primary" : "text-muted-foreground/40"}`} />
          <Workflow className="h-3 w-3 text-blue-400 flex-shrink-0" />
          <span className="text-[9px] uppercase tracking-wider text-muted-foreground font-medium truncate">
            {componentName}
          </span>
        </div>
        {hasCustomLabel && (
          <p className="text-xs font-medium truncate">{d.label}</p>
        )}
        {configPreview ? (
          <p className="text-[9px] text-muted-foreground/70 truncate mt-0.5 font-mono">
            {configPreview}
          </p>
        ) : (
          <p className="text-xs text-muted-foreground truncate">not configured</p>
        )}
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        className="!w-0 !h-0 !border-0 !bg-transparent !min-w-0 !min-h-0"
        isConnectable={false}
      />
    </div>
  );
});

ChildProcessorNode.displayName = "ChildProcessorNode";
