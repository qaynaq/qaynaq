import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import {
  Calendar,
  HardDrive,
  Table2,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Check,
} from "lucide-react";
import { fetchConnections, createFlow } from "@/lib/api";
import { templatePacks } from "@/lib/mcp-tool-templates";
import { buildFlowFromTemplate } from "@/lib/flow-builder-utils";
import type { Connection } from "@/lib/entities";

const PACK_ICONS: Record<string, React.ReactNode> = {
  "google-calendar": <Calendar className="h-6 w-6" />,
  "google-drive": <HardDrive className="h-6 w-6" />,
  "google-sheets": <Table2 className="h-6 w-6" />,
};

const googlePacks = templatePacks.filter(
  (p) =>
    p.id === "google-calendar" ||
    p.id === "google-drive" ||
    p.id === "google-sheets",
);

export type GoogleWorkspaceResult = {
  packName: string;
  toolCount: number;
};

export function GoogleWorkspaceWizard({
  onComplete,
  onBack,
}: {
  onComplete: (result: GoogleWorkspaceResult) => void;
  onBack: () => void;
}) {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [selectedConnection, setSelectedConnection] = useState("");
  const [connectionsLoaded, setConnectionsLoaded] = useState(false);

  const [selectedPack, setSelectedPack] = useState<string | null>(null);
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set());
  const [deploying, setDeploying] = useState(false);
  const [deployProgress, setDeployProgress] = useState({ current: 0, total: 0 });

  useEffect(() => {
    fetchConnections()
      .then((conns) => {
        setConnections(conns);
        setConnectionsLoaded(true);
      })
      .catch(() => setConnectionsLoaded(true));
  }, []);

  const pack = googlePacks.find((p) => p.id === selectedPack);

  const handleSelectPack = (packId: string) => {
    setSelectedPack(packId);
    const p = googlePacks.find((p) => p.id === packId);
    if (p) {
      setSelectedTools(new Set(p.templates.map((t) => t.id)));
    }
  };

  const toggleTool = (toolId: string) => {
    setSelectedTools((prev) => {
      const next = new Set(prev);
      if (next.has(toolId)) {
        next.delete(toolId);
      } else {
        next.add(toolId);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (!pack) return;
    if (selectedTools.size === pack.templates.length) {
      setSelectedTools(new Set());
    } else {
      setSelectedTools(new Set(pack.templates.map((t) => t.id)));
    }
  };

  const handleDeploy = async () => {
    if (!pack) return;

    const tools = pack.templates.filter((t) => selectedTools.has(t.id));
    setDeploying(true);
    setDeployProgress({ current: 0, total: tools.length });

    const sharedConfig: Record<string, string> = {
      oauth_connection: selectedConnection,
    };

    for (let i = 0; i < tools.length; i++) {
      setDeployProgress({ current: i + 1, total: tools.length });
      const flow = buildFlowFromTemplate(
        tools[i],
        sharedConfig,
        pack.sharedConfig,
        pack.id,
      );
      try {
        await createFlow(flow);
      } catch {
        // continue deploying remaining tools
      }
    }

    setDeploying(false);
    onComplete({ packName: pack.name, toolCount: tools.length });
  };

  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
      <div className="max-w-[620px] w-full bg-card border rounded-xl p-10 mx-4">
        <div className="flex items-center gap-2 mb-6">
          <div className="flex gap-1">
            {[1, 2].map((s) => (
              <div
                key={s}
                className={`h-1.5 w-12 rounded-full ${s <= step ? "bg-primary" : "bg-muted"}`}
              />
            ))}
          </div>
          <span className="text-xs text-muted-foreground ml-2">
            Step {step} of 2
          </span>
        </div>

        {step === 1 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">
                Connect Google Workspace
              </h2>
              <p className="text-sm text-muted-foreground">
                Select an OAuth connection for Google APIs
              </p>
            </div>

            {connectionsLoaded && connections.length > 0 && (
              <div>
                <Label className="text-sm mb-1.5 block">
                  OAuth Connection
                </Label>
                <Select
                  value={selectedConnection}
                  onValueChange={setSelectedConnection}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a connection" />
                  </SelectTrigger>
                  <SelectContent>
                    {connections.map((c) => (
                      <SelectItem key={c.name} value={c.name}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {selectedConnection && (
                  <div className="flex items-center gap-2 mt-2">
                    <div className="h-2 w-2 rounded-full bg-green-500" />
                    <span className="text-sm text-green-600">Connected</span>
                  </div>
                )}
              </div>
            )}

            {connectionsLoaded && connections.length === 0 && (
              <div className="text-center py-6">
                <p className="text-sm text-muted-foreground mb-3">
                  No OAuth connections found.
                </p>
                <Button
                  variant="outline"
                  onClick={() => navigate("/connections")}
                >
                  Create a connection
                </Button>
              </div>
            )}

            {!connectionsLoaded && (
              <div className="flex justify-center py-6">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            )}

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={onBack}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <Button
                onClick={() => setStep(2)}
                disabled={!selectedConnection}
              >
                Next
                <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">
                Select tools to deploy
              </h2>
              <p className="text-sm text-muted-foreground">
                Pick a pack and choose which tools to enable
              </p>
            </div>

            {!selectedPack && (
              <div className="grid grid-cols-3 gap-3">
                {googlePacks.map((p) => (
                  <Card
                    key={p.id}
                    className="cursor-pointer hover:border-primary transition-colors duration-200"
                    onClick={() => handleSelectPack(p.id)}
                  >
                    <CardContent className="pt-5 pb-4 text-center">
                      <div className="mx-auto mb-2 text-muted-foreground">
                        {PACK_ICONS[p.id]}
                      </div>
                      <p className="font-medium text-sm mb-1">{p.name}</p>
                      <Badge variant="secondary" className="text-xs">
                        {p.templates.length} tools
                      </Badge>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}

            {selectedPack && pack && (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedPack(null)}
                  >
                    <ChevronLeft className="h-4 w-4 mr-1" />
                    {pack.name}
                  </Button>
                  <button
                    onClick={toggleAll}
                    className="text-xs text-muted-foreground hover:text-foreground"
                  >
                    {selectedTools.size === pack.templates.length
                      ? "Deselect all"
                      : "Select all"}
                  </button>
                </div>

                <div className="max-h-[280px] overflow-y-auto space-y-1">
                  {pack.templates.map((t) => (
                    <label
                      key={t.id}
                      className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-muted/50 cursor-pointer"
                    >
                      <Checkbox
                        checked={selectedTools.has(t.id)}
                        onCheckedChange={() => toggleTool(t.id)}
                      />
                      <div className="min-w-0">
                        <p className="text-sm font-medium truncate">
                          {t.name}
                        </p>
                        <p className="text-xs text-muted-foreground truncate">
                          {t.description}
                        </p>
                      </div>
                    </label>
                  ))}
                </div>
              </div>
            )}

            {deploying && (
              <div className="flex items-center gap-3">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span className="text-sm">
                  Deploying {deployProgress.current}/{deployProgress.total}...
                </span>
              </div>
            )}

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={() => setStep(1)}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <Button
                onClick={handleDeploy}
                disabled={deploying || !selectedPack || selectedTools.size === 0}
              >
                {deploying ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Deploying...
                  </>
                ) : (
                  <>
                    Deploy {selectedTools.size} tool
                    {selectedTools.size !== 1 ? "s" : ""}
                  </>
                )}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
