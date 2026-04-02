import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Plus, GitBranch, Upload } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const SwitchCaseStartNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const check = (d as any).caseCheck as string | undefined;
  const hasFallthrough = (d as any).caseFallthrough === true;
  const hasContinue = (d as any).caseContinue === true;
  const isOutput = d.type === "output";
  const componentName = isOutput ? (d.componentId || "") : "";

  return (
    <div
      className={`relative rounded-md border bg-card text-card-foreground shadow-sm border-l-[3px] cursor-grab active:cursor-grabbing ${
        isOutput ? "border-l-amber-400" : "border-l-violet-400"
      } ${
        d.disconnected
          ? "ring-2 ring-destructive animate-pulse"
          : selected ? "ring-2 ring-primary" : ""
      }`}
      style={{ width: 140 }}
    >
      <Handle
        type="target"
        position={Position.Left}
        className={`!w-3 !h-3 !border-2 !border-background ${isOutput ? "!bg-amber-500" : "!bg-violet-500"}`}
      />

      <div className="px-2 py-1.5">
        {isOutput ? (
          <div className="flex items-center gap-1.5">
            <Upload className="h-3 w-3 text-amber-500 shrink-0" />
            <span className="text-[9px] uppercase tracking-wider text-muted-foreground font-medium truncate">
              {componentName || "output"}
            </span>
            {hasContinue && (
              <span className="text-[8px] bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-300 rounded px-1 shrink-0">
                →
              </span>
            )}
          </div>
        ) : (
          <>
            <div className="flex items-center gap-1.5">
              <GitBranch className="h-3 w-3 text-violet-400 shrink-0" />
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-3 shrink-0" />
              {check ? (
                <span className="text-[9px] truncate font-mono flex-1 text-muted-foreground/70">
                  {check}
                </span>
              ) : (
                <span className="text-[9px] truncate italic flex-1 text-violet-400">
                  default
                </span>
              )}
              {hasFallthrough && (
                <span className="text-[8px] bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-300 rounded px-1 shrink-0">
                  fall
                </span>
              )}
            </div>
          </>
        )}
      </div>

      {!isOutput && (
        <Handle
          type="source"
          position={Position.Right}
          className="!w-3 !h-3 !bg-violet-500 !border-2 !border-background"
        />
      )}

      {!(d as any).readOnly && !isOutput && (
        <button
          type="button"
          className="absolute -right-3 top-1/2 translate-x-full -translate-y-1/2 flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-foreground shadow-md hover:scale-110 transition-transform z-10"
          onClick={(e) => {
            e.stopPropagation();
            d.onAddAndConnect?.(d.nodeId, "processor");
          }}
          title="Add processor to case"
        >
          <Plus className="h-3 w-3" />
        </button>
      )}
    </div>
  );
});

SwitchCaseStartNode.displayName = "SwitchCaseStartNode";
