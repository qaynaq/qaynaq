import { useNavigate, useSearchParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { Loader2, BrainCircuit, ArrowRight, ArrowLeft, Plus, Trash2, Workflow, Package, Check, CircleAlert } from "lucide-react";
import { useToast } from "@/components/toast";
import { FlowBuilder } from "@/components/flow-builder/flow-builder";
import { createFlow, deleteFlow, validateFlow, tryFlow, fetchSecrets, fetchConnections, fetchFlows } from "@/lib/api";
import {
  componentSchemas as rawComponentSchemas,
  componentLists
} from "@/lib/component-schemas";
import type { AllComponentSchemas } from "@/components/flow-builder/node-config-panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import * as yaml from "js-yaml";
import { templatePacks, type TemplatePack, type McpToolTemplate, type SharedConfigField } from "@/lib/mcp-tool-templates";
import { buildFlowFromTemplate } from "@/lib/flow-builder-utils";

export interface StreamNodeData {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
  status?: string;
}

type FlowType = null | "mcp_tool" | "automation" | "template_pack";

interface McpParameter {
  name: string;
  type: string;
  required: boolean;
  description: string;
}

const PARAMETER_TYPES = ["string", "number", "boolean", "array", "object"];

const transformComponentSchemas = (): AllComponentSchemas => {
  const allSchemas: AllComponentSchemas = {
    input: [],
    processor: [],
    output: [],
  };

  for (const typeKey of ["input", "pipeline", "output"] as const) {
    const list = componentLists[typeKey] || [];
    const targetTypeForApp = typeKey === 'pipeline' ? 'processor' : typeKey;

    let schemaCategory: typeof rawComponentSchemas.input | typeof rawComponentSchemas.pipeline | typeof rawComponentSchemas.output | undefined;
    if (typeKey === 'input') schemaCategory = rawComponentSchemas.input;
    else if (typeKey === 'pipeline') schemaCategory = rawComponentSchemas.pipeline;
    else if (typeKey === 'output') schemaCategory = rawComponentSchemas.output;

    list.forEach((componentName: string) => {
      const rawSchema = schemaCategory?.[componentName as keyof typeof schemaCategory];
      if (rawSchema) {
        allSchemas[targetTypeForApp].push({
          id: componentName,
          name: (rawSchema as any).title || componentName,
          component: componentName,
          type: targetTypeForApp,
          schema: rawSchema,
        });
      }
    });
  }
  return allSchemas;
};

function FlowTypeSelector({ onSelect }: { onSelect: (type: FlowType) => void }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <div className="mb-8 text-center">
        <h1 className="text-2xl font-bold">Create New Flow</h1>
        <p className="text-muted-foreground mt-1">
          What would you like to build?
        </p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl w-full px-6">
        <button
          onClick={() => onSelect("mcp_tool")}
          className="group relative flex flex-col items-start p-8 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
        >
          <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary mb-4 transition-transform duration-300 group-hover:scale-110 group-hover:rotate-6">
            <BrainCircuit className="h-6 w-6" />
          </div>
          <h2 className="text-lg font-semibold mb-2">MCP Tool</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Build a tool for AI assistants. Define parameters and connect to any API, database, or service - instantly callable by Claude, Cursor, and AI agents.
          </p>
          <ArrowRight className="absolute top-8 right-6 h-5 w-5 text-muted-foreground/40 group-hover:text-primary transition-colors" />
        </button>

        <button
          onClick={() => onSelect("template_pack")}
          className="group relative flex flex-col items-start p-8 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
        >
          <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary mb-4 transition-transform duration-300 group-hover:scale-110 group-hover:rotate-6">
            <Package className="h-6 w-6" />
          </div>
          <h2 className="text-lg font-semibold mb-2">MCP Tool Pack</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Deploy a pre-built set of MCP tools for a service. Configure once and create multiple tools at once - Google Calendar, and more coming soon.
          </p>
          <ArrowRight className="absolute top-8 right-6 h-5 w-5 text-muted-foreground/40 group-hover:text-primary transition-colors" />
        </button>

        <button
          onClick={() => onSelect("automation")}
          className="group relative flex flex-col items-start p-8 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
        >
          <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary mb-4 transition-transform duration-300 group-hover:scale-110 group-hover:-rotate-6">
            <Workflow className="h-6 w-6" />
          </div>
          <h2 className="text-lg font-semibold mb-2">Automation</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Move, transform, and route data between systems or orchestrate AI workflows that call your MCP tools. Connect 66+ sources and destinations with any trigger.
          </p>
          <ArrowRight className="absolute top-8 right-6 h-5 w-5 text-muted-foreground/40 group-hover:text-primary transition-colors" />
        </button>
      </div>
    </div>
  );
}

function McpToolForm({ onBack, onContinue }: {
  onBack: () => void;
  onContinue: (data: { name: string; description: string; parameters: McpParameter[] }) => void;
}) {
  const [toolName, setToolName] = useState("");
  const [description, setDescription] = useState("");
  const [parameters, setParameters] = useState<McpParameter[]>([]);
  const [nameError, setNameError] = useState("");

  const addParameter = () => {
    setParameters([...parameters, { name: "", type: "string", required: false, description: "" }]);
  };

  const updateParameter = (index: number, field: keyof McpParameter, value: any) => {
    const updated = [...parameters];
    updated[index] = { ...updated[index], [field]: value };
    setParameters(updated);
  };

  const removeParameter = (index: number) => {
    setParameters(parameters.filter((_, i) => i !== index));
  };

  const validateName = (value: string) => {
    if (value && !/^[a-zA-Z0-9_-]*$/.test(value)) {
      setNameError("Only letters, numbers, underscores, and hyphens allowed");
    } else {
      setNameError("");
    }
  };

  const canContinue = toolName.trim() !== "" && description.trim() !== "" && !nameError;

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <div className="w-full max-w-xl px-6">
        <button
          onClick={onBack}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </button>

        <div className="mb-8">
          <h1 className="text-2xl font-bold">Configure MCP Tool</h1>
          <p className="text-muted-foreground mt-1">
            Define how AI assistants will discover and call this tool
          </p>
        </div>

        <div className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="tool-name">Tool Name</Label>
            <Input
              id="tool-name"
              value={toolName}
              onChange={(e) => {
                setToolName(e.target.value);
                validateName(e.target.value);
              }}
              placeholder="e.g. query_customers, check_inventory"
              className={nameError ? "border-destructive" : ""}
            />
            {nameError && <p className="text-xs text-destructive">{nameError}</p>}
            <p className="text-xs text-muted-foreground">
              The identifier AI assistants use to call this tool
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="tool-description">Description</Label>
            <Textarea
              id="tool-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="e.g. Query the customer database by name, email, or ID and return matching records"
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              Helps AI assistants understand when and how to use this tool
            </p>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>Input Parameters</Label>
              <Button
                variant="outline"
                size="sm"
                onClick={addParameter}
                className="h-7 text-xs"
              >
                <Plus className="h-3 w-3 mr-1" />
                Add Parameter
              </Button>
            </div>

            {parameters.length === 0 && (
              <p className="text-sm text-muted-foreground py-4 text-center border border-dashed rounded-lg">
                No parameters defined yet. Add parameters that AI assistants will pass when calling this tool.
              </p>
            )}

            {parameters.map((param, index) => (
              <div
                key={index}
                className="flex items-start gap-3 p-3 rounded-lg border bg-muted/30"
              >
                <div className="flex-1 space-y-2">
                  <div className="grid grid-cols-[1fr_120px] gap-2">
                    <Input
                      value={param.name}
                      onChange={(e) => updateParameter(index, "name", e.target.value)}
                      placeholder="Parameter name"
                      className="h-8 text-sm"
                    />
                    <Select
                      value={param.type}
                      onValueChange={(val) => updateParameter(index, "type", val)}
                    >
                      <SelectTrigger className="h-8 text-sm">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {PARAMETER_TYPES.map((t) => (
                          <SelectItem key={t} value={t}>{t}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <Input
                    value={param.description}
                    onChange={(e) => updateParameter(index, "description", e.target.value)}
                    placeholder="Description"
                    className="h-8 text-sm"
                  />
                  <div className="flex items-center gap-2">
                    <Checkbox
                      checked={param.required}
                      onCheckedChange={(checked) => updateParameter(index, "required", !!checked)}
                      className="h-3.5 w-3.5"
                    />
                    <span className="text-xs text-muted-foreground">Required</span>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => removeParameter(index)}
                  className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>

          <Button
            onClick={() => onContinue({ name: toolName.trim(), description: description.trim(), parameters })}
            disabled={!canContinue}
            className="w-full"
            size="lg"
          >
            Continue to Flow Builder
            <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </div>
    </div>
  );
}

function TemplatePackWizard({ onBack, initialPackId }: { onBack: () => void; initialPackId?: string }) {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [selectedPack, setSelectedPack] = useState<TemplatePack | null>(null);
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set());
  const [sharedConfig, setSharedConfig] = useState<Record<string, string>>({});
  const [secrets, setSecrets] = useState<string[]>([]);
  const [connections, setConnections] = useState<string[]>([]);
  const [secretsLoading, setSecretsLoading] = useState(false);
  const [isDeploying, setIsDeploying] = useState(false);
  const [deployResults, setDeployResults] = useState<Array<{ name: string; success: boolean; error?: string }> | null>(null);
  const [existingManagedFlows, setExistingManagedFlows] = useState<Map<string, Map<string, string>>>(new Map());
  const [overrideMode, setOverrideMode] = useState(false);

  useEffect(() => {
    const loadData = async () => {
      setSecretsLoading(true);
      try {
        const [secretsData, connectionsData, flowsData] = await Promise.all([
          fetchSecrets().catch(() => []),
          fetchConnections().catch(() => []),
          fetchFlows().catch(() => []),
        ]);
        setSecrets(secretsData.map((s) => s.key));
        setConnections(connectionsData.map((c) => c.name));
        const managed = new Map<string, Map<string, string>>();
        for (const flow of flowsData) {
          if (flow.managed_by) {
            if (!managed.has(flow.managed_by)) {
              managed.set(flow.managed_by, new Map());
            }
            managed.get(flow.managed_by)!.set(flow.name, flow.id);
          }
        }
        setExistingManagedFlows(managed);
      } finally {
        setSecretsLoading(false);
      }
    };
    loadData();
  }, []);

  useEffect(() => {
    if (initialPackId && !secretsLoading && !selectedPack) {
      const pack = templatePacks.find((p) => p.id === initialPackId);
      if (pack) handleSelectPack(pack);
    }
  }, [initialPackId, secretsLoading]);

  const handleSelectPack = (pack: TemplatePack) => {
    setSelectedPack(pack);
    const packFlows = existingManagedFlows.get(pack.id);
    const deployable = pack.templates.filter((t) => !packFlows?.has(t.name));
    setSelectedTools(new Set(deployable.map((t) => t.id)));
    const defaults: Record<string, string> = {};
    for (const field of pack.sharedConfig) {
      defaults[field.key] = field.default || "";
    }
    setSharedConfig(defaults);
  };

  const isAlreadyDeployed = (template: McpToolTemplate) => {
    if (!selectedPack) return false;
    return existingManagedFlows.get(selectedPack.id)?.has(template.name) ?? false;
  };

  const deployableTemplates = selectedPack
    ? selectedPack.templates.filter((t) => overrideMode || !isAlreadyDeployed(t))
    : [];

  const toggleTool = (id: string) => {
    setSelectedTools((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (!selectedPack) return;
    if (selectedTools.size === deployableTemplates.length) {
      setSelectedTools(new Set());
    } else {
      setSelectedTools(new Set(deployableTemplates.map((t) => t.id)));
    }
  };

  const canDeploy = () => {
    if (!selectedPack || selectedTools.size === 0) return false;
    for (const field of selectedPack.sharedConfig) {
      if (field.required && !sharedConfig[field.key]) return false;
    }
    return true;
  };

  const handleDeploy = async () => {
    if (!selectedPack) return;
    setIsDeploying(true);
    const results: Array<{ name: string; success: boolean; error?: string }> = [];

    const templates = selectedPack.templates.filter((t) => selectedTools.has(t.id));

    const packFlowMap = existingManagedFlows.get(selectedPack.id);

    for (const template of templates) {
      try {
        const existingFlowId = packFlowMap?.get(template.name);
        if (existingFlowId && !overrideMode) {
          continue;
        }
        if (existingFlowId && overrideMode) {
          await deleteFlow(existingFlowId);
        }
        const flowData = buildFlowFromTemplate(template, sharedConfig, selectedPack.sharedConfig, selectedPack.id);
        await createFlow(flowData);
        results.push({ name: template.name, success: true });
      } catch (error) {
        results.push({
          name: template.name,
          success: false,
          error: error instanceof Error ? error.message : "Unknown error",
        });
      }
    }

    setDeployResults(results);
    setIsDeploying(false);

    const newNames = results.filter((r) => r.success).map((r) => r.name);
    if (newNames.length > 0 && selectedPack) {
      setExistingManagedFlows((prev) => {
        const next = new Map(prev);
        const existing = next.get(selectedPack.id) ?? new Map<string, string>();
        const updated = new Map(existing);
        for (const name of newNames) updated.set(name, "");
        next.set(selectedPack.id, updated);
        return next;
      });
    }

    const successCount = results.filter((r) => r.success).length;
    const failCount = results.filter((r) => !r.success).length;

    if (failCount === 0) {
      addToast({
        id: "template-deploy-success",
        title: "Tools Deployed",
        description: `${successCount} MCP tool${successCount > 1 ? "s" : ""} created successfully.`,
        variant: "success",
      });
    } else {
      addToast({
        id: "template-deploy-partial",
        title: "Deployment Completed",
        description: `${successCount} succeeded, ${failCount} failed.`,
        variant: failCount === results.length ? "error" : "warning",
      });
    }
  };

  if (deployResults) {
    const successCount = deployResults.filter((r) => r.success).length;
    const failCount = deployResults.filter((r) => !r.success).length;

    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh]">
        <div className="w-full max-w-xl px-6">
          <div className="mb-8 text-center">
            <div className={`flex items-center justify-center w-16 h-16 rounded-full mx-auto mb-4 ${failCount === 0 ? "bg-green-500/10 text-green-500" : "bg-yellow-500/10 text-yellow-500"}`}>
              {failCount === 0 ? <Check className="h-8 w-8" /> : <CircleAlert className="h-8 w-8" />}
            </div>
            <h1 className="text-2xl font-bold">
              {failCount === 0 ? "All Tools Deployed" : "Deployment Complete"}
            </h1>
            <p className="text-muted-foreground mt-1">
              {successCount} tool{successCount !== 1 ? "s" : ""} created
              {failCount > 0 ? `, ${failCount} failed` : ""}
            </p>
          </div>

          <div className="space-y-2 mb-8">
            {deployResults.map((result) => (
              <div
                key={result.name}
                className={`flex items-center justify-between p-3 rounded-lg border ${result.success ? "bg-green-500/5 border-green-500/20" : "bg-destructive/5 border-destructive/20"}`}
              >
                <span className="text-sm font-mono">{result.name}</span>
                {result.success ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <span className="text-xs text-destructive">{result.error}</span>
                )}
              </div>
            ))}
          </div>

          <div className="flex gap-3">
            <Button
              variant="outline"
              className="flex-1"
              onClick={() => {
                setDeployResults(null);
                setSelectedPack(null);
                setSelectedTools(new Set());
              }}
            >
              Deploy Another Pack
            </Button>
            <Button className="flex-1" onClick={() => navigate("/flows")}>
              Go to Flows
            </Button>
          </div>
        </div>
      </div>
    );
  }

  if (isDeploying) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        <p className="text-muted-foreground">Deploying MCP tools...</p>
      </div>
    );
  }

  if (!selectedPack) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh]">
        <div className="w-full max-w-2xl px-6">
          <button
            onClick={onBack}
            className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-6 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </button>

          <div className="mb-8">
            <h1 className="text-2xl font-bold">MCP Tool Packs</h1>
            <p className="text-muted-foreground mt-1">
              Select a pack to deploy pre-built MCP tools for a service
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4">
            {templatePacks.map((pack) => (
              <button
                key={pack.id}
                onClick={() => handleSelectPack(pack)}
                className="group relative flex items-start gap-4 p-6 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
              >
                <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary shrink-0 transition-transform duration-300 group-hover:scale-110">
                  <Package className="h-6 w-6" />
                </div>
                <div className="flex-1 min-w-0">
                  <h2 className="text-lg font-semibold mb-1">{pack.name}</h2>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    {pack.description}
                  </p>
                </div>
                <ArrowRight className="h-5 w-5 text-muted-foreground/40 group-hover:text-primary transition-colors mt-1 shrink-0" />
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <div className="w-full max-w-2xl px-6">
        <button
          onClick={() => {
            setSelectedPack(null);
            setSelectedTools(new Set());
          }}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Packs
        </button>

        <div className="mb-8">
          <h1 className="text-2xl font-bold">Deploy {selectedPack.name} Tools</h1>
          <p className="text-muted-foreground mt-1">
            Configure shared settings and select which tools to create
          </p>
        </div>

        <div className="space-y-8">
          <div className="space-y-4">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Configuration
            </h2>
            {selectedPack.sharedConfig.map((field) => (
              <div key={field.key} className="space-y-2">
                <Label>
                  {field.title}
                  {field.required && <span className="text-destructive ml-1">*</span>}
                </Label>
                {field.type === "dynamic_select" && field.dataSource === "secrets" ? (
                  secretsLoading ? (
                    <div className="text-sm text-muted-foreground">Loading secrets...</div>
                  ) : (
                    <Select
                      value={sharedConfig[field.key]?.replace(/^\$\{(.+)\}$/, "$1") || ""}
                      onValueChange={(val) =>
                        setSharedConfig((prev) => ({ ...prev, [field.key]: `\${${val}}` }))
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={secrets.length === 0 ? "No secrets available" : "Select a secret..."} />
                      </SelectTrigger>
                      <SelectContent>
                        {secrets.map((secret) => (
                          <SelectItem key={secret} value={secret}>
                            {secret}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )
                ) : field.type === "dynamic_select" && field.dataSource === "connections" ? (
                  secretsLoading ? (
                    <div className="text-sm text-muted-foreground">Loading connections...</div>
                  ) : (
                    <Select
                      value={sharedConfig[field.key]?.replace(/^\$\{QAYNAQ_CONN_(.+)\}$/, "$1") || ""}
                      onValueChange={(val) =>
                        setSharedConfig((prev) => ({ ...prev, [field.key]: `\${QAYNAQ_CONN_${val}}` }))
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={connections.length === 0 ? "No connections available" : "Select a connection..."} />
                      </SelectTrigger>
                      <SelectContent>
                        {connections.map((conn) => (
                          <SelectItem key={conn} value={conn}>
                            {conn}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )
                ) : (
                  <Input
                    value={sharedConfig[field.key] || ""}
                    onChange={(e) =>
                      setSharedConfig((prev) => ({ ...prev, [field.key]: e.target.value }))
                    }
                    placeholder={field.description}
                  />
                )}
                <p className="text-xs text-muted-foreground">{field.description}</p>
              </div>
            ))}
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
                Tools ({selectedTools.size} of {deployableTemplates.length} available)
              </h2>
              <div className="flex items-center gap-3">
                {(existingManagedFlows.get(selectedPack.id)?.size ?? 0) > 0 && (
                  <label className="flex items-center gap-2 cursor-pointer">
                    <Checkbox
                      checked={overrideMode}
                      onCheckedChange={(checked) => {
                        setOverrideMode(!!checked);
                        if (!checked) {
                          setSelectedTools((prev) => {
                            const next = new Set(prev);
                            for (const t of selectedPack!.templates) {
                              if (isAlreadyDeployed(t)) next.delete(t.id);
                            }
                            return next;
                          });
                        }
                      }}
                    />
                    <span className="text-xs text-muted-foreground">Override existing</span>
                  </label>
                )}
                <Button variant="ghost" size="sm" onClick={toggleAll} className="h-7 text-xs">
                  {selectedTools.size === deployableTemplates.length ? "Deselect All" : "Select All"}
                </Button>
              </div>
            </div>

            {overrideMode && (
              <div className="flex items-start gap-2 p-3 rounded-lg border border-yellow-500/30 bg-yellow-500/5 text-sm">
                <CircleAlert className="h-4 w-4 text-yellow-500 shrink-0 mt-0.5" />
                <span className="text-yellow-700 dark:text-yellow-400">
                  Override mode will delete existing tools and recreate them with the current configuration. Any manual edits to those tools will be lost.
                </span>
              </div>
            )}

            {deployableTemplates.length === 0 && selectedPack.templates.length > 0 && (
              <div className="text-sm text-muted-foreground py-6 text-center border border-dashed rounded-lg">
                All tools from this pack are already deployed. Enable "Override existing" to redeploy them.
              </div>
            )}

            <div className="space-y-2">
              {selectedPack.templates.map((template) => {
                const deployed = isAlreadyDeployed(template);
                const selectable = overrideMode || !deployed;
                return (
                  <label
                    key={template.id}
                    className={`flex items-start gap-3 p-3 rounded-lg border transition-colors ${
                      !selectable
                        ? "border-border bg-muted/50 cursor-default opacity-60"
                        : selectedTools.has(template.id)
                          ? "border-primary/50 bg-primary/5 cursor-pointer"
                          : "border-border bg-card hover:border-border/80 cursor-pointer"
                    }`}
                  >
                    <Checkbox
                      checked={selectedTools.has(template.id)}
                      onCheckedChange={() => selectable && toggleTool(template.id)}
                      disabled={!selectable}
                      className="mt-0.5"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-mono font-medium">{template.name}</span>
                        {deployed && (
                          <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${
                            overrideMode && selectedTools.has(template.id)
                              ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300"
                              : "bg-muted text-muted-foreground"
                          }`}>
                            {overrideMode && selectedTools.has(template.id) ? "Will override" : "Already deployed"}
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground mt-0.5">
                        {template.description}
                      </div>
                      {selectable && (
                        <div className="flex flex-wrap gap-1 mt-1.5">
                          {template.parameters.filter((p) => p.required).map((p) => (
                            <span
                              key={p.name}
                              className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-muted text-muted-foreground"
                            >
                              {p.name}
                            </span>
                          ))}
                          {template.parameters.filter((p) => !p.required).length > 0 && (
                            <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] text-muted-foreground/60">
                              +{template.parameters.filter((p) => !p.required).length} optional
                            </span>
                          )}
                        </div>
                      )}
                    </div>
                  </label>
                );
              })}
            </div>
          </div>

          <Button
            onClick={handleDeploy}
            disabled={!canDeploy()}
            className="w-full"
            size="lg"
          >
            Deploy {selectedTools.size} Tool{selectedTools.size !== 1 ? "s" : ""}
            <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </div>
    </div>
  );
}

export default function NewStreamPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const packParam = searchParams.get("pack");
  const { addToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [transformedSchemas, setTransformedSchemas] = useState<AllComponentSchemas | null>(null);
  const [flowType, setFlowType] = useState<FlowType>(packParam ? "template_pack" : null);
  const [mcpInitialData, setMcpInitialData] = useState<{
    name: string;
    status: string;
    nodes: StreamNodeData[];
  } | null>(null);

  useEffect(() => {
    setTransformedSchemas(transformComponentSchemas());
  }, []);

  const handleMcpContinue = (data: { name: string; description: string; parameters: McpParameter[] }) => {
    const inputSchema = data.parameters.map((p) => ({
      name: p.name,
      type: p.type,
      required: p.required,
      description: p.description,
    }));

    const configObj: Record<string, any> = {
      name: data.name,
      description: data.description,
    };
    if (inputSchema.length > 0) {
      configObj.input_schema = inputSchema;
    }
    const configYaml = yaml.dump(configObj, { lineWidth: -1, noRefs: true });

    setMcpInitialData({
      name: data.name,
      status: "active",
      nodes: [
        {
          label: data.name,
          type: "input",
          componentId: "mcp_tool",
          component: "mcp_tool",
          configYaml,
        },
        {
          label: "mcp_tool_response",
          type: "output",
          componentId: "sync_response",
          component: "sync_response",
          configYaml: "",
        },
      ],
    });
  };

  const handleValidateStream = async (data: { name: string; status: string; bufferId?: number; nodes: StreamNodeData[] }) => {
    const inputNode = data.nodes.find((node) => node.type === "input");
    const processorNodes = data.nodes.filter((node) => node.type === "processor");
    const outputNode = data.nodes.find((node) => node.type === "output");
    if (!inputNode || !outputNode || !inputNode.componentId || !outputNode.componentId) {
      return { valid: false, error: "Stream must have an input and output with components selected." };
    }
    const inputComponent = transformedSchemas?.input.find(c => c.id === inputNode.componentId);
    const outputComponent = transformedSchemas?.output.find(c => c.id === outputNode.componentId);
    if (!inputComponent || !outputComponent) {
      return { valid: false, error: "Selected components not found in available schemas." };
    }
    return validateFlow({
      input_component: inputComponent.component,
      input_label: inputNode.label,
      input_config: inputNode.configYaml || "",
      output_component: outputComponent.component,
      output_label: outputNode.label,
      output_config: outputNode.configYaml || "",
      processors: processorNodes.map(node => {
        const comp = transformedSchemas?.processor.find(c => c.id === node.componentId);
        return { label: node.label, component: comp?.component || node.componentId || "", config: node.configYaml || "" };
      }),
    });
  };

  const handleTryStream = async (data: { processors: Array<{ label: string; component: string; config: string }>; messages: Array<{ content: string }> }) => {
    return tryFlow(data);
  };

  const handleSaveStream = async (data: { name: string; status: string; bufferId?: number; nodes: StreamNodeData[]; builderState: string; isReady: boolean }) => {
    setIsSubmitting(true);
    try {
      const inputNode = data.nodes.find((node) => node.type === "input");
      const processorNodes = data.nodes.filter((node) => node.type === "processor");
      const outputNode = data.nodes.find((node) => node.type === "output");

      if (!inputNode || !outputNode) {
        throw new Error("Stream must have at least one input and one output");
      }

      if (!inputNode.componentId || !outputNode.componentId) {
        throw new Error("Input and output nodes must have components selected");
      }

      const inputComponent = transformedSchemas?.input.find(c => c.id === inputNode.componentId);
      const outputComponent = transformedSchemas?.output.find(c => c.id === outputNode.componentId);

      if (!inputComponent || !outputComponent) {
        throw new Error("Selected components not found in available schemas");
      }

      const processors = processorNodes.map((node) => {
        if (!node.componentId) {
          throw new Error(`Processor node "${node.label}" must have a component selected`);
        }

        const processorComponent = transformedSchemas?.processor.find(c => c.id === node.componentId);
        if (!processorComponent) {
          throw new Error(`Processor component not found for node "${node.label}"`);
        }

        return {
          label: node.label,
          component: processorComponent.component,
          config: node.configYaml || ""
        };
      });

      const streamData = {
        name: data.name,
        status: data.status,
        input_component: inputComponent.component,
        input_label: inputNode.label,
        input_config: inputNode.configYaml || "",
        output_component: outputComponent.component,
        output_label: outputNode.label,
        output_config: outputNode.configYaml || "",
        buffer_id: data.bufferId,
        is_ready: data.isReady,
        builder_state: data.builderState,
        processors: processors
      };

      await createFlow(streamData);
      addToast({
        id: "stream-created",
        title: "Flow Created",
        description: `${data.name} has been created successfully.`,
        variant: "success",
      });
      navigate("/flows");
    } catch (error) {
      console.error("Error creating stream:", error);
      addToast({
        id: "stream-creation-error",
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to create stream.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!transformedSchemas) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (flowType === null) {
    return <FlowTypeSelector onSelect={setFlowType} />;
  }

  if (flowType === "template_pack") {
    return <TemplatePackWizard onBack={() => setFlowType(null)} initialPackId={packParam || undefined} />;
  }

  if (flowType === "mcp_tool" && !mcpInitialData) {
    return (
      <McpToolForm
        onBack={() => setFlowType(null)}
        onContinue={handleMcpContinue}
      />
    );
  }

  if (isSubmitting) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">
          {flowType === "mcp_tool" ? "Build MCP Tool" : "Add New Flow"}
        </h1>
        <p className="text-muted-foreground">
          {flowType === "mcp_tool"
            ? "Add processors to transform data between input and response"
            : "Design your data processing pipeline visually"}
        </p>
      </div>

      <FlowBuilder
        allComponentSchemas={transformedSchemas}
        initialData={mcpInitialData || undefined}
        onSave={handleSaveStream}
        onValidate={handleValidateStream}
        onTry={handleTryStream}
      />
    </div>
  );
}
