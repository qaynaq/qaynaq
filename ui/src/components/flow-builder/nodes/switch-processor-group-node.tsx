import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Plus, GitBranch } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const SwitchProcessorGroupNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const disconnected = d.disconnected;
  const isOutput = d.type === "output";
  const color = isOutput ? "amber" : "violet";

  return (
    <div
      className={`relative rounded-lg border-2 bg-card text-card-foreground shadow-sm border-l-[3px] ${
        isOutput ? "border-l-amber-400" : "border-l-violet-400"
      } ${
        disconnected
          ? "ring-2 ring-destructive animate-pulse"
          : selected
            ? "ring-2 ring-primary"
            : ""
      }`}
      style={{ width: 180 }}
    >
      <Handle
        type="target"
        position={Position.Left}
        className={`!w-3 !h-3 !border-2 !border-background ${isOutput ? "!bg-amber-500" : "!bg-violet-500"}`}
      />

      <Handle
        type="source"
        position={Position.Right}
        className={`!w-3 !h-3 !border-2 !border-background ${isOutput ? "!bg-amber-500" : "!bg-violet-500"}`}
      />

      {!(d as any).readOnly && <button
        type="button"
        className="absolute -left-3 top-1/2 -translate-x-full -translate-y-1/2 flex items-center justify-center w-6 h-6 rounded-full bg-primary text-primary-foreground shadow-md hover:scale-110 transition-transform z-10"
        onClick={(e) => {
          e.stopPropagation();
          d.onAddBefore?.(d.nodeId);
        }}
        title="Add processor before"
      >
        <Plus className="h-3.5 w-3.5" />
      </button>}

      {!isOutput && !(d as any).readOnly && <button
        type="button"
        className="absolute -right-3 top-1/2 translate-x-full -translate-y-1/2 flex items-center justify-center w-6 h-6 rounded-full bg-primary text-primary-foreground shadow-md hover:scale-110 transition-transform z-10"
        onClick={(e) => {
          e.stopPropagation();
          d.onAddAndConnect?.(d.nodeId, "processor");
        }}
        title="Add processor after"
      >
        <Plus className="h-3.5 w-3.5" />
      </button>}

      <div className="px-3 py-2">
        <div className="flex items-center gap-2">
          <GitBranch className={`h-3.5 w-3.5 ${isOutput ? "text-amber-500" : "text-violet-500"}`} />
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Switch
          </span>
          {(d as any).childCount > 0 && (
            <span className="text-[10px] text-muted-foreground">
              · {(d as any).childCount} case{(d as any).childCount !== 1 ? "s" : ""}
            </span>
          )}
        </div>
        <p className="text-sm font-medium truncate">{d.label || "switch"}</p>
      </div>

      {!(d as any).readOnly && <button
        type="button"
        className={`absolute -bottom-3 left-1/2 -translate-x-1/2 translate-y-full flex items-center justify-center w-6 h-6 rounded-full border border-dashed transition-colors z-10 ${
          isOutput
            ? "border-amber-300 text-amber-500 hover:bg-amber-50 dark:hover:bg-amber-900/30"
            : "border-violet-300 text-violet-500 hover:bg-violet-50 dark:hover:bg-violet-900/30"
        }`}
        onClick={(e) => {
          e.stopPropagation();
          (d as any).onAddChildCase?.(d.nodeId);
        }}
        title="Add case"
      >
        <Plus className="h-3.5 w-3.5" />
      </button>}
    </div>
  );
});

SwitchProcessorGroupNode.displayName = "SwitchProcessorGroupNode";
