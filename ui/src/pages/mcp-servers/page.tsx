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
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2, Server } from "lucide-react";
import { useToast } from "@/components/toast";
import { Connection, MCPServer } from "@/lib/entities";
import {
  fetchMCPServers,
  createMCPServer,
  deleteMCPServer,
  fetchConnections,
} from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function MCPServersPage() {
  const { addToast } = useToast();
  const [mcpServers, setMcpServers] = useState<MCPServer[]>([]);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [loading, setLoading] = useState(true);
  const [isAddServerOpen, setIsAddServerOpen] = useState(false);
  const [isAddingServer, setIsAddingServer] = useState(false);
  const [newServerName, setNewServerName] = useState("");
  const [newServerUrl, setNewServerUrl] = useState("");
  const [newServerAuthType, setNewServerAuthType] = useState("none");
  const [newServerAuthHeader, setNewServerAuthHeader] = useState("");
  const [newServerAuthValue, setNewServerAuthValue] = useState("");
  const [newServerConnectionName, setNewServerConnectionName] = useState("");
  const [deleteServerConfirmOpen, setDeleteServerConfirmOpen] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<MCPServer | null>(null);

  useEffect(() => {
    loadMCPServers();
  }, []);

  const loadMCPServers = async () => {
    try {
      const [servers, conns] = await Promise.all([
        fetchMCPServers(),
        fetchConnections().catch(() => []),
      ]);
      setMcpServers(servers);
      setConnections(conns);
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

  const handleAddServer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newServerName.trim() || !newServerUrl.trim()) return;

    setIsAddingServer(true);
    try {
      const server = await createMCPServer({
        name: newServerName.trim(),
        url: newServerUrl.trim(),
        auth_type: newServerAuthType,
        auth_header: newServerAuthHeader.trim(),
        auth_value: newServerAuthValue.trim(),
        connection_name: newServerConnectionName,
      });
      setMcpServers((prev) => [server, ...prev]);
      setNewServerName("");
      setNewServerUrl("");
      setNewServerAuthType("none");
      setNewServerAuthHeader("");
      setNewServerAuthValue("");
      setNewServerConnectionName("");
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
            <Dialog open={isAddServerOpen} onOpenChange={setIsAddServerOpen}>
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
                        <SelectItem value="none">No authentication</SelectItem>
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
                        Uses the OAuth access token from this connection. Token
                        refresh is automatic.
                      </p>
                    </div>
                  )}
                  <div className="flex justify-end gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setIsAddServerOpen(false)}
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
              {mcpServers.map((server) => (
                <div
                  key={server.id}
                  className="flex items-center justify-between border rounded-lg px-4 py-3"
                >
                  <div className="space-y-1 min-w-0 flex-1">
                    <div className="flex items-center gap-2">
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
                      {server.tool_count > 0 && (
                        <span className="text-xs text-muted-foreground">
                          {server.tool_count} tools
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground truncate">
                      {server.url}
                      {server.auth_type === "token" && (
                        <span className="ml-2">(token auth)</span>
                      )}
                      {server.auth_type === "connection" &&
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
                        Added <RelativeTime dateString={server.created_at} />
                      </span>
                      {server.last_sync_at && (
                        <span>
                          Synced{" "}
                          <RelativeTime dateString={server.last_sync_at} />
                        </span>
                      )}
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDeleteServer(server)}
                    className="text-muted-foreground hover:text-destructive ml-2"
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
    </div>
  );
}
