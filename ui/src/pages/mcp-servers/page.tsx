import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2, Server, RefreshCw, Terminal } from "lucide-react";
import { useToast } from "@/components/toast";
import {
  Connection,
  MCPServer,
  MCPCatalogEntry,
  MCPServerLogs,
} from "@/lib/entities";
import {
  fetchMCPServers,
  createMCPServer,
  deleteMCPServer,
  restartMCPServer,
  fetchMCPServerLogs,
  fetchMCPCatalog,
  fetchConnections,
} from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

const MaintainerBadge = ({ maintainer }: { maintainer: string }) => {
  if (maintainer !== "official" && maintainer !== "community") return null;
  const isOfficial = maintainer === "official";
  return (
    <span
      className={
        "inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide " +
        (isOfficial
          ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
          : "bg-zinc-200 text-zinc-700 dark:bg-zinc-700 dark:text-zinc-200")
      }
    >
      {maintainer}
    </span>
  );
};

type Transport = "http" | "stdio";

export default function MCPServersPage() {
  const { addToast } = useToast();
  const [mcpServers, setMcpServers] = useState<MCPServer[]>([]);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [catalog, setCatalog] = useState<MCPCatalogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [isAddServerOpen, setIsAddServerOpen] = useState(false);
  const [isAddingServer, setIsAddingServer] = useState(false);
  const [transport, setTransport] = useState<Transport>("http");

  // HTTP form state
  const [newServerName, setNewServerName] = useState("");
  const [newServerUrl, setNewServerUrl] = useState("");
  const [newServerAuthType, setNewServerAuthType] = useState("none");
  const [newServerAuthHeader, setNewServerAuthHeader] = useState("");
  const [newServerAuthValue, setNewServerAuthValue] = useState("");
  const [newServerConnectionName, setNewServerConnectionName] = useState("");

  // Stdio form state
  const [selectedCatalogId, setSelectedCatalogId] = useState("");
  const [stdioEnv, setStdioEnv] = useState<Record<string, string>>({});

  const [deleteServerConfirmOpen, setDeleteServerConfirmOpen] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<MCPServer | null>(null);

  // Logs viewer state
  const [logsOpenForServer, setLogsOpenForServer] = useState<MCPServer | null>(
    null,
  );
  const [logs, setLogs] = useState<MCPServerLogs | null>(null);

  useEffect(() => {
    loadAll();
  }, []);

  const loadAll = async () => {
    try {
      const [servers, conns, cat] = await Promise.all([
        fetchMCPServers(),
        fetchConnections().catch(() => []),
        fetchMCPCatalog().catch(() => []),
      ]);
      setMcpServers(servers);
      setConnections(conns);
      setCatalog(cat);
    } catch (error) {
      addToast({
        id: "mcp-servers-load-error",
        title: "Failed to load MCP servers",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setNewServerName("");
    setNewServerUrl("");
    setNewServerAuthType("none");
    setNewServerAuthHeader("");
    setNewServerAuthValue("");
    setNewServerConnectionName("");
    setSelectedCatalogId("");
    setStdioEnv({});
    setTransport("http");
  };

  const selectedCatalog = catalog.find((e) => e.id === selectedCatalogId);

  const handleAddServer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newServerName.trim()) return;

    setIsAddingServer(true);
    try {
      let server: MCPServer;
      if (transport === "stdio") {
        if (!selectedCatalogId) {
          addToast({
            id: "stdio-no-catalog",
            title: "Pick a server type",
            description: "Choose an entry from the catalog before saving.",
            variant: "error",
          });
          setIsAddingServer(false);
          return;
        }
        server = await createMCPServer({
          name: newServerName.trim(),
          transport: "stdio",
          catalog_id: selectedCatalogId,
          env: stdioEnv,
        });
      } else {
        if (!newServerUrl.trim()) return;
        server = await createMCPServer({
          name: newServerName.trim(),
          transport: "http",
          url: newServerUrl.trim(),
          auth_type: newServerAuthType,
          auth_header: newServerAuthHeader.trim(),
          auth_value: newServerAuthValue.trim(),
          connection_name: newServerConnectionName,
        });
      }
      setMcpServers((prev) => [server, ...prev]);
      resetForm();
      setIsAddServerOpen(false);
      addToast({
        id: "server-created",
        title: "MCP Server Added",
        description: `Server "${server.name}" registered. Tools will appear within a few seconds.`,
        variant: "success",
      });
    } catch (error) {
      addToast({
        id: "server-create-error",
        title: "Error Adding Server",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsAddingServer(false);
    }
  };

  const handleDeleteServer = (server: MCPServer) => {
    setServerToDelete(server);
    setDeleteServerConfirmOpen(true);
  };

  const confirmDeleteServer = async () => {
    if (!serverToDelete) return;
    try {
      await deleteMCPServer(serverToDelete.id);
      setMcpServers((prev) => prev.filter((s) => s.id !== serverToDelete.id));
      addToast({
        id: "server-deleted",
        title: "Server Removed",
        description: `Server "${serverToDelete.name}" and its proxied tools have been removed.`,
        variant: "success",
      });
    } catch {
      addToast({
        id: "server-delete-error",
        title: "Error",
        description: "Failed to delete server",
        variant: "error",
      });
    } finally {
      setDeleteServerConfirmOpen(false);
      setServerToDelete(null);
    }
  };

  const handleRestart = async (server: MCPServer) => {
    try {
      await restartMCPServer(server.id);
      addToast({
        id: "server-restarted",
        title: "Restart Scheduled",
        description: `Server "${server.name}" will restart on the next sync.`,
        variant: "success",
      });
      await loadAll();
    } catch (error) {
      addToast({
        id: "server-restart-error",
        title: "Restart Failed",
        description:
          error instanceof Error ? error.message : "Unknown error",
        variant: "error",
      });
    }
  };

  const handleViewLogs = async (server: MCPServer) => {
    setLogsOpenForServer(server);
    setLogs(null);
    try {
      const data = await fetchMCPServerLogs(server.id);
      setLogs(data);
    } catch (error) {
      addToast({
        id: "server-logs-error",
        title: "Failed to load logs",
        description:
          error instanceof Error ? error.message : "Unknown error",
        variant: "error",
      });
    }
  };

  if (loading) {
    return (
      <div className="p-6">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">MCP Servers</h1>
        <p className="text-muted-foreground">
          Register external MCP servers to proxy their tools through Qaynaq's
          /mcp endpoint
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>MCP Servers</CardTitle>
              <CardDescription>
                Register external MCP servers to proxy their tools through
                Qaynaq's /mcp endpoint
              </CardDescription>
            </div>
            <Dialog
              open={isAddServerOpen}
              onOpenChange={(open) => {
                setIsAddServerOpen(open);
                if (!open) resetForm();
              }}
            >
              <DialogTrigger asChild>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Server
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-md">
                <DialogHeader>
                  <DialogTitle>Add MCP Server</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleAddServer} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="server-name">Name</Label>
                    <Input
                      id="server-name"
                      placeholder="e.g., slack, github, stripe"
                      value={newServerName}
                      onChange={(e) => setNewServerName(e.target.value)}
                      required
                    />
                    <p className="text-xs text-muted-foreground">
                      Used as a prefix for tool names (e.g.,
                      slack__send_message)
                    </p>
                  </div>

                  <div className="space-y-2">
                    <Label>Server type</Label>
                    <Select
                      value={transport}
                      onValueChange={(v) => setTransport(v as Transport)}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="http">
                          Remote (HTTP URL)
                        </SelectItem>
                        <SelectItem value="stdio">
                          Local (command / npx)
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  {transport === "http" && (
                    <>
                      <div className="space-y-2">
                        <Label htmlFor="server-url">Server URL</Label>
                        <Input
                          id="server-url"
                          placeholder="https://mcp-server.example.com/mcp"
                          value={newServerUrl}
                          onChange={(e) => setNewServerUrl(e.target.value)}
                          required
                        />
                      </div>
                      <div className="space-y-2">
                        <Label>Authentication</Label>
                        <Select
                          value={newServerAuthType}
                          onValueChange={setNewServerAuthType}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">
                              No authentication
                            </SelectItem>
                            <SelectItem value="token">
                              Bearer token / API key
                            </SelectItem>
                            {connections.length > 0 && (
                              <SelectItem value="connection">
                                OAuth connection
                              </SelectItem>
                            )}
                          </SelectContent>
                        </Select>
                      </div>
                      {newServerAuthType === "token" && (
                        <>
                          <div className="space-y-2">
                            <Label htmlFor="server-auth-value">Token</Label>
                            <Input
                              id="server-auth-value"
                              type="password"
                              placeholder="e.g., xoxb-..., sk-..., or your API key"
                              value={newServerAuthValue}
                              onChange={(e) =>
                                setNewServerAuthValue(e.target.value)
                              }
                            />
                            <p className="text-xs text-muted-foreground">
                              Sent as{" "}
                              <code className="bg-muted px-1 py-0.5 rounded">
                                Authorization: Bearer {"{token}"}
                              </code>{" "}
                              by default
                            </p>
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="server-auth-header">
                              Custom Header{" "}
                              <span className="text-muted-foreground">
                                (optional)
                              </span>
                            </Label>
                            <Input
                              id="server-auth-header"
                              placeholder="e.g., X-API-Key (leave empty for Bearer auth)"
                              value={newServerAuthHeader}
                              onChange={(e) =>
                                setNewServerAuthHeader(e.target.value)
                              }
                            />
                          </div>
                        </>
                      )}
                      {newServerAuthType === "connection" && (
                        <div className="space-y-2">
                          <Label>Connection</Label>
                          <Select
                            value={newServerConnectionName}
                            onValueChange={setNewServerConnectionName}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder="Select a connection" />
                            </SelectTrigger>
                            <SelectContent>
                              {connections.map((conn) => (
                                <SelectItem key={conn.name} value={conn.name}>
                                  {conn.name} ({conn.provider})
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <p className="text-xs text-muted-foreground">
                            Uses the OAuth access token from this connection.
                            Token refresh is automatic.
                          </p>
                        </div>
                      )}
                    </>
                  )}

                  {transport === "stdio" && (
                    <>
                      <div className="space-y-2">
                        <Label>Catalog</Label>
                        <Select
                          value={selectedCatalogId}
                          onValueChange={(id) => {
                            setSelectedCatalogId(id);
                            setStdioEnv({});
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Pick a server" />
                          </SelectTrigger>
                          <SelectContent>
                            {catalog.map((entry) => (
                              <SelectItem key={entry.id} value={entry.id}>
                                <span className="flex items-center gap-2">
                                  {entry.display_name}
                                  <MaintainerBadge
                                    maintainer={entry.maintainer}
                                  />
                                </span>
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        {selectedCatalog && (
                          <p className="text-xs text-muted-foreground flex items-center gap-2 flex-wrap">
                            <MaintainerBadge
                              maintainer={selectedCatalog.maintainer}
                            />
                            <span>{selectedCatalog.description}</span>
                            {selectedCatalog.docs_url && (
                              <a
                                href={selectedCatalog.docs_url}
                                target="_blank"
                                rel="noreferrer"
                                className="underline"
                              >
                                docs
                              </a>
                            )}
                          </p>
                        )}
                      </div>

                      {selectedCatalog &&
                        selectedCatalog.env_spec.map((spec) => (
                          <div key={spec.name} className="space-y-2">
                            <Label htmlFor={`env-${spec.name}`}>
                              {spec.name}
                              {spec.required && (
                                <span className="text-destructive"> *</span>
                              )}
                            </Label>
                            <Input
                              id={`env-${spec.name}`}
                              type={spec.secret ? "password" : "text"}
                              placeholder={
                                spec.secret
                                  ? "value or ${SECRET_KEY}"
                                  : "value"
                              }
                              value={stdioEnv[spec.name] || ""}
                              onChange={(e) =>
                                setStdioEnv((prev) => ({
                                  ...prev,
                                  [spec.name]: e.target.value,
                                }))
                              }
                              required={spec.required}
                            />
                            <p className="text-xs text-muted-foreground">
                              {spec.description}
                              {spec.secret && (
                                <>
                                  {" "}
                                  Use{" "}
                                  <code className="bg-muted px-1 rounded">
                                    ${"{KEY}"}
                                  </code>{" "}
                                  to reference a saved secret.
                                </>
                              )}
                            </p>
                          </div>
                        ))}

                      <p className="text-xs text-muted-foreground">
                        Command:{" "}
                        <code className="bg-muted px-1 rounded">
                          {selectedCatalog
                            ? [
                                selectedCatalog.command,
                                ...selectedCatalog.args,
                              ].join(" ")
                            : "select a catalog entry"}
                        </code>
                      </p>
                    </>
                  )}

                  <div className="flex justify-end gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        setIsAddServerOpen(false);
                        resetForm();
                      }}
                      disabled={isAddingServer}
                    >
                      Cancel
                    </Button>
                    <Button type="submit" disabled={isAddingServer}>
                      {isAddingServer ? "Adding..." : "Add Server"}
                    </Button>
                  </div>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        <CardContent>
          {mcpServers.length === 0 ? (
            <div className="text-center py-8">
              <Server className="h-8 w-8 mx-auto text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">
                No external MCP servers registered yet
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Add a server to proxy its tools through Qaynaq's /mcp endpoint
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {mcpServers.map((server) => {
                const isStdio = server.transport === "stdio";
                return (
                  <div
                    key={server.id}
                    className="flex items-center justify-between border rounded-lg px-4 py-3"
                  >
                    <div className="space-y-1 min-w-0 flex-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <p className="font-medium text-sm">{server.name}</p>
                        <Badge
                          variant={
                            server.status === "active"
                              ? "default"
                              : server.status === "error"
                                ? "destructive"
                                : "secondary"
                          }
                          className="text-xs"
                        >
                          {server.status}
                        </Badge>
                        {isStdio && server.process_state && (
                          <Badge variant="outline" className="text-xs">
                            {server.process_state}
                          </Badge>
                        )}
                        <Badge variant="outline" className="text-xs">
                          {isStdio ? "local" : "remote"}
                        </Badge>
                        {server.tool_count > 0 && (
                          <span className="text-xs text-muted-foreground">
                            {server.tool_count} tools
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-muted-foreground truncate">
                        {isStdio
                          ? `catalog: ${server.catalog_id}`
                          : server.url}
                        {!isStdio && server.auth_type === "token" && (
                          <span className="ml-2">(token auth)</span>
                        )}
                        {!isStdio &&
                          server.auth_type === "connection" &&
                          server.connection_name && (
                            <span className="ml-2">
                              (via {server.connection_name})
                            </span>
                          )}
                      </p>
                      {server.last_error && server.status === "error" && (
                        <p className="text-xs text-destructive truncate">
                          {server.last_error}
                        </p>
                      )}
                      <div className="flex gap-3 text-xs text-muted-foreground">
                        <span>
                          Added{" "}
                          <RelativeTime dateString={server.created_at} />
                        </span>
                        {server.last_sync_at && (
                          <span>
                            Synced{" "}
                            <RelativeTime dateString={server.last_sync_at} />
                          </span>
                        )}
                      </div>
                    </div>
                    {isStdio && (
                      <>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleViewLogs(server)}
                          className="text-muted-foreground hover:text-foreground"
                          title="View logs"
                        >
                          <Terminal className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleRestart(server)}
                          className="text-muted-foreground hover:text-foreground"
                          title="Restart"
                        >
                          <RefreshCw className="h-4 w-4" />
                        </Button>
                      </>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDeleteServer(server)}
                      className="text-muted-foreground hover:text-destructive ml-2"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <AlertDialog
        open={deleteServerConfirmOpen}
        onOpenChange={setDeleteServerConfirmOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove MCP Server</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove the server "
              {serverToDelete?.name}"? All proxied tools from this server will
              be removed from the /mcp endpoint immediately. This action cannot
              be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setDeleteServerConfirmOpen(false);
                setServerToDelete(null);
              }}
            >
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDeleteServer}
              className="bg-red-600 hover:bg-red-700 text-white"
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Dialog
        open={!!logsOpenForServer}
        onOpenChange={(open) => {
          if (!open) {
            setLogsOpenForServer(null);
            setLogs(null);
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              Logs for {logsOpenForServer?.name}
            </DialogTitle>
          </DialogHeader>
          {logs ? (
            <div className="space-y-3">
              <div>
                <Label>Process state</Label>
                <p className="text-sm">{logs.process_state || "unknown"}</p>
              </div>
              {logs.last_error && (
                <div>
                  <Label>Last error</Label>
                  <Textarea
                    readOnly
                    value={logs.last_error}
                    className="font-mono text-xs"
                  />
                </div>
              )}
              <div>
                <Label>Stderr (last 4 KB)</Label>
                <Textarea
                  readOnly
                  rows={10}
                  value={logs.stderr || "(empty)"}
                  className="font-mono text-xs"
                />
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">Loading...</p>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
