import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Upload } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const ChildOutputNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const componentName = d.componentId || "Output";
  const hasCustomLabel = d.label && d.label !== componentName;
  const configPreview = d.configYaml?.trim()
    ? d.configYaml.trim().split("\n")[0].slice(0, 30)
    : "";
  const showRotationBadge = d.parentBrokerPattern === "round_robin" || d.parentBrokerPattern === "greedy";

  return (
    <div
      className={`relative rounded-md border bg-card text-card-foreground shadow-sm w-[160px] border-l-[3px] border-l-amber-400 cursor-grab active:cursor-grabbing ${
        selected ? "ring-2 ring-primary" : ""
      }`}
    >
      {showRotationBadge && (
        <div className="absolute -top-2 -right-2 flex items-center justify-center w-5 h-5 rounded-full bg-amber-500 text-white text-[10px] font-bold shadow-sm z-10">
          {(d.childIndex ?? 0) + 1}
        </div>
      )}

      <Handle
        type="target"
        position={Position.Top}
        className="!w-0 !h-0 !border-0 !bg-transparent !min-w-0 !min-h-0"
        isConnectable={false}
      />

      <div className="px-2 py-1.5">
        <div className="flex items-center gap-1.5 mb-0.5">
          <Upload className="h-3 w-3 text-amber-400" />
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

ChildOutputNode.displayName = "ChildOutputNode";
