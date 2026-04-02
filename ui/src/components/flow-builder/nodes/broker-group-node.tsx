import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Plus, Split } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const BrokerGroupNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const disconnected = d.disconnected;

  return (
    <div
      className={`relative rounded-lg border-2 border-dashed bg-card/30 ${
        disconnected
          ? "ring-2 ring-destructive animate-pulse border-destructive"
          : selected
            ? "ring-2 ring-primary border-amber-400 dark:border-amber-700"
            : "border-amber-300 dark:border-amber-800"
      }`}
      style={{ width: "100%", height: "100%" }}
    >
      <Handle
        type="target"
        position={Position.Left}
        className="!w-3 !h-3 !bg-amber-500 !border-2 !border-background"
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

      {/* Header */}
      <div className="px-3 py-2 border-b border-dashed border-amber-300 dark:border-amber-800">
        <div className="flex items-center gap-2">
          <Split className="h-3.5 w-3.5 text-amber-500" />
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
            Broker
          </span>
          {(d as any).childCount > 0 && (
            <span className="text-[10px] text-muted-foreground">
              · {(d as any).childCount} output{(d as any).childCount !== 1 ? "s" : ""}
            </span>
          )}
        </div>
        <p className="text-sm font-medium truncate">{d.label || "broker"}</p>
        {(d as any).brokerPattern && (
          <span className="text-[9px] font-mono text-muted-foreground/70">
            {(d as any).brokerPattern}
          </span>
        )}
      </div>

      {!(d as any).readOnly && <button
        type="button"
        className="absolute bottom-2 left-1/2 -translate-x-1/2 flex items-center justify-center w-6 h-6 rounded-full border border-dashed border-muted-foreground/40 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        onClick={(e) => {
          e.stopPropagation();
          (d as any).onAddChildOutput?.(d.nodeId);
        }}
        title="Add output"
      >
        <Plus className="h-3.5 w-3.5" />
      </button>}
    </div>
  );
});

BrokerGroupNode.displayName = "BrokerGroupNode";
