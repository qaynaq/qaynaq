import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { GitBranch } from "lucide-react";
import type { StreamFlowNodeData } from "../flow-builder";

export const ChildCaseNode = memo(({ data, selected }: NodeProps) => {
  const d = data as StreamFlowNodeData;
  const componentName = d.componentId || "Output";
  const check = (d as any).caseCheck as string | undefined;
  const hasContinue = (d as any).caseContinue === true;

  return (
    <div
      className={`relative rounded-md border bg-card text-card-foreground shadow-sm w-[160px] border-l-[3px] border-l-amber-400 cursor-grab active:cursor-grabbing ${
        selected ? "ring-2 ring-primary" : ""
      }`}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="!w-0 !h-0 !border-0 !bg-transparent !min-w-0 !min-h-0"
        isConnectable={false}
      />

      <div className="px-2 py-1.5">
        <div className="flex items-center gap-1.5 mb-0.5">
          <GitBranch className="h-3 w-3 text-amber-500" />
          <span className="text-[9px] uppercase tracking-wider text-muted-foreground font-medium truncate">
            {componentName}
          </span>
          {hasContinue && (
            <span className="text-[8px] bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-300 rounded px-1">
              →
            </span>
          )}
        </div>
        {check ? (
          <p className="text-[9px] text-muted-foreground/70 truncate font-mono">
            {check}
          </p>
        ) : (
          <p className="text-[9px] text-amber-400 truncate italic">
            default
          </p>
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

ChildCaseNode.displayName = "ChildCaseNode";
