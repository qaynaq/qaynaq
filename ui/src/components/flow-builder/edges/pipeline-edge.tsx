import { memo } from "react";
import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
  type EdgeProps,
} from "@xyflow/react";
import { X, GitBranch } from "lucide-react";

export const PipelineEdge = memo(({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  data,
}: EdgeProps) => {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const onDelete = data?.onDeleteEdge as ((id: string) => void) | undefined;
  const isInternal = data?.internal as boolean;
  const onEditCaseNode = data?.onEditCaseNode as ((nodeId: string) => void) | undefined;
  const caseTargetId = data?.switchCase ? (data.caseTargetId as string | undefined) : undefined;

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          ...style,
          strokeWidth: isInternal ? 1.5 : 2,
          stroke: isInternal
            ? "hsl(var(--muted-foreground) / 0.4)"
            : "hsl(var(--muted-foreground))",
          ...(isInternal
            ? {}
            : { strokeDasharray: "6 3", animation: "dashmove 0.5s linear infinite" }),
        }}
      />
      <EdgeLabelRenderer>
        {typeof data?.edgeLabel === "string" && (
          <div
            style={{
              position: "absolute",
              transform: `translate(-50%, -100%) translate(${labelX}px,${labelY - 4}px)`,
              pointerEvents: onEditCaseNode && caseTargetId ? "all" : "none",
              cursor: onEditCaseNode && caseTargetId ? "pointer" : "default",
            }}
            className="nodrag nopan"
            onDoubleClick={onEditCaseNode && caseTargetId ? () => onEditCaseNode(caseTargetId) : undefined}
            title={onEditCaseNode && caseTargetId ? "Double-click to edit case" : undefined}
          >
            <span className={`inline-flex items-center gap-1 text-[9px] font-mono px-1.5 py-0.5 rounded border whitespace-nowrap ${
              data.edgeLabelColor === "amber"
                ? "bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-300 border-amber-200 dark:border-amber-800"
                : "bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-300 border-violet-200 dark:border-violet-800"
            }`}>
              <GitBranch className="h-2.5 w-2.5 shrink-0" />
              {data.edgeLabel as string}
            </span>
          </div>
        )}
        {onDelete && !isInternal && (
          <div
            style={{
              position: "absolute",
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              pointerEvents: "all",
            }}
            className="nodrag nopan"
          >
            <button
              type="button"
              className="flex items-center justify-center w-5 h-5 rounded-full bg-muted border border-border text-muted-foreground shadow-sm hover:bg-destructive hover:text-destructive-foreground hover:border-destructive transition-colors"
              onClick={() => onDelete(id)}
              title="Delete connection"
            >
              <X className="h-3 w-3" />
            </button>
          </div>
        )}
      </EdgeLabelRenderer>
    </>
  );
});

PipelineEdge.displayName = "PipelineEdge";
