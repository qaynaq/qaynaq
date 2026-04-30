import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { Trash2, AlertTriangle, ShieldOff, Check } from "lucide-react";
import { useToast } from "@/components/toast";
import { OAuthClient, OAuthClients } from "@/lib/entities";
import {
  fetchOAuthClients,
  deleteOAuthClient,
  revokeOAuthConsent,
} from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function ClientsSettings() {
  const { addToast } = useToast();
  const [oauthClients, setOauthClients] = useState<OAuthClients | null>(null);
  const [loading, setLoading] = useState(true);
  const [clientToDelete, setClientToDelete] = useState<OAuthClient | null>(
    null,
  );
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState("");
  const deleteConfirmed = deleteConfirmText.trim().toLowerCase() === "delete";
  const [clientToRevokeConsent, setClientToRevokeConsent] =
    useState<OAuthClient | null>(null);
  const [revokeConsentOpen, setRevokeConsentOpen] = useState(false);

  const closeDeleteDialog = () => {
    setDeleteConfirmOpen(false);
    setClientToDelete(null);
    setDeleteConfirmText("");
  };

  const closeRevokeConsentDialog = () => {
    setRevokeConsentOpen(false);
    setClientToRevokeConsent(null);
  };

  useEffect(() => {
    loadOAuthClients();
  }, []);

  const loadOAuthClients = async () => {
    try {
      setLoading(true);
      const data = await fetchOAuthClients();
      setOauthClients(data);
    } catch {
      // Non-fatal: OAuth may not be enabled.
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteClient = (client: OAuthClient) => {
    setClientToDelete(client);
    setDeleteConfirmOpen(true);
  };

  const confirmDeleteClient = async () => {
    if (!clientToDelete || !deleteConfirmed) return;
    try {
      await deleteOAuthClient(clientToDelete.id);
      setOauthClients((prev) =>
        prev
          ? {
              ...prev,
              clients: prev.clients.filter((c) => c.id !== clientToDelete.id),
            }
          : prev,
      );
      addToast({
        id: "oauth-client-deleted",
        title: "Client deleted",
        description: `"${clientToDelete.name}" has been removed. It must register again on next connect.`,
        variant: "success",
      });
    } catch {
      addToast({
        id: "oauth-client-delete-error",
        title: "Error",
        description: "Failed to delete OAuth client",
        variant: "error",
      });
    } finally {
      closeDeleteDialog();
    }
  };

  const handleRevokeConsent = (client: OAuthClient) => {
    setClientToRevokeConsent(client);
    setRevokeConsentOpen(true);
  };

  const confirmRevokeConsent = async () => {
    if (!clientToRevokeConsent) return;
    try {
      await revokeOAuthConsent(clientToRevokeConsent.id);
      setOauthClients((prev) =>
        prev
          ? {
              ...prev,
              clients: prev.clients.map((c) =>
                c.id === clientToRevokeConsent.id
                  ? { ...c, consented: false }
                  : c,
              ),
            }
          : prev,
      );
      addToast({
        id: "oauth-consent-revoked",
        title: "Consent revoked",
        description: `"${clientToRevokeConsent.name}" will prompt for consent on its next connection.`,
        variant: "success",
      });
    } catch {
      addToast({
        id: "oauth-consent-revoke-error",
        title: "Error",
        description: "Failed to revoke consent",
        variant: "error",
      });
    } finally {
      closeRevokeConsentDialog();
    }
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading...</p>;
  }

  if (!oauthClients || !oauthClients.oauth_enabled) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Registered MCP Clients</CardTitle>
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
          <CardTitle>Registered MCP Clients</CardTitle>
          <CardDescription>
            Every MCP client that has registered itself via Dynamic Client
            Registration. Deleting a client removes its registration and ends
            any active sessions; the same client will register again
            automatically the next time it connects.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {oauthClients.clients.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-6">
              No clients registered yet
            </p>
          ) : (
            <div className="space-y-3">
              {oauthClients.clients.map((client) => (
                <div
                  key={client.id}
                  className="flex items-center justify-between border rounded-lg px-4 py-3 gap-3"
                >
                  <div className="space-y-1 min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-sm">{client.name}</p>
                      {client.consented ? (
                        <span className="inline-flex items-center gap-1 text-xs text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 px-1.5 py-0.5 rounded">
                          <Check className="h-3 w-3" />
                          Consented
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                          Not consented
                        </span>
                      )}
                    </div>
                    <div className="flex gap-3 text-xs text-muted-foreground">
                      <span className="font-mono truncate">{client.id}</span>
                      <span>
                        Registered{" "}
                        <RelativeTime dateString={client.created_at} />
                      </span>
                      <span>
                        {client.last_used_at ? (
                          <>
                            Last used{" "}
                            <RelativeTime dateString={client.last_used_at} />
                          </>
                        ) : (
                          "Never used"
                        )}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-1 flex-shrink-0">
                    {client.consented && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRevokeConsent(client)}
                        className="text-muted-foreground hover:text-destructive gap-1"
                        title="Revoke consent and active sessions"
                      >
                        <ShieldOff className="h-4 w-4" />
                        <span className="hidden sm:inline">Revoke consent</span>
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDeleteClient(client)}
                      className="text-muted-foreground hover:text-destructive"
                      title="Delete client"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <AlertDialog
        open={deleteConfirmOpen}
        onOpenChange={(open) => {
          if (!open) closeDeleteDialog();
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete MCP client</AlertDialogTitle>
            <AlertDialogDescription asChild>
              <div className="space-y-3">
                <p>
                  Delete the registration for <strong>{clientToDelete?.name}</strong>{" "}
                  and all of its refresh tokens? This cannot be undone.
                </p>
                <div className="flex gap-2 rounded-md border border-amber-300 bg-amber-50 dark:border-amber-700 dark:bg-amber-950 p-3 text-sm text-amber-900 dark:text-amber-100">
                  <AlertTriangle className="h-4 w-4 mt-0.5 flex-shrink-0" />
                  <div className="space-y-1">
                    <p className="font-medium">Disconnect the client first</p>
                    <p>
                      Some MCP clients (notably <code className="text-xs bg-amber-100 dark:bg-amber-900 px-1 rounded">mcp-remote</code>{" "}
                      / Claude Code CLI) cache the deleted client_id and keep
                      retrying with it, which can produce a reconnection loop.
                    </p>
                    <p>
                      Before deleting, quit or disconnect this MCP client on
                      the user's machine. Prefer{" "}
                      <strong>Settings &gt; Sessions</strong> for routine sign-out;
                      only delete the client when you actually want the
                      registration gone.
                    </p>
                  </div>
                </div>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="space-y-2">
            <Label htmlFor="delete-confirm">
              Type <span className="font-mono font-semibold">delete</span> to
              confirm:
            </Label>
            <Input
              id="delete-confirm"
              value={deleteConfirmText}
              onChange={(e) => setDeleteConfirmText(e.target.value)}
              placeholder="delete"
              autoComplete="off"
              autoFocus
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={closeDeleteDialog}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDeleteClient}
              disabled={!deleteConfirmed}
              className="bg-red-600 hover:bg-red-700 text-white disabled:bg-red-300 disabled:cursor-not-allowed dark:disabled:bg-red-900/40"
            >
              Delete client
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={revokeConsentOpen}
        onOpenChange={(open) => {
          if (!open) closeRevokeConsentDialog();
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke consent</AlertDialogTitle>
            <AlertDialogDescription>
              Revoke consent for <strong>{clientToRevokeConsent?.name}</strong>?
              All active sessions for this client will be revoked, and the
              user will be prompted to consent again on the next connection.
              The client registration is kept.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={closeRevokeConsentDialog}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmRevokeConsent}
              className="bg-red-600 hover:bg-red-700 text-white"
            >
              Revoke consent
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
