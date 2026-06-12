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
} from "lucide-react";
import {
  fetchConnections,
  fetchTemplates,
  installTemplate,
} from "@/lib/api";
import type { Connection, Template } from "@/lib/entities";

const TEMPLATE_ICONS: Record<string, React.ReactNode> = {
  google_calendar: <Calendar className="h-6 w-6" />,
  google_drive: <HardDrive className="h-6 w-6" />,
  google_sheets: <Table2 className="h-6 w-6" />,
};

const GOOGLE_TEMPLATE_IDS = ["google_calendar", "google_drive", "google_sheets"];

export type GoogleWorkspaceResult = {
  templateName: string;
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
  const [googleTemplates, setGoogleTemplates] = useState<Template[]>([]);
  const [templatesLoaded, setTemplatesLoaded] = useState(false);

  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set());
  const [deploying, setDeploying] = useState(false);
  const [deployError, setDeployError] = useState("");

  useEffect(() => {
    fetchConnections()
      .then((conns) => {
        setConnections(conns);
        setConnectionsLoaded(true);
      })
      .catch(() => setConnectionsLoaded(true));
    fetchTemplates()
      .then((data) => {
        setGoogleTemplates(data.filter((p) => GOOGLE_TEMPLATE_IDS.includes(p.id)));
        setTemplatesLoaded(true);
      })
      .catch(() => setTemplatesLoaded(true));
  }, []);

  const tmpl = googleTemplates.find((p) => p.id === selectedTemplate);

  const handleSelectTemplate = (templateId: string) => {
    setSelectedTemplate(templateId);
    const p = googleTemplates.find((p) => p.id === templateId);
    if (p) {
      setSelectedTools(new Set(p.flows.map((f) => f.name)));
    }
  };

  const toggleTool = (name: string) => {
    setSelectedTools((prev) => {
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
    if (!tmpl) return;
    if (selectedTools.size === tmpl.flows.length) {
      setSelectedTools(new Set());
    } else {
      setSelectedTools(new Set(tmpl.flows.map((f) => f.name)));
    }
  };

  const handleDeploy = async () => {
    if (!tmpl) return;
    setDeploying(true);
    setDeployError("");

    try {
      const results = await installTemplate({
        id: tmpl.id,
        variables: { oauth_connection: selectedConnection },
        flow_names: [...selectedTools],
        override: true,
      });
      setDeploying(false);
      onComplete({
        templateName: tmpl.name,
        toolCount: results.filter((r) => r.success).length,
      });
    } catch (error) {
      setDeploying(false);
      setDeployError(
        error instanceof Error ? error.message : "Failed to deploy tools",
      );
    }
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
                Pick a template and choose which tools to enable
              </p>
            </div>

            {!templatesLoaded && (
              <div className="flex justify-center py-6">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            )}

            {templatesLoaded && !selectedTemplate && (
              <div className="grid grid-cols-3 gap-3">
                {googleTemplates.map((p) => (
                  <Card
                    key={p.id}
                    className="cursor-pointer hover:border-primary transition-colors duration-200"
                    onClick={() => handleSelectTemplate(p.id)}
                  >
                    <CardContent className="pt-5 pb-4 text-center">
                      <div className="mx-auto mb-2 text-muted-foreground">
                        {TEMPLATE_ICONS[p.id]}
                      </div>
                      <p className="font-medium text-sm mb-1">{p.name}</p>
                      <Badge variant="secondary" className="text-xs">
                        {p.flows.length} tools
                      </Badge>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}

            {selectedTemplate && tmpl && (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedTemplate(null)}
                  >
                    <ChevronLeft className="h-4 w-4 mr-1" />
                    {tmpl.name}
                  </Button>
                  <button
                    onClick={toggleAll}
                    className="text-xs text-muted-foreground hover:text-foreground"
                  >
                    {selectedTools.size === tmpl.flows.length
                      ? "Deselect all"
                      : "Select all"}
                  </button>
                </div>

                <div className="max-h-[280px] overflow-y-auto space-y-1">
                  {tmpl.flows.map((f) => (
                    <label
                      key={f.name}
                      className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-muted/50 cursor-pointer"
                    >
                      <Checkbox
                        checked={selectedTools.has(f.name)}
                        onCheckedChange={() => toggleTool(f.name)}
                      />
                      <div className="min-w-0">
                        <p className="text-sm font-medium truncate">
                          {f.name}
                        </p>
                        <p className="text-xs text-muted-foreground truncate">
                          {f.description}
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
                  Deploying {selectedTools.size} tool
                  {selectedTools.size !== 1 ? "s" : ""}...
                </span>
              </div>
            )}

            {deployError && (
              <p className="text-sm text-destructive">{deployError}</p>
            )}

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={() => setStep(1)}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <Button
                onClick={handleDeploy}
                disabled={deploying || !selectedTemplate || selectedTools.size === 0}
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
