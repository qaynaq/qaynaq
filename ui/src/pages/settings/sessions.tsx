import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Trash2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { OAuthSession, OAuthSessions } from "@/lib/entities";
import { fetchOAuthSessions, revokeOAuthSession } from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function SessionsSettings() {
  const { addToast } = useToast();
  const [oauthSessions, setOauthSessions] = useState<OAuthSessions | null>(
    null,
  );
  const [loading, setLoading] = useState(true);
  const [sessionToRevoke, setSessionToRevoke] = useState<OAuthSession | null>(
    null,
  );
  const [revokeConfirmOpen, setRevokeConfirmOpen] = useState(false);

  useEffect(() => {
    loadOAuthSessions();
  }, []);

  const loadOAuthSessions = async () => {
    try {
      setLoading(true);
      const data = await fetchOAuthSessions();
      setOauthSessions(data);
    } catch {
      // Non-fatal: OAuth may not be enabled.
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeSession = (session: OAuthSession) => {
    setSessionToRevoke(session);
    setRevokeConfirmOpen(true);
  };

  const confirmRevokeSession = async () => {
    if (!sessionToRevoke) return;
    try {
      await revokeOAuthSession(sessionToRevoke.id);
      setOauthSessions((prev) =>
        prev
          ? {
              ...prev,
              sessions: prev.sessions.filter(
                (s) => s.id !== sessionToRevoke.id,
              ),
            }
          : prev,
      );
      addToast({
        id: "oauth-session-revoked",
        title: "Session revoked",
        description: `${sessionToRevoke.client_name} will be asked to log in again within the next hour.`,
        variant: "success",
      });
    } catch {
      addToast({
        id: "oauth-session-revoke-error",
        title: "Error",
        description: "Failed to revoke session",
        variant: "error",
      });
    } finally {
      setRevokeConfirmOpen(false);
      setSessionToRevoke(null);
    }
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading...</p>;
  }

  if (!oauthSessions || !oauthSessions.oauth_enabled) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Active MCP Sessions</CardTitle>
          <CardDescription>
            MCP OAuth is not enabled on this server. Set
            <code className="text-xs bg-muted px-1 py-0.5 rounded mx-1">
              MCP_OAUTH_ENABLED=true
            </code>
            to enable OAuth-based authentication.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Active MCP Sessions</CardTitle>
          <CardDescription>
            Each session is one MCP client connected to Qaynaq. Revoke a
            session to sign that client out; the user is asked to log in again
            the next time their access token expires (within an hour).
          </CardDescription>
        </CardHeader>
        <CardContent>
          {oauthSessions.sessions.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-6">
              No active sessions
            </p>
          ) : (
            <div className="space-y-3">
              {oauthSessions.sessions.map((session) => (
                <div
                  key={session.id}
                  className="flex items-center justify-between border rounded-lg px-4 py-3"
                >
                  <div className="space-y-1">
                    <p className="font-medium text-sm">
                      {session.client_name || session.client_id}
                    </p>
                    <div className="flex gap-3 text-xs text-muted-foreground">
                      <span>{session.user_email}</span>
                      <span>
                        Started{" "}
                        <RelativeTime dateString={session.created_at} />
                      </span>
                      <span>
                        Expires{" "}
                        <RelativeTime dateString={session.expires_at} />
                      </span>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRevokeSession(session)}
                    className="text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <AlertDialog
        open={revokeConfirmOpen}
        onOpenChange={setRevokeConfirmOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke session</AlertDialogTitle>
            <AlertDialogDescription>
              Revoke this session for{" "}
              {sessionToRevoke?.client_name || "the MCP client"}? Its refresh
              token will be invalidated; the user will be asked to log in
              again the next time their access token expires (within the next
              hour).
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setRevokeConfirmOpen(false);
                setSessionToRevoke(null);
              }}
            >
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmRevokeSession}
              className="bg-red-600 hover:bg-red-700 text-white"
            >
              Revoke
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
