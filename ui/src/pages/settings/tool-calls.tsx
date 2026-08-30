import { useCallback, useEffect, useState } from "react";
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { ChevronLeft, ChevronRight, RefreshCw } from "lucide-react";
import { MCPCallLog } from "@/lib/entities";
import { fetchMCPCallLogs } from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

const PAGE_SIZE = 50;

const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

function statusVariant(status: string): "default" | "secondary" | "destructive" {
  if (status === "ok") return "secondary";
  if (status === "denied") return "destructive";
  return "default";
}

export default function ToolCallsSettings() {
  const [logs, setLogs] = useState<MCPCallLog[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async (nextOffset: number) => {
    try {
      setLoading(true);
      const data = await fetchMCPCallLogs(PAGE_SIZE, nextOffset);
      setLogs(data.logs);
      setTotal(data.total);
      setOffset(nextOffset);
    } catch {
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load(0);
  }, [load]);

  return (
    <TooltipProvider>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Tool Calls</CardTitle>
              <CardDescription>
                Every call made through the MCP endpoint: who called which
                tool, whether it succeeded, and how long it took.
              </CardDescription>
            </div>
            <Button
              variant="outline"
              size="icon"
              onClick={() => load(offset)}
              disabled={loading}
              aria-label="Refresh"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {logs.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {loading ? "Loading..." : "No tool calls recorded yet."}
            </p>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Time</TableHead>
                    <TableHead>Caller</TableHead>
                    <TableHead>Tool</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Duration</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="text-muted-foreground whitespace-nowrap">
                        <RelativeTime dateString={log.created_at} />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{log.actor}</span>
                          <Badge variant="outline">{log.actor_kind}</Badge>
                        </div>
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {log.tool_name}
                      </TableCell>
                      <TableCell>
                        {log.error ? (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Badge variant={statusVariant(log.status)}>
                                {log.status}
                              </Badge>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-96 break-words">
                              {log.error}
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          <Badge variant={statusVariant(log.status)}>
                            {log.status}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        {log.duration_ms} ms
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <div className="flex items-center justify-between mt-4">
                <p className="text-sm text-muted-foreground">
                  {offset + 1}-{Math.min(offset + PAGE_SIZE, total)} of {total}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="icon"
                    disabled={offset === 0 || loading}
                    onClick={() => load(Math.max(0, offset - PAGE_SIZE))}
                    aria-label="Previous page"
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    disabled={offset + PAGE_SIZE >= total || loading}
                    onClick={() => load(offset + PAGE_SIZE)}
                    aria-label="Next page"
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </TooltipProvider>
  );
}
