import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { Trash2, Upload } from "lucide-react";
import { useToast } from "@/components/toast";
import { GroupInfo } from "@/lib/entities";
import { deleteGroup, fetchGroups, importGroups } from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function GroupsSettings() {
  const { addToast } = useToast();
  const [groups, setGroups] = useState<GroupInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [importOpen, setImportOpen] = useState(false);
  const [importText, setImportText] = useState("");
  const [importing, setImporting] = useState(false);

  const loadGroups = async () => {
    try {
      setLoading(true);
      setGroups(await fetchGroups());
    } catch {
      addToast({
        id: "groups-load-error",
        title: "Error",
        description: "Failed to load groups",
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadGroups();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleImport = async () => {
    const names = importText
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
    if (names.length === 0) return;
    try {
      setImporting(true);
      const imported = await importGroups(names);
      addToast({
        id: "groups-imported",
        title: "Groups imported",
        description: `${imported} new group${imported === 1 ? "" : "s"} added, ${names.length - imported} already existed.`,
        variant: "success",
      });
      setImportOpen(false);
      setImportText("");
      await loadGroups();
    } catch (error) {
      addToast({
        id: "groups-import-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to import groups",
        variant: "error",
      });
    } finally {
      setImporting(false);
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteGroup(name);
      setGroups((prev) => prev.filter((g) => g.name !== name));
      addToast({
        id: "group-deleted",
        title: "Group removed",
        description: `"${name}" removed from the registry. Tools tagged with it keep their tag.`,
        variant: "success",
      });
    } catch {
      addToast({
        id: "group-delete-error",
        title: "Error",
        description: "Failed to delete group",
        variant: "error",
      });
    }
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading...</p>;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Groups</CardTitle>
            <CardDescription>
              IdP groups used to restrict who can call MCP tools. Groups are
              recorded automatically when users sign in; import them by hand to
              tag tools before the first login. Access checks always use the
              live groups from the caller's identity provider, never this list.
            </CardDescription>
          </div>
          <Dialog open={importOpen} onOpenChange={setImportOpen}>
            <DialogTrigger asChild>
              <Button className="shrink-0">
                <Upload className="h-4 w-4 mr-2" />
                Import groups
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Import groups</DialogTitle>
              </DialogHeader>
              <p className="text-sm text-muted-foreground">
                Paste group names from your identity provider's admin console,
                one per line. Copy-paste the exact names; access matching is
                exact.
              </p>
              <Textarea
                value={importText}
                onChange={(e) => setImportText(e.target.value)}
                placeholder={"accounting\nlegal\nengineering"}
                rows={8}
              />
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setImportOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleImport}
                  disabled={importing || !importText.trim()}
                >
                  {importing ? "Importing..." : "Import"}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent>
        {groups.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No groups yet. They appear here automatically when users sign in
            through your identity provider (set
            AUTH_OAUTH2_GROUPS_ATTRIBUTE_PATH), or import them above.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Last seen at login</TableHead>
                <TableHead className="w-12" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {groups.map((group) => (
                <TableRow key={group.name}>
                  <TableCell className="font-medium">{group.name}</TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        group.source === "observed" ? "default" : "secondary"
                      }
                    >
                      {group.source}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {group.last_seen_at ? (
                      <RelativeTime dateString={group.last_seen_at} />
                    ) : (
                      "never"
                    )}
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(group.name)}
                      aria-label={`Delete ${group.name}`}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
