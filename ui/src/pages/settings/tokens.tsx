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
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
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
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Plus, Trash2, Copy, Check, ShieldAlert } from "lucide-react";
import { useToast } from "@/components/toast";
import { APIToken, MCPSettings } from "@/lib/entities";
import {
  fetchMCPSettings,
  createAPIToken,
  deleteAPIToken,
} from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function TokensSettings() {
  const { addToast } = useToast();
  const [settings, setSettings] = useState<MCPSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [newTokenName, setNewTokenName] = useState("");
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [tokenToDelete, setTokenToDelete] = useState<APIToken | null>(null);

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
        description: "Failed to load API tokens",
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreateToken = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTokenName.trim()) return;
    setIsSubmitting(true);
    try {
      const token = await createAPIToken(newTokenName.trim(), ["mcp"]);
      setCreatedToken(token.token!);
      setSettings((prev) =>
        prev
          ? {
              ...prev,
              tokens: [{ ...token, token: undefined }, ...prev.tokens],
            }
          : prev,
      );
      setNewTokenName("");
    } catch (error) {
      addToast({
        id: "token-create-error",
        title: "Error Creating Token",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCopyToken = async () => {
    if (!createdToken) return;
    await navigator.clipboard.writeText(createdToken);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCloseCreateDialog = (open: boolean) => {
    if (!open) {
      setCreatedToken(null);
      setNewTokenName("");
      setCopied(false);
    }
    setIsCreateOpen(open);
  };

  const handleDeleteToken = (token: APIToken) => {
    setTokenToDelete(token);
    setDeleteConfirmOpen(true);
  };

  const confirmDelete = async () => {
    if (!tokenToDelete) return;
    try {
      await deleteAPIToken(tokenToDelete.id);
      setSettings((prev) =>
        prev
          ? {
              ...prev,
              tokens: prev.tokens.filter((t) => t.id !== tokenToDelete.id),
            }
          : prev,
      );
      addToast({
        id: "token-deleted",
        title: "Token Deleted",
        description: `Token "${tokenToDelete.name}" has been deleted.`,
        variant: "success",
      });
    } catch {
      addToast({
        id: "token-delete-error",
        title: "Error",
        description: "Failed to delete token",
        variant: "error",
      });
    } finally {
      setDeleteConfirmOpen(false);
      setTokenToDelete(null);
    }
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading...</p>;
  }
  if (!settings) return null;

  const authDisabled = !settings.auth_enabled;

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>API Tokens</CardTitle>
              <CardDescription>
                Tokens used to authenticate with protected endpoints
              </CardDescription>
            </div>
            <Dialog open={isCreateOpen} onOpenChange={handleCloseCreateDialog}>
              <DialogTrigger asChild>
                <Button disabled={authDisabled}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create Token
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-md">
                <DialogHeader>
                  <DialogTitle>
                    {createdToken ? "Token Created" : "Create API Token"}
                  </DialogTitle>
                </DialogHeader>
                {createdToken ? (
                  <div className="space-y-4">
                    <Alert>
                      <ShieldAlert className="h-4 w-4" />
                      <AlertTitle>Save this token now</AlertTitle>
                      <AlertDescription>
                        This token will only be shown once. Copy it and store
                        it securely.
                      </AlertDescription>
                    </Alert>
                    <div className="flex items-center gap-2">
                      <code className="flex-1 bg-muted px-3 py-2 rounded text-sm font-mono break-all">
                        {createdToken}
                      </code>
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={handleCopyToken}
                      >
                        {copied ? (
                          <Check className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    <div className="flex justify-end">
                      <Button onClick={() => handleCloseCreateDialog(false)}>
                        Done
                      </Button>
                    </div>
                  </div>
                ) : (
                  <form onSubmit={handleCreateToken} className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="token-name">Name</Label>
                      <Input
                        id="token-name"
                        placeholder="e.g., Claude Desktop, Production MCP Client"
                        value={newTokenName}
                        onChange={(e) => setNewTokenName(e.target.value)}
                        required
                      />
                    </div>
                    <div className="flex justify-end gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => handleCloseCreateDialog(false)}
                        disabled={isSubmitting}
                      >
                        Cancel
                      </Button>
                      <Button type="submit" disabled={isSubmitting}>
                        {isSubmitting ? "Creating..." : "Create"}
                      </Button>
                    </div>
                  </form>
                )}
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        <CardContent>
          {authDisabled && (
            <Alert className="mb-4">
              <ShieldAlert className="h-4 w-4" />
              <AlertDescription>
                Enable app authentication to create and manage tokens.
              </AlertDescription>
            </Alert>
          )}
          {settings.tokens.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-6">
              No API tokens created yet
            </p>
          ) : (
            <div className="space-y-3">
              {settings.tokens.map((token) => (
                <div
                  key={token.id}
                  className="flex items-center justify-between border rounded-lg px-4 py-3"
                >
                  <div className="space-y-1">
                    <p className="font-medium text-sm">{token.name}</p>
                    <div className="flex gap-3 text-xs text-muted-foreground">
                      <span>
                        Created <RelativeTime dateString={token.created_at} />
                      </span>
                      <span>
                        {token.last_used_at ? (
                          <>
                            Last used{" "}
                            <RelativeTime dateString={token.last_used_at} />
                          </>
                        ) : (
                          "Never used"
                        )}
                      </span>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDeleteToken(token)}
                    disabled={authDisabled}
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

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Token</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the token "{tokenToDelete?.name}"?
              Any MCP clients using this token will lose access immediately.
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setDeleteConfirmOpen(false);
                setTokenToDelete(null);
              }}
            >
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-red-600 hover:bg-red-700 text-white"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
