import { useSearchParams } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ShieldAlert } from "lucide-react";

type ErrorContent = {
  title: string;
  description: string;
  body: React.ReactNode;
};

function content(code: string, message: string): ErrorContent {
  switch (code) {
    case "stale_client":
      return {
        title: "Stale MCP client credentials",
        description: `This server does not recognize the client_id "${message}".`,
        body: (
          <>
            <p>
              Your MCP client is using cached registration data that no longer
              exists on this server.
            </p>
            <p>
              Clear the cached credentials on your machine and reconnect the
              MCP client:
            </p>
            <pre className="bg-muted p-3 rounded text-xs overflow-x-auto">
              <code>rm -rf ~/.mcp-auth/mcp-remote-*</code>
            </pre>
            <p>
              Then reconnect from your MCP client (Claude Desktop, Cursor,
              etc.). The client will register itself again automatically.
            </p>
          </>
        ),
      };
    case "invalid_redirect_uri":
      return {
        title: "Invalid redirect URI",
        description: message,
        body: (
          <p>
            The MCP client requested a redirect URI that is not registered.
            This usually means the client is misconfigured. Reconnect the MCP
            client to register a fresh redirect URI.
          </p>
        ),
      };
    case "no_login":
      return {
        title: "Login flow not configured",
        description: message,
        body: (
          <p>
            Configure application authentication (basic or OAuth2) before
            using MCP OAuth.
          </p>
        ),
      };
    default:
      return {
        title: "Authorization error",
        description: message || "Something went wrong with the OAuth flow.",
        body: (
          <p>
            Try reconnecting from your MCP client. If the issue persists,
            contact the administrator of this Qaynaq instance.
          </p>
        ),
      };
  }
}

export default function OAuthErrorPage() {
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code") || "";
  const message = searchParams.get("message") || "";
  const c = content(code, message);

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardHeader>
          <div className="flex items-center gap-2">
            <ShieldAlert className="h-5 w-5 text-amber-600 dark:text-amber-400" />
            <CardTitle>{c.title}</CardTitle>
          </div>
          <CardDescription>{c.description}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm leading-relaxed">
          {c.body}
          <p className="text-muted-foreground text-xs pt-2">
            You can safely close this tab.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
