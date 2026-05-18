import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Check, Copy, LogOut } from "lucide-react";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { useAuth } from "@/lib/auth";

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

function getMcpUrl() {
  return `${window.location.origin}/mcp`;
}

function getClaudeCodeCommand() {
  return `claude mcp add Qaynaq -- npx mcp-remote ${getMcpUrl()}`;
}

function getClaudeDesktopConfig() {
  return JSON.stringify(
    {
      mcpServers: {
        Qaynaq: {
          command: "npx",
          args: ["-y", "mcp-remote", getMcpUrl()],
        },
      },
    },
    null,
    2,
  );
}

function getCursorConfig() {
  return JSON.stringify(
    {
      mcpServers: {
        Qaynaq: {
          url: getMcpUrl(),
        },
      },
    },
    null,
    2,
  );
}

export default function McpAccessPage() {
  const { email, logout } = useAuth();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 px-4 py-12">
      <div className="absolute top-4 right-4 flex items-center gap-2">
        <ThemeSwitcher />
        <Button variant="ghost" size="icon" onClick={logout} aria-label="Sign out">
          <LogOut className="h-5 w-5" />
        </Button>
      </div>
      <Card className="w-full max-w-2xl">
        <CardHeader className="space-y-3 text-center">
          <div className="flex flex-col items-center gap-2">
            <img src="/logo.png" alt="Qaynaq" className="h-16 w-auto" />
            <span className="text-xl font-bold">Qaynaq</span>
          </div>
          <CardTitle className="text-2xl">You have MCP-only access</CardTitle>
          <CardDescription>
            {email ? `Signed in as ${email}. ` : ""}
            Connect your AI client using the configuration below. The dashboard isn't available to your account.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              MCP server URL
            </label>
            <CopyBlock content={getMcpUrl()} />
          </div>

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
              <p className="text-xs text-muted-foreground mb-2">Run in your terminal:</p>
              <CopyBlock content={getClaudeCodeCommand()} />
            </TabsContent>
            <TabsContent value="claude-desktop" className="mt-3">
              <p className="text-xs text-muted-foreground mb-2">
                Add to your <code>claude_desktop_config.json</code>:
              </p>
              <CopyBlock content={getClaudeDesktopConfig()} />
            </TabsContent>
            <TabsContent value="cursor" className="mt-3">
              <p className="text-xs text-muted-foreground mb-2">
                Add to your Cursor MCP settings:
              </p>
              <CopyBlock content={getCursorConfig()} />
            </TabsContent>
          </Tabs>

          <p className="text-xs text-muted-foreground">
            Your AI client will open a browser window to complete an OAuth handshake with Qaynaq the first time you connect. After that, the client uses a refreshing bearer token on every MCP request.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
