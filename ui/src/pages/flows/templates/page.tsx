import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useEffect, useState } from "react";
import {
  Loader2,
  ArrowRight,
  ArrowLeft,
  Package,
  Check,
  CircleAlert,
} from "lucide-react";
import { useToast } from "@/components/toast";
import {
  createSecret,
  fetchTemplates,
  installTemplate,
} from "@/lib/api";
import {
  ConnectionPickerField,
  TextField,
} from "@/components/form-primitives";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import type {
  Template,
  TemplateVariable,
  TemplateInstallResult,
} from "@/lib/entities";

export default function TemplatesPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const templateParam = searchParams.get("template");

  return (
    <TemplateWizard
      onBack={() => navigate("/flows")}
      initialTemplateId={templateParam || undefined}
    />
  );
}
function TemplateVariableField({
  variable,
  value,
  onChange,
}: {
  variable: TemplateVariable;
  value: string;
  onChange: (next: string) => void;
}) {
  const [creatingSecret, setCreatingSecret] = useState(false);
  const [newSecretKey, setNewSecretKey] = useState("");
  const [newSecretValue, setNewSecretValue] = useState("");
  const [secretSaving, setSecretSaving] = useState(false);
  const [secretError, setSecretError] = useState("");
  const [pickerRefresh, setPickerRefresh] = useState(0);

  if (variable.type === "connection") {
    return (
      <div className="space-y-2">
        <ConnectionPickerField
          label={variable.title}
          description={variable.description}
          required={variable.required}
          source="connections"
          value={value ? `\${QAYNAQ_CONN_${value}}` : ""}
          onChange={(next) =>
            onChange(next.replace(/^\$\{QAYNAQ_CONN_(.+)\}$/, "$1"))
          }
        />
        <Link
          to="/connections"
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs text-primary hover:underline"
        >
          + Set up a new connection
        </Link>
      </div>
    );
  }

  if (variable.type === "secret") {
    const handleCreateSecret = async () => {
      const key = newSecretKey.trim();
      if (!key || !newSecretValue) return;
      setSecretSaving(true);
      setSecretError("");
      try {
        await createSecret({ key, value: newSecretValue });
        onChange(key);
        setCreatingSecret(false);
        setNewSecretKey("");
        setNewSecretValue("");
        setPickerRefresh((n) => n + 1);
      } catch (error) {
        setSecretError(
          error instanceof Error ? error.message : "Failed to create secret",
        );
      } finally {
        setSecretSaving(false);
      }
    };

    return (
      <div className="space-y-2">
        <ConnectionPickerField
          key={`${variable.key}-${pickerRefresh}`}
          label={variable.title}
          description={variable.description}
          required={variable.required}
          source="secrets"
          value={value ? `\${${value}}` : ""}
          onChange={(next) => onChange(next.replace(/^\$\{(.+)\}$/, "$1"))}
        />
        {creatingSecret ? (
          <div className="space-y-2 p-3 rounded-lg border border-border bg-muted/30">
            <Input
              value={newSecretKey}
              onChange={(e) =>
                setNewSecretKey(e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, "_"))
              }
              placeholder="SECRET_KEY"
              className="font-mono text-sm"
            />
            <Input
              type="password"
              value={newSecretValue}
              onChange={(e) => setNewSecretValue(e.target.value)}
              placeholder={variable.placeholder || "Secret value"}
              className="font-mono text-sm"
            />
            {secretError && (
              <p className="text-xs text-destructive">{secretError}</p>
            )}
            <div className="flex gap-2">
              <Button
                size="sm"
                onClick={handleCreateSecret}
                disabled={secretSaving || !newSecretKey.trim() || !newSecretValue}
              >
                {secretSaving ? (
                  <Loader2 className="h-3 w-3 animate-spin" />
                ) : (
                  "Save Secret"
                )}
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  setCreatingSecret(false);
                  setSecretError("");
                }}
              >
                Cancel
              </Button>
            </div>
          </div>
        ) : (
          <button
            onClick={() => setCreatingSecret(true)}
            className="text-xs text-primary hover:underline"
          >
            + Create new secret
          </button>
        )}
      </div>
    );
  }

  return (
    <TextField
      label={variable.title}
      description={variable.description}
      required={variable.required}
      value={value}
      onChange={onChange}
      placeholder={variable.placeholder || variable.description}
    />
  );
}

function TemplateWizard({
  onBack,
  initialTemplateId,
}: {
  onBack: () => void;
  initialTemplateId?: string;
}) {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [templateList, setTemplateList] = useState<Template[]>([]);
  const [templatesLoading, setTemplatesLoading] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [selectedFlows, setSelectedFlows] = useState<Set<string>>(new Set());
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [isDeploying, setIsDeploying] = useState(false);
  const [deployResults, setDeployResults] = useState<
    TemplateInstallResult[] | null
  >(null);
  const [overrideMode, setOverrideMode] = useState(false);

  const loadTemplates = async (): Promise<Template[]> => {
    setTemplatesLoading(true);
    try {
      const data = await fetchTemplates();
      setTemplateList(data);
      return data;
    } catch {
      addToast({
        id: "templates-load-error",
        title: "Failed to load templates",
        description: "Could not fetch the template catalog from the server.",
        variant: "error",
      });
      return [];
    } finally {
      setTemplatesLoading(false);
    }
  };

  useEffect(() => {
    loadTemplates();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (initialTemplateId && !templatesLoading && !selectedTemplate) {
      const tmpl = templateList.find((p) => p.id === initialTemplateId);
      if (tmpl) handleSelectTemplate(tmpl);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialTemplateId, templatesLoading, templateList]);

  const handleSelectTemplate = (tmpl: Template) => {
    setSelectedTemplate(tmpl);
    setSelectedFlows(
      new Set(tmpl.flows.filter((f) => !f.installed).map((f) => f.name)),
    );
    const defaults: Record<string, string> = {};
    for (const v of tmpl.variables) {
      defaults[v.key] = v.default || "";
    }
    setVariables(defaults);
    setOverrideMode(false);
  };

  const deployableFlows = selectedTemplate
    ? selectedTemplate.flows.filter((f) => overrideMode || !f.installed)
    : [];

  const hasInstalledFlows =
    selectedTemplate?.flows.some((f) => f.installed) ?? false;

  const toggleFlow = (name: string) => {
    setSelectedFlows((prev) => {
      const next = new Set(prev);
      if (next.has(name)) {
        next.delete(name);
      } else {
        next.add(name);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (!selectedTemplate) return;
    if (selectedFlows.size === deployableFlows.length) {
      setSelectedFlows(new Set());
    } else {
      setSelectedFlows(new Set(deployableFlows.map((f) => f.name)));
    }
  };

  const canDeploy = () => {
    if (!selectedTemplate || selectedFlows.size === 0) return false;
    for (const v of selectedTemplate.variables) {
      if (v.required && !variables[v.key]) return false;
    }
    return true;
  };

  const handleDeploy = async () => {
    if (!selectedTemplate) return;
    setIsDeploying(true);
    try {
      const results = await installTemplate({
        id: selectedTemplate.id,
        variables,
        flow_names: [...selectedFlows],
        override: overrideMode,
      });
      setDeployResults(results);

      const successCount = results.filter((r) => r.success).length;
      const failCount = results.filter((r) => !r.success && !r.skipped).length;

      if (failCount === 0) {
        addToast({
          id: "template-deploy-success",
          title: "Flows Deployed",
          description: `${successCount} flow${successCount !== 1 ? "s" : ""} created successfully.`,
          variant: "success",
        });
      } else {
        addToast({
          id: "template-deploy-partial",
          title: "Deployment Completed",
          description: `${successCount} succeeded, ${failCount} failed.`,
          variant: successCount === 0 ? "error" : "warning",
        });
      }
      await loadTemplates();
    } catch (error) {
      addToast({
        id: "template-deploy-error",
        title: "Deployment Failed",
        description:
          error instanceof Error ? error.message : "Unknown error occurred.",
        variant: "error",
      });
    } finally {
      setIsDeploying(false);
    }
  };

  if (deployResults) {
    const successCount = deployResults.filter((r) => r.success).length;
    const failCount = deployResults.filter(
      (r) => !r.success && !r.skipped,
    ).length;

    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh]">
        <div className="w-full max-w-xl px-6">
          <div className="mb-8 text-center">
            <div
              className={`flex items-center justify-center w-16 h-16 rounded-full mx-auto mb-4 ${failCount === 0 ? "bg-green-500/10 text-green-500" : "bg-yellow-500/10 text-yellow-500"}`}
            >
              {failCount === 0 ? (
                <Check className="h-8 w-8" />
              ) : (
                <CircleAlert className="h-8 w-8" />
              )}
            </div>
            <h1 className="text-2xl font-bold">
              {failCount === 0 ? "All Flows Deployed" : "Deployment Complete"}
            </h1>
            <p className="text-muted-foreground mt-1">
              {successCount} flow{successCount !== 1 ? "s" : ""} created
              {failCount > 0 ? `, ${failCount} failed` : ""}
            </p>
          </div>

          <div className="space-y-2 mb-8">
            {deployResults.map((result) => (
              <div
                key={result.name}
                className={`flex items-center justify-between p-3 rounded-lg border ${
                  result.success
                    ? "bg-green-500/5 border-green-500/20"
                    : result.skipped
                      ? "bg-muted/50 border-border"
                      : "bg-destructive/5 border-destructive/20"
                }`}
              >
                <span className="text-sm font-mono">{result.name}</span>
                {result.success ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : result.skipped ? (
                  <span className="text-xs text-muted-foreground">
                    Skipped (already deployed)
                  </span>
                ) : (
                  <span className="text-xs text-destructive">
                    {result.error}
                  </span>
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
                setSelectedTemplate(null);
                setSelectedFlows(new Set());
              }}
            >
              Deploy Another Template
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
        <p className="text-muted-foreground">Deploying flows...</p>
      </div>
    );
  }

  if (templatesLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        <p className="text-muted-foreground">Loading templates...</p>
      </div>
    );
  }

  if (!selectedTemplate) {
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
            <h1 className="text-2xl font-bold">Templates</h1>
            <p className="text-muted-foreground mt-1">
              Select a template to deploy pre-built flows for a service
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4">
            {templateList.map((tmpl) => (
              <button
                key={tmpl.id}
                onClick={() => handleSelectTemplate(tmpl)}
                className="group relative flex items-start gap-4 p-6 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
              >
                <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary shrink-0 transition-transform duration-300 group-hover:scale-110">
                  <Package className="h-6 w-6" />
                </div>
                <div className="flex-1 min-w-0">
                  <h2 className="text-lg font-semibold mb-1">{tmpl.name}</h2>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    {tmpl.description}
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
            setSelectedTemplate(null);
            setSelectedFlows(new Set());
          }}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Templates
        </button>

        <div className="mb-8">
          <h1 className="text-2xl font-bold">Deploy {selectedTemplate.name} Template</h1>
          <p className="text-muted-foreground mt-1">
            Configure shared settings and select which flows to create
          </p>
        </div>

        <div className="space-y-8">
          <div className="space-y-4">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Configuration
            </h2>
            {selectedTemplate.variables.map((variable) => (
              <TemplateVariableField
                key={variable.key}
                variable={variable}
                value={variables[variable.key] || ""}
                onChange={(val) =>
                  setVariables((prev) => ({ ...prev, [variable.key]: val }))
                }
              />
            ))}
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
                Flows ({selectedFlows.size} of {deployableFlows.length}{" "}
                available)
              </h2>
              <div className="flex items-center gap-3">
                {hasInstalledFlows && (
                  <label className="flex items-center gap-2 cursor-pointer">
                    <Checkbox
                      checked={overrideMode}
                      onCheckedChange={(checked) => {
                        setOverrideMode(!!checked);
                        if (!checked) {
                          setSelectedFlows((prev) => {
                            const next = new Set(prev);
                            for (const f of selectedTemplate!.flows) {
                              if (f.installed) next.delete(f.name);
                            }
                            return next;
                          });
                        }
                      }}
                    />
                    <span className="text-xs text-muted-foreground">
                      Override existing
                    </span>
                  </label>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={toggleAll}
                  className="h-7 text-xs"
                >
                  {selectedFlows.size === deployableFlows.length
                    ? "Deselect All"
                    : "Select All"}
                </Button>
              </div>
            </div>

            {overrideMode && (
              <div className="flex items-start gap-2 p-3 rounded-lg border border-yellow-500/30 bg-yellow-500/5 text-sm">
                <CircleAlert className="h-4 w-4 text-yellow-500 shrink-0 mt-0.5" />
                <span className="text-yellow-700 dark:text-yellow-400">
                  Override mode will replace the configuration of existing
                  flows. Any manual edits to those flows will be lost.
                </span>
              </div>
            )}

            {deployableFlows.length === 0 && selectedTemplate.flows.length > 0 && (
              <div className="text-sm text-muted-foreground py-6 text-center border border-dashed rounded-lg">
                All flows from this template are already deployed. Enable "Override
                existing" to redeploy them.
              </div>
            )}

            <div className="space-y-2">
              {selectedTemplate.flows.map((flow) => {
                const selectable = overrideMode || !flow.installed;
                return (
                  <label
                    key={flow.name}
                    className={`flex items-start gap-3 p-3 rounded-lg border transition-colors ${
                      !selectable
                        ? "border-border bg-muted/50 cursor-default opacity-60"
                        : selectedFlows.has(flow.name)
                          ? "border-primary/50 bg-primary/5 cursor-pointer"
                          : "border-border bg-card hover:border-border/80 cursor-pointer"
                    }`}
                  >
                    <Checkbox
                      checked={selectedFlows.has(flow.name)}
                      onCheckedChange={() =>
                        selectable && toggleFlow(flow.name)
                      }
                      disabled={!selectable}
                      className="mt-0.5"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-mono font-medium">
                          {flow.name}
                        </span>
                        <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-primary/10 text-primary">
                          {flow.kind === "tool" ? "MCP tool" : "Automation"}
                        </span>
                        {flow.installed && (
                          <span
                            className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${
                              overrideMode && selectedFlows.has(flow.name)
                                ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300"
                                : "bg-muted text-muted-foreground"
                            }`}
                          >
                            {overrideMode && selectedFlows.has(flow.name)
                              ? "Will override"
                              : "Already deployed"}
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground mt-0.5">
                        {flow.description}
                      </div>
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
            Deploy {selectedFlows.size} Flow
            {selectedFlows.size !== 1 ? "s" : ""}
            <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </div>
    </div>
  );
}
