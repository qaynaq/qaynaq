import { useState, useCallback, useEffect, useRef } from "react";
import { createPortal } from "react-dom";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Check,
  ChevronLeft,
  ChevronRight,
  Loader2,
  AlertTriangle,
  RefreshCw,
  X,
} from "lucide-react";
import { testConnection, createFlow, fetchWorkers } from "@/lib/api";
import * as yaml from "js-yaml";
import type { Worker } from "@/lib/entities";

// ---------------------------------------------------------------------------
// Guided tour - non-blocking, state-driven
// ---------------------------------------------------------------------------

type TourStop = {
  id: string;
  target: string;
  text: string;
  side?: "top" | "bottom" | "left" | "right";
};

// Only action buttons get floating tooltips - form fields use inline hints instead
const STOPS: TourStop[] = [
  // Step 1
  { id: "test-connection", target: "test-connection", text: "Click here to verify Qaynaq can reach your database.", side: "right" },
  { id: "step1-next", target: "step1-next", text: "Connection verified - continue to the next step.", side: "left" },
  // Step 2
  { id: "step2-next", target: "step2-next", text: "Looking good - continue to review and deploy.", side: "left" },
  // Step 3
  { id: "deploy-btn", target: "deploy-btn", text: "Hit deploy to make this tool available to AI assistants.", side: "left" },
];

type WizardState = {
  connectionTested: boolean;
  step: number;
  toolName: string;
  sqlQuery: string;
  workersReady: boolean;
};

function resolveCurrentStop(state: WizardState): string | null {
  if (state.step === 1) {
    if (!state.connectionTested) return "test-connection";
    return "step1-next";
  }
  if (state.step === 2) {
    if (state.toolName && state.sqlQuery) return "step2-next";
    return null;
  }
  if (state.step === 3) {
    if (state.workersReady) return "deploy-btn";
    return null;
  }
  return null;
}

function FloatingTooltip({ stop, onDismiss }: {
  stop: TourStop;
  onDismiss: () => void;
}) {
  const [pos, setPos] = useState<{ top: number; left: number; arrow: string } | null>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const compute = () => {
      const el = document.querySelector(`[data-tour="${stop.target}"]`);
      if (!el || !tooltipRef.current) return;

      const r = el.getBoundingClientRect();
      const tt = tooltipRef.current.getBoundingClientRect();
      const gap = 10;
      const side = stop.side || "right";

      let top = 0;
      let left = 0;
      let arrow = "";

      if (side === "right" && r.right + gap + tt.width < window.innerWidth) {
        top = r.top + r.height / 2 - tt.height / 2;
        left = r.right + gap;
        arrow = "left";
      } else if (side === "left" && r.left - gap - tt.width > 0) {
        top = r.top + r.height / 2 - tt.height / 2;
        left = r.left - gap - tt.width;
        arrow = "right";
      } else if (side === "bottom" && r.bottom + gap + tt.height < window.innerHeight) {
        top = r.bottom + gap;
        left = r.left + r.width / 2 - tt.width / 2;
        arrow = "top";
      } else {
        top = r.top - gap - tt.height;
        left = r.left + r.width / 2 - tt.width / 2;
        arrow = "bottom";
      }

      top = Math.max(8, Math.min(top, window.innerHeight - tt.height - 8));
      left = Math.max(8, Math.min(left, window.innerWidth - tt.width - 8));

      setPos({ top, left, arrow });
    };

    // two frames: one for DOM, one for layout
    let raf1: number, raf2: number;
    raf1 = requestAnimationFrame(() => {
      raf2 = requestAnimationFrame(compute);
    });
    window.addEventListener("resize", compute);
    window.addEventListener("scroll", compute, true);

    // recompute when the wizard modal content changes (e.g. sample data preview toggled)
    const el = document.querySelector(`[data-tour="${stop.target}"]`);
    const modal = el?.closest("[data-wizard-modal]");
    let observer: MutationObserver | undefined;
    if (modal) {
      observer = new MutationObserver(() => {
        requestAnimationFrame(compute);
      });
      observer.observe(modal, { childList: true, subtree: true, attributes: true });
    }

    return () => {
      cancelAnimationFrame(raf1);
      cancelAnimationFrame(raf2);
      window.removeEventListener("resize", compute);
      window.removeEventListener("scroll", compute, true);
      observer?.disconnect();
    };
  }, [stop]);

  // add a subtle ring to the target element
  useEffect(() => {
    const el = document.querySelector(`[data-tour="${stop.target}"]`) as HTMLElement | null;
    if (!el) return;
    el.style.outline = "2px solid hsl(var(--primary) / 0.4)";
    el.style.outlineOffset = "3px";
    el.style.borderRadius = "6px";
    return () => {
      el.style.outline = "";
      el.style.outlineOffset = "";
      el.style.borderRadius = "";
    };
  }, [stop.target]);

  const arrowBorder: Record<string, string> = {
    left: "border-r-0 border-t-0",
    right: "border-l-0 border-b-0",
    top: "border-b-0 border-l-0",
    bottom: "border-t-0 border-r-0",
  };

  const arrowPos: Record<string, string> = {
    left: "left-0 top-1/2 -translate-x-1/2 -translate-y-1/2 rotate-45",
    right: "right-0 top-1/2 translate-x-1/2 -translate-y-1/2 rotate-45",
    top: "top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 rotate-45",
    bottom: "bottom-0 left-1/2 -translate-x-1/2 translate-y-1/2 rotate-45",
  };

  return createPortal(
    <div
      ref={tooltipRef}
      key={stop.id}
      className="fixed z-[70] max-w-[240px] pointer-events-auto animate-in fade-in slide-in-from-bottom-1 duration-200"
      style={{ top: pos?.top ?? -9999, left: pos?.left ?? -9999 }}
    >
      <div className="relative bg-popover border rounded-lg shadow-lg px-3 py-2.5">
        {pos?.arrow && (
          <div className={`absolute w-2.5 h-2.5 bg-popover ${arrowBorder[pos.arrow]} border ${arrowPos[pos.arrow]}`} />
        )}
        <button
          onClick={onDismiss}
          className="absolute top-1.5 right-1.5 text-muted-foreground hover:text-foreground"
        >
          <X className="h-3 w-3" />
        </button>
        <p className="text-xs leading-relaxed pr-3">{stop.text}</p>
      </div>
    </div>,
    document.body,
  );
}

function GuidedTour({ wizardState }: { wizardState: WizardState }) {
  const [dismissed, setDismissed] = useState(false);

  // reset when wizard step changes
  const stepRef = useRef(wizardState.step);
  useEffect(() => {
    if (wizardState.step !== stepRef.current) {
      stepRef.current = wizardState.step;
      setDismissed(false);
    }
  }, [wizardState.step]);

  if (dismissed) return null;

  const stopId = resolveCurrentStop(wizardState);
  if (!stopId) return null;

  const stop = STOPS.find((s) => s.id === stopId);
  if (!stop) return null;

  return <FloatingTooltip stop={stop} onDismiss={() => setDismissed(true)} />;
}

// ---------------------------------------------------------------------------
// Wizard
// ---------------------------------------------------------------------------

type ParameterDef = {
  name: string;
  type: string;
  description: string;
};

export type DatabaseWizardResult = {
  flowName: string;
  flowDescription: string;
};

const SAMPLE_DB_PATH = "/tmp/qaynaq/sample.db";
const SAMPLE_TOOL_NAME = "get_monthly_revenue";
const SAMPLE_TOOL_DESCRIPTION = "Returns monthly revenue for a given year";
const SAMPLE_QUERY = `SELECT
  strftime('%m', created_at) as month,
  SUM(amount) as revenue
FROM orders
WHERE strftime('%Y', created_at) = \${year}
GROUP BY 1`;
const SAMPLE_PARAMS: ParameterDef[] = [
  { name: "year", type: "string", description: "Year to get revenue for (e.g. 2025)" },
];

export function DatabaseWizard({
  onComplete,
  onBack,
}: {
  onComplete: (result: DatabaseWizardResult) => void;
  onBack: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const [step, setStep] = useState(1);
  const [useSampleData, setUseSampleData] = useState(false);
  const [driver, setDriver] = useState("postgres");
  const [connectionString, setConnectionString] = useState("");
  const [connectionTested, setConnectionTested] = useState(false);
  const [testingConnection, setTestingConnection] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);

  const [toolName, setToolName] = useState("");
  const [toolDescription, setToolDescription] = useState("");
  const [sqlQuery, setSqlQuery] = useState("");
  const [parameters, setParameters] = useState<ParameterDef[]>([]);

  const [workers, setWorkers] = useState<Worker[]>([]);
  const [workersLoaded, setWorkersLoaded] = useState(false);
  const [checkingWorkers, setCheckingWorkers] = useState(false);
  const [deploying, setDeploying] = useState(false);
  const [deployError, setDeployError] = useState<string | null>(null);

  const handleSampleToggle = (checked: boolean) => {
    setUseSampleData(checked);
    if (checked) {
      setDriver("sqlite");
      setConnectionString(SAMPLE_DB_PATH);
      setConnectionTested(false);
      setConnectionError(null);
      setToolName(SAMPLE_TOOL_NAME);
      setToolDescription(SAMPLE_TOOL_DESCRIPTION);
      setSqlQuery(SAMPLE_QUERY);
      setParameters(SAMPLE_PARAMS);
    } else {
      setDriver("postgres");
      setConnectionString("");
      setConnectionTested(false);
      setConnectionError(null);
      setToolName("");
      setToolDescription("");
      setSqlQuery("");
      setParameters([]);
    }
  };

  const handleTestConnection = async () => {
    setTestingConnection(true);
    setConnectionError(null);
    try {
      const result = await testConnection(driver, connectionString);
      if (result.ok) {
        setConnectionTested(true);
      } else {
        setConnectionError(result.error || "Connection failed");
        setConnectionTested(false);
      }
    } catch {
      setConnectionError("Failed to test connection");
      setConnectionTested(false);
    } finally {
      setTestingConnection(false);
    }
  };

  const extractParams = useCallback((query: string) => {
    const regex = /\$\{(\w+)\}/g;
    const names = new Set<string>();
    let match;
    while ((match = regex.exec(query)) !== null) {
      names.add(match[1]);
    }
    setParameters((prev) => {
      const existing = new Map(prev.map((p) => [p.name, p]));
      return Array.from(names).map(
        (name) =>
          existing.get(name) || { name, type: "string", description: "" },
      );
    });
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => extractParams(sqlQuery), 300);
    return () => clearTimeout(timer);
  }, [sqlQuery, extractParams]);

  const handleQueryChange = (value: string) => {
    setSqlQuery(value);
  };

  const updateParameterDescription = (name: string, description: string) => {
    setParameters((prev) =>
      prev.map((p) => (p.name === name ? { ...p, description } : p)),
    );
  };

  useEffect(() => {
    if (step === 3 && !workersLoaded) {
      fetchWorkers()
        .then((w) => {
          setWorkers(w.filter((w) => w.status === "active"));
          setWorkersLoaded(true);
        })
        .catch(() => setWorkersLoaded(true));
    }
  }, [step, workersLoaded]);

  const recheckWorkers = async () => {
    setCheckingWorkers(true);
    try {
      const w = await fetchWorkers();
      setWorkers(w.filter((w) => w.status === "active"));
    } catch {
      // keep current state
    } finally {
      setCheckingWorkers(false);
    }
  };

  const handleDeploy = async () => {
    setDeploying(true);
    setDeployError(null);

    const paramOrder: string[] = [];
    const parameterizedQuery = sqlQuery.replace(
      /\$\{(\w+)\}/g,
      (_match, name) => {
        paramOrder.push(name);
        return "?";
      },
    );

    const argsMapping = `[${paramOrder.map((n) => `this.${n}`).join(", ")}]`;

    const inputSchema: Record<string, any> = {
      type: "object",
      required: parameters.map((p) => p.name),
      properties: Object.fromEntries(
        parameters.map((p) => [
          p.name,
          { type: p.type, description: p.description },
        ]),
      ),
    };

    const flow = {
      name: toolName,
      status: "active",
      input_component: "mcp_tool",
      input_label: toolName,
      input_config: yaml.dump(
        {
          name: toolName,
          description: toolDescription,
          input_schema: inputSchema,
        },
        { lineWidth: -1, noRefs: true },
      ),
      output_component: "sync_response",
      output_label: "response",
      output_config: "",
      processors: [
        {
          label: "sql_raw",
          component: "sql_raw",
          config: yaml.dump(
            {
              driver: driver,
              dsn: connectionString,
              query: parameterizedQuery,
              args_mapping: argsMapping,
            },
            { lineWidth: -1, noRefs: true },
          ),
        },
        {
          label: "error_handler",
          component: "catch",
          config: yaml.dump(
            [
              {
                mapping:
                  'meta status_code = error().re_find_all("^\\\\[[0-9]+\\\\]").index(0).trim("[]").or("500")\nroot.error = error().re_replace_all("^\\\\[[0-9]+\\\\] ", "")',
              },
            ],
            { lineWidth: -1, noRefs: true },
          ),
        },
      ],
      is_ready: true,
      builder_state: "",
      managed_by: "",
    };

    try {
      await createFlow(flow);
      onComplete({ flowName: toolName, flowDescription: toolDescription });
    } catch {
      setDeployError("Failed to deploy. Try again?");
    } finally {
      setDeploying(false);
    }
  };

  const tourState: WizardState = {
    connectionTested,
    step,
    toolName,
    sqlQuery,
    workersReady: workersLoaded && workers.length > 0,
  };

  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
      <GuidedTour wizardState={tourState} />
      <div data-wizard-modal className="max-w-[620px] w-full bg-card border rounded-xl p-10 mx-4">
        <div className="flex items-center gap-2 mb-6">
          <div className="flex gap-1">
            {[1, 2, 3].map((s) => (
              <div
                key={s}
                className={`h-1.5 w-8 rounded-full ${s <= step ? "bg-primary" : "bg-muted"}`}
              />
            ))}
          </div>
          <span className="text-xs text-muted-foreground ml-2">
            Step {step} of 3
          </span>
        </div>

        {step === 1 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">
                Connect your database
              </h2>
              <p className="text-sm text-muted-foreground">
                We'll create an MCP tool that queries your database
              </p>
            </div>

            <div>
              <div className="flex items-center gap-3" data-tour="sample-toggle">
                <Switch
                  checked={useSampleData}
                  onCheckedChange={handleSampleToggle}
                />
                <Label className="text-sm">Try with sample data</Label>
              </div>
              <p className="text-xs text-muted-foreground mt-1.5 ml-11">Use a built-in demo database so you can try things without any setup.</p>
            </div>

            {useSampleData && (
              <div className="rounded-lg border bg-muted/30 p-4 space-y-3">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Sample database preview</p>
                <div className="grid grid-cols-3 gap-3 text-xs">
                  <div className="space-y-1">
                    <p className="font-medium">customers</p>
                    <p className="text-muted-foreground">20 rows</p>
                    <p className="text-muted-foreground font-mono">id, name, email, created_at</p>
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium">products</p>
                    <p className="text-muted-foreground">15 rows</p>
                    <p className="text-muted-foreground font-mono">id, name, price, category, stock</p>
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium">orders</p>
                    <p className="text-muted-foreground">100 rows</p>
                    <p className="text-muted-foreground font-mono">id, customer_id, amount, status, created_at</p>
                  </div>
                </div>
              </div>
            )}

            <div className="space-y-4">
              <div>
                <Label className="text-sm mb-1.5 block">Driver</Label>
                <Select
                  value={driver}
                  onValueChange={(v) => {
                    setDriver(v);
                    setConnectionTested(false);
                  }}
                  disabled={useSampleData}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="postgres">PostgreSQL</SelectItem>
                    <SelectItem value="mysql">MySQL</SelectItem>
                    <SelectItem value="sqlite">SQLite</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label className="text-sm mb-1.5 block">
                  Connection string
                </Label>
                <Input
                  value={connectionString}
                  onChange={(e) => {
                    setConnectionString(e.target.value);
                    setConnectionTested(false);
                  }}
                  placeholder={
                    driver === "postgres"
                      ? "postgres://user:pass@host:5432/db"
                      : driver === "mysql"
                        ? "user:pass@tcp(host:3306)/db"
                        : "/path/to/database.db"
                  }
                  readOnly={useSampleData}
                />
              </div>

              <div className="flex items-center gap-3">
                <div data-tour="test-connection">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleTestConnection}
                    disabled={!connectionString || testingConnection}
                  >
                    {testingConnection && (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    )}
                    {connectionTested && (
                      <Check className="h-4 w-4 mr-2 text-green-500" />
                    )}
                    Test Connection
                  </Button>
                </div>
                {connectionTested && (
                  <span className="text-sm text-green-600">Connected</span>
                )}
              </div>
              {connectionError && (
                <p className="text-sm text-destructive">{connectionError}</p>
              )}
            </div>

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={onBack}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <div data-tour="step1-next">
                <Button
                  onClick={() => setStep(2)}
                  disabled={!connectionTested}
                >
                  Next
                  <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">
                Configure your tool
              </h2>
              <p className="text-sm text-muted-foreground">
                Define what your MCP tool does
              </p>
            </div>

            <div className="space-y-4">
              <div>
                <Label className="text-sm mb-1.5 block">Tool name</Label>
                <Input
                  value={toolName}
                  onChange={(e) => {
                    const v = e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, "_");
                    setToolName(v);
                  }}
                  placeholder="get_monthly_revenue"
                />
                <p className="text-xs text-muted-foreground mt-1">This is how AI assistants will find and call your tool. Lowercase letters, numbers, and underscores only.</p>
              </div>

              <div>
                <Label className="text-sm mb-1.5 block">Description</Label>
                <Input
                  value={toolDescription}
                  onChange={(e) => setToolDescription(e.target.value)}
                  placeholder="Returns monthly revenue for a given year"
                />
                <p className="text-xs text-muted-foreground mt-1">Helps AI decide when to pick this tool over others. Be specific about what data it returns.</p>
              </div>

              <div>
                <Label className="text-sm mb-1.5 block">SQL Query</Label>
                <div className="rounded-md border overflow-hidden">
                  <Editor
                    height="140px"
                    language="sql"
                    theme={resolvedTheme === "dark" ? "vs-dark" : "light"}
                    value={sqlQuery}
                    onChange={(v) => handleQueryChange(v ?? "")}
                    options={{
                      minimap: { enabled: false },
                      lineNumbers: "off",
                      glyphMargin: false,
                      folding: false,
                      scrollBeyondLastLine: false,
                      renderLineHighlight: "none",
                      overviewRulerLanes: 0,
                      hideCursorInOverviewRuler: true,
                      overviewRulerBorder: false,
                      scrollbar: { vertical: "hidden", horizontal: "auto" },
                      fontSize: 13,
                      padding: { top: 10, bottom: 10 },
                      readOnly: useSampleData,
                      domReadOnly: useSampleData,
                    }}
                  />
                </div>
                <p className="text-xs text-muted-foreground mt-1.5">
                  {useSampleData
                    ? <>The <code className="text-[11px]">{"${param}"}</code> notation is used here for readability. Under the hood, parameters are mapped to safe <code className="text-[11px]">?</code> / <code className="text-[11px]">$1</code> placeholders via <code className="text-[11px]">args_mapping</code>.</>
                    : <>Use <code className="text-[11px]">{"${param}"}</code> to define parameters. At deploy time, each one becomes a <code className="text-[11px]">?</code> placeholder with a corresponding <code className="text-[11px]">args_mapping</code> entry.</>
                  }
                </p>
              </div>

              {parameters.length > 0 && (
                <div>
                  <Label className="text-sm mb-2 block">Parameters</Label>
                  <div className="space-y-2">
                    {parameters.map((p) => (
                      <div
                        key={p.name}
                        className="flex items-center gap-2 text-sm"
                      >
                        <code className="bg-muted px-2 py-1 rounded text-xs min-w-[80px]">
                          {p.name}
                        </code>
                        <span className="text-xs text-muted-foreground w-[60px]">
                          string
                        </span>
                        <Input
                          value={p.description}
                          onChange={(e) =>
                            updateParameterDescription(
                              p.name,
                              e.target.value,
                            )
                          }
                          placeholder="Description"
                          className="flex-1"
                          readOnly={useSampleData}
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={() => setStep(1)}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <div data-tour="step2-next">
                <Button
                  onClick={() => setStep(3)}
                  disabled={!toolName || !sqlQuery}
                >
                  Next
                  <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </div>
          </div>
        )}

        {step === 3 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">Deploy your tool</h2>
              <p className="text-sm text-muted-foreground">
                Review and deploy your MCP tool
              </p>
            </div>

            <div>
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">{toolName}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  {toolDescription && (
                    <p className="text-muted-foreground">{toolDescription}</p>
                  )}
                  <p>
                    {parameters.length} parameter
                    {parameters.length !== 1 ? "s" : ""}
                  </p>
                  <p className="text-muted-foreground">
                    {driver} - {useSampleData ? "sample data" : "custom database"}
                  </p>
                </CardContent>
              </Card>
            </div>

            {workersLoaded && workers.length === 0 && (
              <Alert>
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription className="space-y-2">
                  <p>No active worker found. Open a new terminal and start a worker - keep the coordinator running in this one.</p>
                  <code className="block text-xs bg-muted px-2 py-1 rounded">qaynaq --role worker --grpc-port 50001 --secret.key YOUR_KEY</code>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={recheckWorkers}
                    disabled={checkingWorkers}
                    className="mt-1"
                  >
                    {checkingWorkers ? (
                      <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                    ) : (
                      <RefreshCw className="h-3.5 w-3.5 mr-1.5" />
                    )}
                    Check again
                  </Button>
                </AlertDescription>
              </Alert>
            )}

            {workersLoaded && workers.length > 0 && (
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-green-500" />
                <span className="text-sm text-green-600">Ready to deploy</span>
              </div>
            )}

            {deployError && (
              <p className="text-sm text-destructive">{deployError}</p>
            )}

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={() => setStep(2)}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <div data-tour="deploy-btn">
                <Button
                  onClick={handleDeploy}
                  disabled={
                    deploying || (workersLoaded && workers.length === 0)
                  }
                >
                  {deploying && (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  )}
                  {deploying ? "Deploying..." : "Deploy"}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
