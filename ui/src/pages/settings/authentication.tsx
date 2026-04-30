import { useEffect, useState } from "react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ShieldAlert, Info } from "lucide-react";
import { useToast } from "@/components/toast";
import { MCPSettings } from "@/lib/entities";
import { fetchMCPSettings, updateMCPProtected } from "@/lib/api";

export default function AuthenticationSettings() {
  const { addToast } = useToast();
  const [settings, setSettings] = useState<MCPSettings | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const data = await fetchMCPSettings();
      setSettings(data);
    } catch {
      addToast({
        id: "settings-load-error",
        title: "Error",
        description: "Failed to load MCP settings",
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleToggleProtection = async (checked: boolean) => {
    if (!settings) return;
    try {
      await updateMCPProtected(checked);
      setSettings({ ...settings, protected: checked });
      addToast({
        id: "mcp-protection-updated",
        title: "MCP Protection Updated",
        description: checked
          ? "MCP endpoint is now protected. Requests require a valid API token."
          : "MCP endpoint is now open. No authentication required.",
        variant: "success",
      });
    } catch {
      addToast({
        id: "mcp-protection-error",
        title: "Error",
        description: "Failed to update MCP protection setting",
        variant: "error",
      });
    }
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading...</p>;
  }
  if (!settings) return null;

  const authDisabled = !settings.auth_enabled;

  return (
    <Card>
      <CardHeader>
        <CardTitle>MCP Authentication</CardTitle>
        <CardDescription>
          {settings.oauth_enabled
            ? "Require MCP clients to authenticate via API token or OAuth"
            : "Require MCP clients to authenticate via API token"}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {authDisabled && (
          <Alert>
            <ShieldAlert className="h-4 w-4" />
            <AlertTitle>App authentication is disabled</AlertTitle>
            <AlertDescription>
              To use MCP protection, you need to enable application
              authentication first (basic or OAuth2). Without app auth, anyone
              can access this settings page and manage tokens.
            </AlertDescription>
          </Alert>
        )}

        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="mcp-protection" className="text-base">
              Require authentication
            </Label>
            <p className="text-sm text-muted-foreground">
              {settings.oauth_enabled
                ? "When enabled, MCP clients must authenticate with either an API token or via the OAuth flow"
                : "When enabled, MCP clients must provide a valid API token via the Authorization header"}
            </p>
          </div>
          <Switch
            id="mcp-protection"
            checked={settings.protected}
            onCheckedChange={handleToggleProtection}
            disabled={authDisabled}
          />
        </div>

        {settings.protected && (
          <Alert>
            <Info className="h-4 w-4" />
            <AlertTitle>How to authenticate</AlertTitle>
            <AlertDescription className="space-y-2">
              <p>
                Via header:{" "}
                <code className="text-xs bg-muted px-1 py-0.5 rounded">
                  Authorization: Bearer &lt;token&gt;
                </code>
              </p>
              <p>
                Via query parameter:{" "}
                <code className="text-xs bg-muted px-1 py-0.5 rounded">
                  /mcp?token=&lt;token&gt;
                </code>
              </p>
              {settings.oauth_enabled && (
                <p>
                  Or via OAuth: MCP clients (Claude Desktop, Cursor, ...)
                  discover the flow automatically. See{" "}
                  <a
                    href="/docs/guides/mcp-oauth"
                    className="underline"
                    target="_blank"
                    rel="noreferrer"
                  >
                    MCP OAuth
                  </a>
                  .
                </p>
              )}
            </AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );
}
