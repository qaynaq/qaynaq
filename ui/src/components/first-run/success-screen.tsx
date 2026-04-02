import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Check, ChevronRight, Copy } from "lucide-react";
import { completeSetup } from "@/lib/api";

export type ToolInfo = {
  name: string;
  description?: string;
  type: "database" | "google-workspace" | "shopify";
  toolCount?: number;
};

function CopyBlock({ content }: { content: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative">
      <pre className="bg-muted rounded-md p-4 text-xs overflow-x-auto font-mono whitespace-pre-wrap break-all">
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

function getJsonConfig(port: number) {
  return JSON.stringify(
    {
      mcpServers: {
        Qaynaq: {
          command: "npx",
          args: [
            "-y",
            "mcp-remote",
            `http://localhost:${port}/mcp`,
          ],
        },
      },
    },
    null,
    2,
  );
}

function getCursorConfig(port: number) {
  return JSON.stringify(
    {
      mcpServers: {
        Qaynaq: {
          url: `http://localhost:${port}/mcp`,
        },
      },
    },
    null,
    2,
  );
}

function getCliCommand(port: number) {
  return `claude mcp add Qaynaq -- npx mcp-remote http://localhost:${port}/mcp`;
}

export function SuccessScreen({
  toolInfo,
  onBuildAnother,
  onDone,
  onChainToGoogleWorkspace,
}: {
  toolInfo: ToolInfo;
  onBuildAnother: () => void;
  onDone: () => void;
  onChainToGoogleWorkspace?: () => void;
}) {
  const port = window.location.port || "8080";

  const handleDone = async () => {
    await completeSetup();
    onDone();
  };

  const suggestion =
    toolInfo.type === "database"
      ? `Ask: "Use the ${toolInfo.name} tool to get data"`
      : toolInfo.type === "shopify"
        ? `Try: "List my recent Shopify orders"`
        : `Try using your ${toolInfo.toolCount} new Google Workspace tools`;

  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
      <div className="max-w-[620px] w-full bg-card border rounded-xl p-10 mx-4">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-green-100 dark:bg-green-900/30 mb-4">
            <Check className="h-6 w-6 text-green-600" />
          </div>
          <h1 className="text-2xl font-semibold mb-2">
            Your MCP tool is live
          </h1>
          <p className="text-sm text-muted-foreground">
            {toolInfo.type === "database"
              ? `${toolInfo.name} - ${toolInfo.description || "Ready to query"}`
              : `${toolInfo.toolCount} ${toolInfo.name} tools deployed`}
          </p>
        </div>

        <div className="mb-6">
          <Tabs defaultValue="claude-code">
            <TabsList className="w-full">
              <TabsTrigger value="claude-code" className="flex-1">
                Claude Code
              </TabsTrigger>
              <TabsTrigger value="claude-desktop" className="flex-1">
                Claude Desktop
              </TabsTrigger>
              <TabsTrigger value="cursor" className="flex-1">
                Cursor
              </TabsTrigger>
            </TabsList>
            <TabsContent value="claude-code" className="mt-3">
              <p className="text-xs text-muted-foreground mb-2">
                Run in your terminal:
              </p>
              <CopyBlock content={getCliCommand(Number(port))} />
            </TabsContent>
            <TabsContent value="claude-desktop" className="mt-3">
              <p className="text-xs text-muted-foreground mb-2">
                Add to your <code>claude_desktop_config.json</code>:
              </p>
              <CopyBlock content={getJsonConfig(Number(port))} />
            </TabsContent>
            <TabsContent value="cursor" className="mt-3">
              <p className="text-xs text-muted-foreground mb-2">
                Add to your Cursor MCP settings:
              </p>
              <CopyBlock content={getCursorConfig(Number(port))} />
            </TabsContent>
          </Tabs>
        </div>

        <div className="bg-muted/50 rounded-lg p-4 mb-6">
          <p className="text-sm font-medium mb-1">Try it now</p>
          <p className="text-sm text-muted-foreground">{suggestion}</p>
        </div>

        {toolInfo.type === "shopify" && onChainToGoogleWorkspace && (
          <div className="border border-primary/20 bg-primary/5 rounded-lg p-4 mb-6">
            <p className="text-sm font-medium mb-1">Make it more powerful</p>
            <p className="text-sm text-muted-foreground mb-3">
              Add Google Sheets tools so your AI can export Shopify data to spreadsheets
            </p>
            <Button
              variant="outline"
              size="sm"
              onClick={onChainToGoogleWorkspace}
            >
              Add Google Sheets
              <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        )}

        <div className="flex gap-3">
          <Button variant="outline" className="flex-1" onClick={onBuildAnother}>
            Build another tool
          </Button>
          <Button className="flex-1" onClick={handleDone}>
            Go to Dashboard
          </Button>
        </div>
      </div>
    </div>
  );
}
