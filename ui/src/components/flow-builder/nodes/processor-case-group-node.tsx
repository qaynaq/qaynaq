import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Plus, GitBranch } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const ProcessorCaseGroupNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const check = (d as any).caseCheck as string | undefined;
  const hasFallthrough = (d as any).caseFallthrough === true;

  return (
    <div
      className={`relative rounded-md border-2 border-dashed bg-card/40 ${
        selected
          ? "ring-2 ring-primary border-violet-400 dark:border-violet-600"
          : "border-violet-200 dark:border-violet-900"
      }`}
      style={{ width: "100%", height: "100%" }}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="!w-0 !h-0 !border-0 !bg-transparent !min-w-0 !min-h-0"
        isConnectable={false}
      />

      <div className="px-2 py-1 border-b border-dashed border-violet-200 dark:border-violet-900">
        <div className="flex items-center gap-1.5">
          <GitBranch className="h-3 w-3 text-violet-400" />
          {check ? (
            <span className="text-[9px] text-muted-foreground/70 truncate font-mono flex-1">
              {check}
            </span>
          ) : (
            <span className="text-[9px] text-violet-400 truncate italic">
              default
            </span>
          )}
          {hasFallthrough && (
            <span className="text-[8px] bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-300 rounded px-1 shrink-0">
              fall
            </span>
          )}
        </div>
      </div>

      {!(d as any).readOnly && (
        <button
          type="button"
          className="absolute bottom-1 left-1/2 -translate-x-1/2 flex items-center justify-center w-5 h-5 rounded-full border border-dashed border-muted-foreground/30 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
          onClick={(e) => {
            e.stopPropagation();
            (d as any).onAddCaseProcessor?.(d.nodeId);
          }}
          title="Add processor to case"
        >
          <Plus className="h-3 w-3" />
        </button>
      )}

      <Handle
        type="source"
        position={Position.Bottom}
        className="!w-0 !h-0 !border-0 !bg-transparent !min-w-0 !min-h-0"
        isConnectable={false}
      />
    </div>
  );
});

ProcessorCaseGroupNode.displayName = "ProcessorCaseGroupNode";
