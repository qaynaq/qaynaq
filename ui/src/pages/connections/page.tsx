import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DataTable } from "@/components/data-table";
import { Plus, Link2, Eye, EyeOff, ExternalLink, RefreshCw, Search } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { useToast } from "@/components/toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Connection } from "@/lib/entities";
import { fetchConnections, deleteConnection, fetchProviders, type Provider } from "@/lib/api";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function ConnectionsPage() {
  const { addToast } = useToast();
  const [connections, setConnections] = useState<Connection[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [showSecret, setShowSecret] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [connToDelete, setConnToDelete] = useState<Connection | null>(null);
  const [reAuthOpen, setReAuthOpen] = useState(false);
  const [reAuthConn, setReAuthConn] = useState<Connection | null>(null);
  const [reAuthData, setReAuthData] = useState({ clientId: "", clientSecret: "" });
  const [showReAuthSecret, setShowReAuthSecret] = useState(false);
  const [scopeSearch, setScopeSearch] = useState("");
  const [selectedScopes, setSelectedScopes] = useState<Set<string>>(new Set());
  const [formData, setFormData] = useState({
    name: "",
    provider: "",
    clientId: "",
    clientSecret: "",
  });

  const handleDelete = async (conn: Connection) => {
    if (!conn || !conn.name) return;
    setConnToDelete(conn);
    setDeleteConfirmOpen(true);
  };

  const confirmDelete = async () => {
    if (!connToDelete) return;

    try {
      await deleteConnection(connToDelete.name);
      setConnections(connections.filter((c) => c.name !== connToDelete.name));
      addToast({
        id: "connection-deleted",
        title: "Connection Deleted",
        description: `Connection "${connToDelete.name}" has been deleted.`,
        variant: "success",
      });
    } catch (error) {
      addToast({
        id: "connection-delete-error",
        title: "Error Deleting Connection",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setDeleteConfirmOpen(false);
      setConnToDelete(null);
    }
  };

  const openOAuthPopup = (
    provider: string,
    name: string,
    clientId: string,
    clientSecret: string,
    scopes: string[],
    onSuccess: () => void,
  ) => {
    const params = new URLSearchParams({
      provider,
      name,
      client_id: clientId,
      client_secret: clientSecret,
    });
    if (scopes.length > 0) {
      params.set("scopes", scopes.join(","));
    }

    const popup = window.open(
      `/connections/oauth/authorize?${params.toString()}`,
      "oauth_authorize",
      "width=600,height=700,scrollbars=yes",
    );

    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === "oauth_callback") {
        window.removeEventListener("message", handleMessage);
        if (popup) popup.close();

        if (event.data.status === "success") {
          loadConnections();
          onSuccess();
          setTimeout(() => {
            addToast({
              id: "connection-saved",
              title: "Connection Saved",
              description: `Connection "${event.data.message}" has been saved.`,
              variant: "success",
            });
          }, 100);
        } else {
          addToast({
            id: "connection-error",
            title: "Authorization Failed",
            description: event.data.message || "Failed to authorize.",
            variant: "error",
          });
        }
      }
    };

    window.addEventListener("message", handleMessage);
  };

  const handleAuthorize = () => {
    if (
      !formData.name.trim() ||
      !formData.provider ||
      !formData.clientId.trim() ||
      !formData.clientSecret.trim() ||
      selectedScopes.size === 0
    ) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: "Name, Provider, Client ID, Client Secret, and at least one scope are required.",
        variant: "error",
      });
      return;
    }

    openOAuthPopup(
      formData.provider,
      formData.name.trim(),
      formData.clientId.trim(),
      formData.clientSecret.trim(),
      Array.from(selectedScopes),
      () => {
        setFormData({ name: "", provider: "", clientId: "", clientSecret: "" });
        setSelectedScopes(new Set());
        setScopeSearch("");
        setShowSecret(false);
        setIsModalOpen(false);
      },
    );
  };

  const [reAuthScopes, setReAuthScopes] = useState<Set<string>>(new Set());
  const [reAuthScopeSearch, setReAuthScopeSearch] = useState("");

  const handleReauthorize = (conn: Connection) => {
    setReAuthConn(conn);
    setReAuthData({ clientId: conn.clientId, clientSecret: "" });
    setReAuthScopes(new Set(conn.scopes || []));
    setReAuthScopeSearch("");
    setShowReAuthSecret(false);
    setReAuthOpen(true);
  };

  const submitReauthorize = () => {
    if (!reAuthConn || !reAuthData.clientId.trim() || reAuthScopes.size === 0) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: "Client ID and at least one scope are required.",
        variant: "error",
      });
      return;
    }

    openOAuthPopup(reAuthConn.provider, reAuthConn.name, reAuthData.clientId.trim(), reAuthData.clientSecret.trim(), Array.from(reAuthScopes), () => {
      setReAuthOpen(false);
      setReAuthConn(null);
      setReAuthData({ clientId: "", clientSecret: "" });
      setReAuthScopes(new Set());
      setReAuthScopeSearch("");
    });
  };

  const resetForm = () => {
    setFormData({ name: "", provider: "", clientId: "", clientSecret: "" });
    setSelectedScopes(new Set());
    setScopeSearch("");
    setShowSecret(false);
  };

  const handleCancel = () => {
    resetForm();
    setIsModalOpen(false);
  };

  const handleModalOpenChange = (open: boolean) => {
    setIsModalOpen(open);
    if (!open) {
      resetForm();
    }
  };

  const columns = [
    {
      key: "name" as keyof Connection,
      title: "Name",
      render: (value: string) => (
        <div className="flex items-center gap-2">
          <Link2 className="h-4 w-4 text-muted-foreground" />
          <span className="font-medium">{value || "Unknown"}</span>
        </div>
      ),
    },
    {
      key: "provider" as keyof Connection,
      title: "Provider",
      render: (value: string) => (
        <span className="capitalize">{value || "Unknown"}</span>
      ),
    },
    {
      key: "createdAt" as keyof Connection,
      title: "Created",
      render: (value: string) =>
        value && value !== "Unknown" ? (
          <RelativeTime dateString={value} />
        ) : (
          "Unknown"
        ),
    },
  ];

  const loadConnections = async () => {
    try {
      setLoading(true);
      const [connsData, providersData] = await Promise.all([
        fetchConnections(),
        fetchProviders(),
      ]);
      setConnections(connsData);
      setProviders(providersData);
      setError(null);
    } catch (err) {
      setError("Failed to fetch connections");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConnections();
  }, []);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Connections</h1>
          <p className="text-muted-foreground">
            Manage OAuth connections for Google and other services
          </p>
        </div>
        <Dialog open={isModalOpen} onOpenChange={handleModalOpenChange}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              New Connection
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>New Connection</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="conn-name">Connection Name</Label>
                <Input
                  id="conn-name"
                  type="text"
                  placeholder="e.g. my_google"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  autoComplete="off"
                />
              </div>

              <div className="space-y-2">
                <Label>Provider</Label>
                <Select
                  value={formData.provider}
                  onValueChange={(val) => {
                    setFormData({ ...formData, provider: val });
                    const provider = providers.find((p) => p.id === val);
                    if (provider) {
                      setSelectedScopes(new Set(provider.scopes.map((s) => s.scope)));
                    }
                    setScopeSearch("");
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a provider..." />
                  </SelectTrigger>
                  <SelectContent>
                    {providers.map((p) => (
                      <SelectItem key={p.id} value={p.id}>
                        {p.id.charAt(0).toUpperCase() + p.id.slice(1)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="client-id">OAuth Client ID</Label>
                <Input
                  id="client-id"
                  type="text"
                  placeholder="From GCP Console > APIs & Services > Credentials"
                  value={formData.clientId}
                  onChange={(e) =>
                    setFormData({ ...formData, clientId: e.target.value })
                  }
                  autoComplete="off"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="client-secret">OAuth Client Secret</Label>
                <div className="relative">
                  <Input
                    id="client-secret"
                    type={showSecret ? "text" : "password"}
                    placeholder="OAuth client secret"
                    value={formData.clientSecret}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        clientSecret: e.target.value,
                      })
                    }
                    className="pr-10"
                    autoComplete="new-password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowSecret(!showSecret)}
                    className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                  >
                    {showSecret ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>

              {formData.provider && (() => {
                const provider = providers.find((p) => p.id === formData.provider);
                if (!provider) return null;
                const filtered = provider.scopes.filter(
                  (s) =>
                    !scopeSearch ||
                    s.label.toLowerCase().includes(scopeSearch.toLowerCase()) ||
                    s.description.toLowerCase().includes(scopeSearch.toLowerCase()),
                );
                return (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label>Scopes</Label>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="text-xs text-primary hover:underline"
                          onClick={() =>
                            setSelectedScopes(new Set(provider.scopes.map((s) => s.scope)))
                          }
                        >
                          Select all
                        </button>
                        <button
                          type="button"
                          className="text-xs text-primary hover:underline"
                          onClick={() => setSelectedScopes(new Set())}
                        >
                          Deselect all
                        </button>
                      </div>
                    </div>
                    {provider.scopes.length > 5 && (
                      <div className="relative">
                        <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                          placeholder="Search scopes..."
                          value={scopeSearch}
                          onChange={(e) => setScopeSearch(e.target.value)}
                          className="pl-9 h-9"
                        />
                      </div>
                    )}
                    <div className="border rounded-md max-h-48 overflow-y-auto">
                      {filtered.map((scope) => (
                        <label
                          key={scope.scope}
                          className="flex items-start gap-3 px-3 py-2 hover:bg-muted/50 cursor-pointer"
                        >
                          <Checkbox
                            checked={selectedScopes.has(scope.scope)}
                            onCheckedChange={(checked) => {
                              const next = new Set(selectedScopes);
                              if (checked) {
                                next.add(scope.scope);
                              } else {
                                next.delete(scope.scope);
                              }
                              setSelectedScopes(next);
                            }}
                            className="mt-0.5"
                          />
                          <div className="flex-1 min-w-0">
                            <div className="text-sm font-medium">{scope.label}</div>
                            <div className="text-xs text-muted-foreground">{scope.description}</div>
                          </div>
                        </label>
                      ))}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {selectedScopes.size} of {provider.scopes.length} scopes selected
                    </p>
                  </div>
                );
              })()}

              <p className="text-xs text-muted-foreground">
                Create an OAuth Client ID (Web application type) in{" "}
                <a
                  href="https://console.cloud.google.com/apis/credentials"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline inline-flex items-center gap-1"
                >
                  GCP Console
                  <ExternalLink className="h-3 w-3" />
                </a>
                . Add{" "}
                <code className="text-xs bg-muted px-1 rounded">
                  {window.location.origin}/connections/oauth/callback
                </code>{" "}
                as an authorized redirect URI.
              </p>

              <div className="flex justify-end space-x-2">
                <Button variant="outline" onClick={handleCancel}>
                  Cancel
                </Button>
                <Button onClick={handleAuthorize}>
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Authorize
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Connections</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p>Loading connections...</p>
          ) : error ? (
            <p className="text-red-500">{error}</p>
          ) : (
            <DataTable
              data={connections}
              columns={columns}
              onDelete={handleDelete}
              additionalActions={(conn) => (
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => handleReauthorize(conn)}
                  title="Re-authorize"
                >
                  <RefreshCw className="h-4 w-4" />
                </Button>
              )}
              getRowId={(conn) => conn?.name || `unknown-${Math.random()}`}
            />
          )}
        </CardContent>
      </Card>

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Connection</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the connection "
              {connToDelete?.name}"?
              <br />
              <br />
              <strong>Warning:</strong> Flows using this connection will stop
              working.
              <br />
              <br />
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setDeleteConfirmOpen(false);
                setConnToDelete(null);
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

      <Dialog open={reAuthOpen} onOpenChange={setReAuthOpen}>
        <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Re-authorize "{reAuthConn?.name}"</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Enter your OAuth credentials to re-authorize this connection with fresh tokens. You can also update the scopes.
            </p>

            <div className="space-y-2">
              <Label htmlFor="reauth-client-id">OAuth Client ID</Label>
              <Input
                id="reauth-client-id"
                type="text"
                placeholder="Client ID from GCP Console"
                value={reAuthData.clientId}
                onChange={(e) =>
                  setReAuthData({ ...reAuthData, clientId: e.target.value })
                }
                autoComplete="off"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="reauth-client-secret">OAuth Client Secret</Label>
              <div className="relative">
                <Input
                  id="reauth-client-secret"
                  type={showReAuthSecret ? "text" : "password"}
                  placeholder={reAuthConn?.clientSecretHint || "Client secret"}
                  value={reAuthData.clientSecret}
                  onChange={(e) =>
                    setReAuthData({ ...reAuthData, clientSecret: e.target.value })
                  }
                  className="pr-10"
                  autoComplete="new-password"
                />
                <button
                  type="button"
                  onClick={() => setShowReAuthSecret(!showReAuthSecret)}
                  className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                >
                  {showReAuthSecret ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            {reAuthConn && (() => {
              const provider = providers.find((p) => p.id === reAuthConn.provider);
              if (!provider) return null;
              const filtered = provider.scopes.filter(
                (s) =>
                  !reAuthScopeSearch ||
                  s.label.toLowerCase().includes(reAuthScopeSearch.toLowerCase()) ||
                  s.description.toLowerCase().includes(reAuthScopeSearch.toLowerCase()),
              );
              return (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label>Scopes</Label>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        className="text-xs text-primary hover:underline"
                        onClick={() =>
                          setReAuthScopes(new Set(provider.scopes.map((s) => s.scope)))
                        }
                      >
                        Select all
                      </button>
                      <button
                        type="button"
                        className="text-xs text-primary hover:underline"
                        onClick={() => setReAuthScopes(new Set())}
                      >
                        Deselect all
                      </button>
                    </div>
                  </div>
                  {provider.scopes.length > 5 && (
                    <div className="relative">
                      <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Search scopes..."
                        value={reAuthScopeSearch}
                        onChange={(e) => setReAuthScopeSearch(e.target.value)}
                        className="pl-9 h-9"
                      />
                    </div>
                  )}
                  <div className="border rounded-md max-h-48 overflow-y-auto">
                    {filtered.map((scope) => (
                      <label
                        key={scope.scope}
                        className="flex items-start gap-3 px-3 py-2 hover:bg-muted/50 cursor-pointer"
                      >
                        <Checkbox
                          checked={reAuthScopes.has(scope.scope)}
                          onCheckedChange={(checked) => {
                            const next = new Set(reAuthScopes);
                            if (checked) {
                              next.add(scope.scope);
                            } else {
                              next.delete(scope.scope);
                            }
                            setReAuthScopes(next);
                          }}
                          className="mt-0.5"
                        />
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium">{scope.label}</div>
                          <div className="text-xs text-muted-foreground">{scope.description}</div>
                        </div>
                      </label>
                    ))}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {reAuthScopes.size} of {provider.scopes.length} scopes selected
                  </p>
                </div>
              );
            })()}

            <div className="flex justify-end space-x-2">
              <Button
                variant="outline"
                onClick={() => {
                  setReAuthOpen(false);
                  setReAuthConn(null);
                }}
              >
                Cancel
              </Button>
              <Button onClick={submitReauthorize}>
                <ExternalLink className="mr-2 h-4 w-4" />
                Re-authorize
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
