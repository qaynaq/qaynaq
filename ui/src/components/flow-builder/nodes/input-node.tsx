import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Plus, Download } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const InputNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const disconnected = d.disconnected;

  return (
    <div
      className={`relative rounded-lg border bg-card text-card-foreground shadow-sm min-w-[180px] max-w-[220px] border-l-4 border-l-green-500 ${
        disconnected
          ? "ring-2 ring-destructive animate-pulse"
          : selected
            ? "ring-2 ring-primary"
            : ""
      }`}
    >
      <div className="p-3">
        <div className="flex items-center gap-2 mb-1">
          <Download className="h-3.5 w-3.5 text-green-500" />
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Input
          </span>
        </div>
        <p className="text-sm font-medium truncate">{d.label || "new_input"}</p>
        {(d.component || d.componentId) && (
          <p className="text-xs text-muted-foreground truncate mt-0.5">
            {d.component || d.componentId}
          </p>
        )}
      </div>

      <Handle type="source" position={Position.Right} className="!w-3 !h-3 !bg-green-500 !border-2 !border-background" />

      {!d.readOnly && <button
        type="button"
        className="absolute -right-3 top-1/2 translate-x-full -translate-y-1/2 flex items-center justify-center w-6 h-6 rounded-full bg-primary text-primary-foreground shadow-md hover:scale-110 transition-transform"
        onClick={(e) => {
          e.stopPropagation();
          d.onAddAndConnect?.(d.nodeId, "input");
        }}
        title="Add processor"
      >
        <Plus className="h-3.5 w-3.5" />
      </button>}
    </div>
  );
});

InputNode.displayName = "InputNode";
