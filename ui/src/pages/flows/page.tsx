import { useEffect, useState, useMemo, lazy, Suspense } from "react";
import { Button } from "@/components/ui/button";
import {
  Plus,
  Eye,
  Play,
  Pause,
  RotateCcw,
  ScrollText,
  ChevronRight,
  Package,
  Trash2,
  RefreshCw,
  PencilIcon,
} from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/components/toast";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Flow } from "@/lib/entities";
import { columns as getColumns } from "@/components/flow-columns";
import { fetchFlows, updateFlowStatus, deleteFlow } from "@/lib/api";

const FlowPreview = lazy(
  () => import("@/components/flow-builder/flow-preview"),
);

const MANAGED_BY_LABELS: Record<string, string> = {
  google_calendar: "Google Calendar Template",
  google_drive: "Google Drive Template",
  google_sheets: "Google Sheets Template",
  shopify: "Shopify Template",
};

interface TemplateGroup {
  templateId: string;
  label: string;
  flows: Flow[];
}

function useGroupedFlows(flows: Flow[]) {
  return useMemo(() => {
    const standalone: Flow[] = [];
    const templateMap = new Map<string, Flow[]>();

    for (const flow of flows) {
      if (flow.managed_by) {
        const existing = templateMap.get(flow.managed_by) || [];
        existing.push(flow);
        templateMap.set(flow.managed_by, existing);
      } else {
        standalone.push(flow);
      }
    }

    const templateGroups: TemplateGroup[] = [];
    for (const [templateId, templateFlows] of templateMap) {
      templateGroups.push({
        templateId,
        label: MANAGED_BY_LABELS[templateId] || templateId,
        flows: templateFlows,
      });
    }

    return { standalone, templateGroups };
  }, [flows]);
}

function PackStatusSummary({ flows }: { flows: Flow[] }) {
  const counts: Record<string, number> = {};
  for (const f of flows) {
    counts[f.status] = (counts[f.status] || 0) + 1;
  }

  const colorMap: Record<string, string> = {
    active: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
    completed:
      "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
    paused:
      "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
    failed: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  };

  return (
    <span className="flex items-center gap-1.5">
      {Object.entries(counts).map(([status, count]) => (
        <Badge
          key={status}
          className={colorMap[status] || ""}
          variant="outline"
        >
          {count} {status}
        </Badge>
      ))}
    </span>
  );
}

export default function FlowsPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [flows, setFlows] = useState<Flow[]>([]);
  const [previewFlow, setPreviewFlow] = useState<Flow | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedTemplates, setExpandedPacks] = useState<Set<string>>(new Set());
  const [packToDelete, setPackToDelete] = useState<TemplateGroup | null>(null);
  const [isDeletingTemplate, setIsDeletingTemplate] = useState(false);

  const { standalone, templateGroups } = useGroupedFlows(flows);
  const cols = useMemo(() => getColumns(), []);
  const colCount = cols.length + 1;

  const toggleTemplate = (templateId: string) => {
    setExpandedPacks((prev) => {
      const next = new Set(prev);
      if (next.has(templateId)) {
        next.delete(templateId);
      } else {
        next.add(templateId);
      }
      return next;
    });
  };

  const handleRowClick = (flow: Flow) => {
    navigate(`/flows/${flow.id}/edit`);
  };

  const handleAddNew = () => {
    navigate("/flows/new");
  };

  const handleDelete = async (flow: Flow) => {
    try {
      await deleteFlow(String(flow.id));
      setFlows((prev) => prev.filter((s) => s.id !== flow.id));
      addToast({
        id: "flow-deleted",
        title: "Flow Deleted",
        description: `${flow.name} has been deleted successfully.`,
        variant: "success",
      });
    } catch (err) {
      addToast({
        id: "flow-delete-error",
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to delete flow.",
        variant: "error",
      });
    }
  };

  const handleDeletePack = async (tmpl: TemplateGroup) => {
    setIsDeletingTemplate(true);
    const deletedIds = new Set<string>();

    for (const flow of tmpl.flows) {
      try {
        await deleteFlow(String(flow.id));
        deletedIds.add(flow.id);
      } catch {
        // skip
      }
    }

    setFlows((prev) => prev.filter((f) => !deletedIds.has(f.id)));

    const failed = tmpl.flows.length - deletedIds.size;
    if (failed === 0) {
      addToast({
        id: "tmpl-deleted",
        title: "Template Deleted",
        description: `All ${deletedIds.size} flows from ${tmpl.label} have been deleted.`,
        variant: "success",
      });
    } else {
      addToast({
        id: "tmpl-delete-partial",
        title: "Partial Deletion",
        description: `${deletedIds.size} deleted, ${failed} failed from ${tmpl.label}.`,
        variant: "warning",
      });
    }

    setIsDeletingTemplate(false);
    setPackToDelete(null);
  };

  const handlePreview = (flow: Flow) => {
    setPreviewFlow(flow);
  };

  const handleStatusUpdate = async (flow: Flow, newStatus: string) => {
    if (newStatus === "active" && !flow.is_ready) {
      addToast({
        id: "flow-not-ready",
        title: "Cannot start flow",
        description: `${flow.name} is a draft. Open the flow builder and complete the configuration before starting.`,
        variant: "error",
      });
      return;
    }

    try {
      const updatedFlow = await updateFlowStatus(flow.id, newStatus);
      setFlows((prev) =>
        prev.map((s) => (s.id === flow.id ? updatedFlow : s)),
      );
      addToast({
        id: "flow-status-updated",
        title: "Status Updated",
        description: `${flow.name} status has been updated to ${newStatus}.`,
        variant: "success",
      });
    } catch (err) {
      addToast({
        id: "flow-status-error",
        title: "Failed to update flow status",
        description:
          err instanceof Error ? err.message : "An unknown error occurred.",
        variant: "error",
      });
    }
  };

  useEffect(() => {
    const loadFlows = async () => {
      try {
        setLoading(true);
        const data = await fetchFlows();
        setFlows(data);
        setError(null);
      } catch (err) {
        setError("Failed to fetch flows");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadFlows();
  }, []);

  const renderFlowActions = (flow: Flow) => (
    <div className="flex space-x-2">
      <Button
        variant="ghost"
        size="icon"
        onClick={() => handleRowClick(flow)}
        title="Edit"
      >
        <PencilIcon className="h-4 w-4" />
      </Button>
      {flow.status === "active" && (
        <Button
          variant="ghost"
          size="icon"
          onClick={() => handleStatusUpdate(flow, "paused")}
          title="Pause Flow"
        >
          <Pause className="h-4 w-4" />
        </Button>
      )}
      {flow.status === "paused" && (
        <Button
          variant="ghost"
          size="icon"
          onClick={() => handleStatusUpdate(flow, "active")}
          title="Resume Flow"
        >
          <Play className="h-4 w-4" />
        </Button>
      )}
      {flow.status === "completed" && (
        <Button
          variant="ghost"
          size="icon"
          onClick={() => handleStatusUpdate(flow, "active")}
          title="Restart Flow"
        >
          <RotateCcw className="h-4 w-4" />
        </Button>
      )}
      <Button
        variant="ghost"
        size="icon"
        onClick={() =>
          navigate(`/flows/${flow.parentID || flow.id}/events`)
        }
        title="Events"
      >
        <ScrollText className="h-4 w-4" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        onClick={() => handlePreview(flow)}
        title="Preview"
      >
        <Eye className="h-4 w-4" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        onClick={() => handleDelete(flow)}
        title="Delete"
      >
        <Trash2 className="h-4 w-4" />
      </Button>
    </div>
  );

  const renderFlowRow = (flow: Flow, indent = false) => (
    <TableRow key={flow.id}>
      {cols.map((col) => {
        const value = flow[col.key as keyof Flow];
        return (
          <TableCell
            key={`${flow.id}-${String(col.key)}`}
            className={indent && col.key === "name" ? "pl-10" : ""}
          >
            {col.render
              ? col.render(value as any, flow)
              : String(value ?? "")}
          </TableCell>
        );
      })}
      <TableCell>{renderFlowActions(flow)}</TableCell>
    </TableRow>
  );

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Flows</h1>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => navigate("/flows/templates")}
          >
            <Package className="mr-2 h-4 w-4" />
            Install Template
          </Button>
          <Button onClick={handleAddNew}>
            <Plus className="mr-2 h-4 w-4" />
            Add New
          </Button>
        </div>
      </div>

      {loading ? (
        <p>Loading flows...</p>
      ) : error ? (
        <p className="text-red-500">{error}</p>
      ) : flows.length === 0 ? (
        <div className="rounded-md border">
          <Table>
            <TableBody>
              <TableRow>
                <TableCell
                  colSpan={colCount}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      ) : (
        <div className="space-y-4">
          {templateGroups.map((tmpl) => {
            const isExpanded = expandedTemplates.has(tmpl.templateId);
            return (
              <div key={`tmpl-${tmpl.templateId}`} className="rounded-md border">
                <div
                  className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => toggleTemplate(tmpl.templateId)}
                >
                  <div className="flex items-center gap-3">
                    <ChevronRight
                      className={`h-4 w-4 text-muted-foreground transition-transform ${isExpanded ? "rotate-90" : ""}`}
                    />
                    <Package className="h-4 w-4 text-indigo-500" />
                    <span className="font-semibold">{tmpl.label}</span>
                    <Badge variant="outline">
                      {tmpl.flows.length} tools
                    </Badge>
                    <PackStatusSummary flows={tmpl.flows} />
                  </div>
                  <div
                    className="flex items-center gap-1"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() =>
                        navigate(`/flows/templates?template=${tmpl.templateId}`)
                      }
                      title="Redeploy from Template"
                    >
                      <RefreshCw className="h-4 w-4 mr-1.5" />
                      Redeploy
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setPackToDelete(tmpl)}
                      title="Delete All"
                      className="text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-4 w-4 mr-1.5" />
                      Delete All
                    </Button>
                  </div>
                </div>
                {isExpanded && (
                  <div className="border-t">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          {cols.map((col) => (
                            <TableHead key={String(col.key)}>{col.title}</TableHead>
                          ))}
                          <TableHead className="w-[100px]">Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {tmpl.flows.map((flow) => renderFlowRow(flow))}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </div>
            );
          })}

          {standalone.length > 0 && (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    {cols.map((col) => (
                      <TableHead key={String(col.key)}>{col.title}</TableHead>
                    ))}
                    <TableHead className="w-[100px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {standalone.map((flow) => renderFlowRow(flow))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      )}

      <AlertDialog
        open={!!packToDelete}
        onOpenChange={(open) => !open && setPackToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete entire tmpl?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete all{" "}
              <strong>{packToDelete?.flows.length}</strong> flows from{" "}
              <strong>{packToDelete?.label}</strong>. This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeletingTemplate}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={isDeletingTemplate}
              onClick={(e) => {
                e.preventDefault();
                if (packToDelete) handleDeletePack(packToDelete);
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeletingTemplate ? "Deleting..." : "Delete All"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Dialog
        open={!!previewFlow}
        onOpenChange={(open: boolean) => !open && setPreviewFlow(null)}
      >
        <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col">
          <DialogHeader className="flex-shrink-0">
            <DialogTitle>{previewFlow?.name} - Visual Preview</DialogTitle>
          </DialogHeader>
          <div className="flex-1 min-h-0 overflow-y-auto">
            <Suspense
              fallback={
                <div className="flex items-center justify-center h-32">
                  Loading Preview...
                </div>
              }
            >
              {previewFlow && <FlowPreview flow={previewFlow} />}
            </Suspense>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
