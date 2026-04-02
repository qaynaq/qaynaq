import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Bar, BarChart, CartesianGrid, XAxis, YAxis, Cell, Pie, PieChart } from "recharts";
import {
  Activity,
  AlertTriangle,
  ArrowDownToLine,
  ArrowUpFromLine,
  Check,
  Copy,
  Loader2,
  Plus,
  Server,
  Shield,
  Workflow,
  X,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { fetchAnalytics, fetchSetupStatus } from "@/lib/api";
import type { Analytics } from "@/lib/entities";
import { WelcomeOverlay } from "@/components/first-run/welcome-overlay";
import { DatabaseWizard, type DatabaseWizardResult } from "@/components/first-run/database-wizard";
import { GoogleWorkspaceWizard, type GoogleWorkspaceResult } from "@/components/first-run/google-workspace-wizard";
import { ShopifyWizard, type ShopifyWizardResult } from "@/components/first-run/shopify-wizard";
import { SuccessScreen, type ToolInfo } from "@/components/first-run/success-screen";

const STATUS_COLORS: Record<string, string> = {
  active: "hsl(142, 76%, 36%)",
  completed: "hsl(221, 83%, 53%)",
  paused: "hsl(38, 92%, 50%)",
  failed: "hsl(0, 84%, 60%)",
};

const eventsChartConfig = {
  input_events: { label: "Input", color: "hsl(221, 83%, 53%)" },
  output_events: { label: "Output", color: "hsl(142, 76%, 36%)" },
  error_events: { label: "Errors", color: "hsl(0, 84%, 60%)" },
} satisfies ChartConfig;

const statusChartConfig = {
  active: { label: "Active", color: STATUS_COLORS.active },
  completed: { label: "Completed", color: STATUS_COLORS.completed },
  paused: { label: "Paused", color: STATUS_COLORS.paused },
  failed: { label: "Failed", color: STATUS_COLORS.failed },
} satisfies ChartConfig;

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toString();
}

type WizardState = "overlay" | "database" | "google-workspace" | "shopify" | "success" | null;

function CopySnippet({ content }: { content: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = async () => {
    await navigator.clipboard.writeText(content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="relative">
      <pre className="bg-muted rounded-md p-3 text-xs overflow-x-auto font-mono whitespace-pre-wrap break-all">
        {content}
      </pre>
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 p-1.5 rounded-md bg-background/80 hover:bg-background border text-muted-foreground hover:text-foreground transition-colors"
      >
        {copied ? (
          <Check className="h-3.5 w-3.5 text-green-500" />
        ) : (
          <Copy className="h-3.5 w-3.5" />
        )}
      </button>
    </div>
  );
}

function NextStepsCard({ onDismiss }: { onDismiss: () => void }) {
  const navigate = useNavigate();
  const port = window.location.port || "8080";
  const mcpUrl = `http://localhost:${port}/mcp`;

  return (
    <Card className="mb-6 animate-glow-border bg-primary/[0.02]">
      <CardContent className="pt-5 pb-4">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
            <p className="text-sm font-semibold">Get started</p>
          </div>
          <button
            onClick={onDismiss}
            className="text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="grid grid-cols-3 gap-4">
          <Dialog>
            <DialogTrigger asChild>
              <button
                className="flex items-center gap-3 p-3 rounded-lg border hover:border-primary/30 hover:bg-primary/5 text-left transition-colors"
              >
                <Copy className="h-5 w-5 text-muted-foreground flex-shrink-0" />
                <div>
                  <p className="text-sm font-medium">Connect your AI</p>
                  <p className="text-xs text-muted-foreground">Setup instructions</p>
                </div>
              </button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Connect your AI assistant</DialogTitle>
              </DialogHeader>
              <Tabs defaultValue="claude-code" className="mt-2">
                <TabsList className="w-full">
                  <TabsTrigger value="claude-code" className="flex-1">Claude Code</TabsTrigger>
                  <TabsTrigger value="claude-desktop" className="flex-1">Claude Desktop</TabsTrigger>
                  <TabsTrigger value="cursor" className="flex-1">Cursor</TabsTrigger>
                </TabsList>
                <TabsContent value="claude-code" className="mt-3">
                  <p className="text-xs text-muted-foreground mb-2">
                    Run in your terminal:
                  </p>
                  <CopySnippet content={`claude mcp add Qaynaq -- npx mcp-remote ${mcpUrl}`} />
                </TabsContent>
                <TabsContent value="claude-desktop" className="mt-3">
                  <p className="text-xs text-muted-foreground mb-2">
                    Add to your <code>claude_desktop_config.json</code>:
                  </p>
                  <CopySnippet content={JSON.stringify({ mcpServers: { Qaynaq: { command: "npx", args: ["-y", "mcp-remote", mcpUrl] } } }, null, 2)} />
                </TabsContent>
                <TabsContent value="cursor" className="mt-3">
                  <p className="text-xs text-muted-foreground mb-2">
                    Add to your Cursor MCP settings:
                  </p>
                  <CopySnippet content={JSON.stringify({ mcpServers: { Qaynaq: { url: mcpUrl } } }, null, 2)} />
                </TabsContent>
              </Tabs>
            </DialogContent>
          </Dialog>
          <button
            onClick={() => navigate("/flows/new")}
            className="flex items-center gap-3 p-3 rounded-lg border hover:border-primary/30 hover:bg-primary/5 text-left transition-colors"
          >
            <Plus className="h-5 w-5 text-muted-foreground flex-shrink-0" />
            <div>
              <p className="text-sm font-medium">Build more tools</p>
              <p className="text-xs text-muted-foreground">Create new flows</p>
            </div>
          </button>
          <button
            onClick={() => navigate("/settings")}
            className="flex items-center gap-3 p-3 rounded-lg border hover:border-primary/30 hover:bg-primary/5 text-left transition-colors"
          >
            <Shield className="h-5 w-5 text-muted-foreground flex-shrink-0" />
            <div>
              <p className="text-sm font-medium">Secure endpoint</p>
              <p className="text-xs text-muted-foreground">Enable MCP auth</p>
            </div>
          </button>
        </div>
      </CardContent>
    </Card>
  );
}

export default function Home() {
  const navigate = useNavigate();
  const [data, setData] = useState<Analytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [setupStatus, setSetupStatus] = useState<{ first_run_complete: boolean } | null>(null);
  const [wizardState, setWizardState] = useState<WizardState>(null);
  const [toolInfo, setToolInfo] = useState<ToolInfo | null>(null);
  const [nextStepsDismissed, setNextStepsDismissed] = useState(
    () => localStorage.getItem("next-steps-dismissed") === "true",
  );

  useEffect(() => {
    fetchSetupStatus()
      .then(setSetupStatus)
      .catch(() => setSetupStatus({ first_run_complete: true }));
  }, []);

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const analytics = await fetchAnalytics();
        setData(analytics);
      } catch (err) {
        setError("Failed to load analytics");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const handleSelectPath = (path: string) => {
    if (path === "skip") {
      setSetupStatus({ first_run_complete: true });
      setWizardState(null);
    } else if (path === "other") {
      navigate("/flows/new");
    } else {
      setWizardState(path as WizardState);
    }
  };

  const handleDatabaseComplete = (result: DatabaseWizardResult) => {
    setToolInfo({
      name: result.flowName,
      description: result.flowDescription,
      type: "database",
    });
    setWizardState("success");
  };

  const handleGoogleComplete = (result: GoogleWorkspaceResult) => {
    setToolInfo({
      name: result.packName,
      type: "google-workspace",
      toolCount: result.toolCount,
    });
    setWizardState("success");
  };

  const handleShopifyComplete = (result: ShopifyWizardResult) => {
    setToolInfo({
      name: "Shopify",
      type: "shopify",
      toolCount: result.toolCount,
    });
    setWizardState("success");
  };

  const handleChainToGoogleWorkspace = () => {
    setWizardState("google-workspace");
  };

  const handleBuildAnother = () => {
    setWizardState("overlay");
  };

  const handleDone = () => {
    setSetupStatus({ first_run_complete: true });
    setWizardState(null);
    // Reload analytics since new flows were created
    fetchAnalytics().then(setData).catch(() => {});
  };

  const handleDismissNextSteps = () => {
    setNextStepsDismissed(true);
    localStorage.setItem("next-steps-dismissed", "true");
  };

  // First-run overlay
  if (setupStatus && !setupStatus.first_run_complete) {
    if (!wizardState || wizardState === "overlay") {
      return <WelcomeOverlay onSelectPath={handleSelectPath} />;
    }
    if (wizardState === "database") {
      return (
        <DatabaseWizard
          onComplete={handleDatabaseComplete}
          onBack={() => setWizardState("overlay")}
        />
      );
    }
    if (wizardState === "google-workspace") {
      return (
        <GoogleWorkspaceWizard
          onComplete={handleGoogleComplete}
          onBack={() => setWizardState("overlay")}
        />
      );
    }
    if (wizardState === "shopify") {
      return (
        <ShopifyWizard
          onComplete={handleShopifyComplete}
          onBack={() => setWizardState("overlay")}
        />
      );
    }
    if (wizardState === "success" && toolInfo) {
      return (
        <SuccessScreen
          toolInfo={toolInfo}
          onBuildAnother={handleBuildAnother}
          onDone={handleDone}
          onChainToGoogleWorkspace={
            toolInfo.type === "shopify" ? handleChainToGoogleWorkspace : undefined
          }
        />
      );
    }
  }

  // Success screen can also show after setup is marked complete
  if (wizardState === "success" && toolInfo) {
    return (
      <SuccessScreen
        toolInfo={toolInfo}
        onBuildAnother={handleBuildAnother}
        onDone={handleDone}
        onChainToGoogleWorkspace={
          toolInfo.type === "shopify" ? handleChainToGoogleWorkspace : undefined
        }
      />
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[50vh]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold">Dashboard</h1>
        </div>
        <Card>
          <CardContent className="flex items-center justify-center py-10">
            <p className="text-muted-foreground">{error || "No data available"}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const statusData = data.flows_by_status.map((s) => ({
    name: s.status,
    value: s.count,
    fill: STATUS_COLORS[s.status] || "hsl(var(--muted))",
  }));

  const timeSeriesData = data.events_over_time.map((pt) => ({
    date: pt.timestamp,
    input_events: pt.input_events,
    output_events: pt.output_events,
    error_events: pt.error_events,
  }));

  const showNextSteps =
    setupStatus?.first_run_complete && data.total_flows <= 2 && !nextStepsDismissed;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Dashboard</h1>
      </div>

      {showNextSteps && <NextStepsCard onDismiss={handleDismissNextSteps} />}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <Card>
          <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
            <CardTitle className="text-sm font-medium">Total Flows</CardTitle>
            <Workflow className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data.total_flows}</div>
            <div className="flex gap-1.5 mt-2 flex-wrap">
              {data.flows_by_status.map((s) => (
                <Badge key={s.status} variant="secondary" className="text-xs">
                  {s.status}: {s.count}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
            <CardTitle className="text-sm font-medium">Input Events</CardTitle>
            <ArrowDownToLine className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(data.total_input_events)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Total events ingested across all flows
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
            <CardTitle className="text-sm font-medium">Output Events</CardTitle>
            <ArrowUpFromLine className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(data.total_output_events)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Total events delivered to outputs
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
            <CardTitle className="text-sm font-medium">Active Workers</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data.active_workers}</div>
            <div className="flex items-center gap-1.5 mt-1">
              {data.total_processor_errors > 0 && (
                <span className="flex items-center text-xs text-destructive">
                  <AlertTriangle className="h-3 w-3 mr-1" />
                  {formatNumber(data.total_processor_errors)} processor errors
                </span>
              )}
              {data.total_processor_errors === 0 && (
                <span className="flex items-center text-xs text-muted-foreground">
                  <Activity className="h-3 w-3 mr-1" />
                  No processor errors
                </span>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-7 mb-6">
        <Card className="md:col-span-5">
          <CardHeader>
            <CardTitle>Events Over Time</CardTitle>
            <CardDescription>Daily event counts for the last 30 days</CardDescription>
          </CardHeader>
          <CardContent>
            {timeSeriesData.length > 0 ? (
              <ChartContainer config={eventsChartConfig} className="h-[300px] w-full">
                <BarChart data={timeSeriesData} accessibilityLayer>
                  <CartesianGrid vertical={false} />
                  <XAxis
                    dataKey="date"
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={(value) => {
                      const d = new Date(value);
                      return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
                    }}
                  />
                  <YAxis tickLine={false} axisLine={false} tickFormatter={formatNumber} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Bar dataKey="input_events" fill="var(--color-input_events)" radius={[2, 2, 0, 0]} />
                  <Bar dataKey="output_events" fill="var(--color-output_events)" radius={[2, 2, 0, 0]} />
                  <Bar dataKey="error_events" fill="var(--color-error_events)" radius={[2, 2, 0, 0]} />
                </BarChart>
              </ChartContainer>
            ) : (
              <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                No event data available yet
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>Flow Status</CardTitle>
            <CardDescription>Distribution by status</CardDescription>
          </CardHeader>
          <CardContent>
            {statusData.length > 0 ? (
              <ChartContainer config={statusChartConfig} className="h-[300px] w-full">
                <PieChart accessibilityLayer>
                  <ChartTooltip content={<ChartTooltipContent nameKey="name" />} />
                  <Pie data={statusData} dataKey="value" nameKey="name" innerRadius={50} strokeWidth={2}>
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.fill} />
                    ))}
                  </Pie>
                </PieChart>
              </ChartContainer>
            ) : (
              <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                No flows yet
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Top Input Components</CardTitle>
            <CardDescription>Most used input connectors</CardDescription>
          </CardHeader>
          <CardContent>
            {data.top_input_components.length > 0 ? (
              <div className="space-y-3">
                {data.top_input_components.map((c) => (
                  <div key={c.component} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-2 rounded-full bg-blue-500" />
                      <span className="text-sm font-medium">{c.component}</span>
                    </div>
                    <span className="text-sm text-muted-foreground">
                      {c.count} {c.count === 1 ? "flow" : "flows"}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No flows configured yet</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top Output Components</CardTitle>
            <CardDescription>Most used output connectors</CardDescription>
          </CardHeader>
          <CardContent>
            {data.top_output_components.length > 0 ? (
              <div className="space-y-3">
                {data.top_output_components.map((c) => (
                  <div key={c.component} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-2 rounded-full bg-green-500" />
                      <span className="text-sm font-medium">{c.component}</span>
                    </div>
                    <span className="text-sm text-muted-foreground">
                      {c.count} {c.count === 1 ? "flow" : "flows"}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No flows configured yet</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
