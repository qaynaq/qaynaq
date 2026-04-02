import { useEffect, useState, useRef, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  ArrowLeft,
  ArrowDown,
  ChevronRight,
  ChevronDown,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { fetchFlowEvents, fetchStream } from "@/lib/api";
import { FlowEvent } from "@/lib/entities";

const TIME_RANGES = [
  { label: "Last 15 minutes", value: "15" },
  { label: "Last 30 minutes", value: "30" },
  { label: "Last 1 hour", value: "60" },
  { label: "Last 6 hours", value: "360" },
  { label: "Last 1 day", value: "1440" },
  { label: "Last 2 days", value: "2880" },
  { label: "Last 3 days", value: "4320" },
  { label: "Last 1 week", value: "10080" },
  { label: "Custom", value: "custom" },
];

const PAGE_SIZE = 50;

const SECTION_ORDER: Record<string, number> = {
  input: 0,
  pipeline: 1,
  output: 2,
};

type FlowGroup = {
  traceId: string;
  events: FlowEvent[];
  firstEvent: FlowEvent;
  lastTimestamp: string;
};

function groupByTraceId(events: FlowEvent[]): FlowGroup[] {
  const map = new Map<string, FlowEvent[]>();
  for (const event of events) {
    const key = event.trace_id || `no-flow-${event.id}`;
    const group = map.get(key);
    if (group) {
      group.push(event);
    } else {
      map.set(key, [event]);
    }
  }

  const groups: FlowGroup[] = [];
  for (const [traceId, flowEvents] of map) {
    const sorted = flowEvents.sort((a, b) => {
      const sectionDiff =
        (SECTION_ORDER[a.section] ?? 99) - (SECTION_ORDER[b.section] ?? 99);
      if (sectionDiff !== 0) return sectionDiff;
      return (
        new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      );
    });
    groups.push({
      traceId,
      events: sorted,
      firstEvent: sorted[0],
      lastTimestamp: sorted[sorted.length - 1].created_at,
    });
  }

  return groups.sort(
    (a, b) =>
      new Date(b.lastTimestamp).getTime() -
      new Date(a.lastTimestamp).getTime(),
  );
}

function sectionColor(section: string): string {
  switch (section) {
    case "input":
      return "bg-blue-500/20 text-blue-400 border-blue-500/30";
    case "pipeline":
      return "bg-yellow-500/20 text-yellow-400 border-yellow-500/30";
    case "output":
      return "bg-green-500/20 text-green-400 border-green-500/30";
    default:
      return "bg-gray-500/20 text-gray-400 border-gray-500/30";
  }
}

function typeBorderColor(type: string): string {
  switch (type) {
    case "PRODUCE":
      return "border-l-green-500";
    case "CONSUME":
      return "border-l-blue-500";
    case "ERROR":
      return "border-l-red-500";
    case "DELETE":
      return "border-l-gray-500";
    default:
      return "border-l-gray-600";
  }
}

function formatTimestamp(ts: string): string {
  const d = new Date(ts);
  const opts: Intl.DateTimeFormatOptions & { fractionalSecondDigits?: number } =
    {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      fractionalSecondDigits: 3,
    };
  return d.toLocaleTimeString("en-US", opts);
}

function FlowRow({ group }: { group: FlowGroup }) {
  const [expanded, setExpanded] = useState(false);
  const inputEvent = group.events.find((e) => e.section === "input");
  const displayEvent = inputEvent || group.firstEvent;
  const borderClass = group.events.some((e) => e.type === "ERROR")
    ? "border-l-red-500"
    : typeBorderColor(displayEvent.type);

  return (
    <div className="border-b border-gray-800 last:border-b-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className={`w-full text-left px-4 py-2.5 hover:bg-gray-800/50 transition-colors flex items-center gap-3 border-l-2 ${borderClass} min-w-0`}
      >
        <span className="text-gray-500 flex-shrink-0">
          {expanded ? (
            <ChevronDown className="h-3.5 w-3.5" />
          ) : (
            <ChevronRight className="h-3.5 w-3.5" />
          )}
        </span>
        <span className="text-gray-500 font-mono text-xs flex-shrink-0">
          {formatTimestamp(displayEvent.created_at)}
        </span>
        {!expanded && (
          <span className="text-gray-300 font-mono text-xs flex-1 min-w-0 truncate">
            {displayEvent.content}
          </span>
        )}
        {expanded && (
          <span className="text-gray-400 font-mono text-xs flex-1 min-w-0 truncate">
            Message journey of {group.traceId}
          </span>
        )}
        <Badge
          variant="outline"
          className="text-[10px] px-1.5 py-0 border-gray-700 text-gray-400 flex-shrink-0"
        >
          {group.events.length} event{group.events.length !== 1 ? "s" : ""}
        </Badge>
      </button>

      {expanded && (
        <div className="bg-gray-900/50 border-l-2 border-l-gray-700 py-2 overflow-hidden">
          {group.events.map((event, idx) => (
            <div key={event.id}>
              {idx > 0 && (
                <div className="flex justify-center py-1">
                  <ArrowDown className="h-3 w-3 text-gray-600" />
                </div>
              )}
              <div
                className={`mx-4 px-3 py-2 rounded border-l-2 ${typeBorderColor(event.type)} bg-gray-800/40 overflow-hidden`}
              >
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-gray-500 font-mono text-xs flex-shrink-0">
                    {formatTimestamp(event.created_at)}
                  </span>
                  <Badge
                    variant="outline"
                    className={`text-[10px] px-1.5 py-0 flex-shrink-0 ${sectionColor(event.section)}`}
                  >
                    {event.section}
                  </Badge>
                  {event.component_label && (
                    <span className="text-cyan-400 font-mono text-xs truncate">
                      {event.component_label}
                    </span>
                  )}
                  <Badge
                    variant="outline"
                    className={`text-[10px] px-1.5 py-0 flex-shrink-0 ${
                      event.type === "ERROR"
                        ? "bg-red-500/20 text-red-400 border-red-500/30"
                        : "border-gray-700 text-gray-400"
                    }`}
                  >
                    {event.type}
                  </Badge>
                </div>
                <div className="text-gray-300 font-mono text-xs mt-1 break-all whitespace-pre-wrap">
                  {event.content}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function FlowEventsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [streamName, setStreamName] = useState<string>("");
  const [events, setEvents] = useState<FlowEvent[]>([]);
  const [totalFlows, setTotalFlows] = useState(0);
  const [loadedFlowPages, setLoadedFlowPages] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [timeRange, setTimeRange] = useState("30");
  const [customStart, setCustomStart] = useState("");
  const [customEnd, setCustomEnd] = useState("");
  const [customOpen, setCustomOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  const toLocalISO = (d: Date): string => {
    const pad = (n: number, len = 2) => n.toString().padStart(len, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}.${pad(d.getMilliseconds(), 3)}Z`;
  };

  const toDatetimeLocal = (d: Date): string => {
    const pad = (n: number) => n.toString().padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  };

  const getTimeParams = useCallback(
    () => {
      if (timeRange === "custom") {
        return {
          startTime: toLocalISO(new Date(customStart)),
          endTime: toLocalISO(new Date(customEnd)),
        };
      }
      const end = new Date();
      const start = new Date(end.getTime() - parseInt(timeRange) * 60 * 1000);
      return {
        startTime: toLocalISO(start),
        endTime: toLocalISO(end),
      };
    },
    [timeRange, customStart, customEnd],
  );

  const handleTimeRangeChange = (value: string) => {
    if (value === "custom") {
      const now = new Date();
      const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
      setCustomStart(toDatetimeLocal(oneHourAgo));
      setCustomEnd(toDatetimeLocal(now));
      setCustomOpen(true);
    } else {
      setTimeRange(value);
    }
  };

  const applyCustomRange = () => {
    if (customStart && customEnd) {
      setCustomOpen(false);
      setTimeRange("custom");
    }
  };

  const formatCustomLabel = (): string => {
    if (!customStart || !customEnd) return "Custom";
    const fmt = (v: string) => {
      const d = new Date(v);
      const pad = (n: number) => n.toString().padStart(2, "0");
      return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
    };
    return `${fmt(customStart)} – ${fmt(customEnd)}`;
  };

  const loadEvents = useCallback(
    async (reset: boolean) => {
      if (!id) return;

      try {
        if (reset) {
          setLoading(true);
          setEvents([]);
          setLoadedFlowPages(0);
        } else {
          setLoadingMore(true);
        }

        const currentPage = reset ? 0 : loadedFlowPages;
        const offset = currentPage * PAGE_SIZE;
        const { startTime, endTime } = getTimeParams();

        const result = await fetchFlowEvents(id, {
          limit: PAGE_SIZE,
          offset,
          startTime,
          endTime,
        });

        if (reset) {
          setEvents(result.data);
          setLoadedFlowPages(1);
        } else {
          setEvents((prev) => [...prev, ...result.data]);
          setLoadedFlowPages((prev) => prev + 1);
        }
        setTotalFlows(result.total);
        setError(null);
      } catch (err) {
        setError("Failed to fetch events");
        console.error(err);
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [id, loadedFlowPages, timeRange, getTimeParams],
  );

  useEffect(() => {
    if (!id) return;
    fetchStream(id)
      .then((flow) => setStreamName(flow.name))
      .catch(() => {});
  }, [id]);

  useEffect(() => {
    loadEvents(true);
  }, [id, timeRange]);

  const handleScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el || loadingMore || loading) return;
    if (loadedFlowPages * PAGE_SIZE >= totalFlows) return;

    const { scrollTop, scrollHeight, clientHeight } = el;
    if (scrollHeight - scrollTop - clientHeight < 100) {
      loadEvents(false);
    }
  }, [loadingMore, loading, loadedFlowPages, totalFlows, loadEvents]);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    el.addEventListener("scroll", handleScroll);
    return () => el.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  const flowGroups = groupByTraceId(events);

  return (
    <div className="p-6 flex flex-col h-[calc(100vh-5rem)] max-w-full overflow-hidden">
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/flows")}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Events</h1>
            {streamName && (
              <p className="text-sm text-muted-foreground">{streamName}</p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">
            {totalFlows} flows
          </span>
          <Popover open={customOpen} onOpenChange={setCustomOpen}>
            <PopoverTrigger asChild>
              <div>
                <Select
                  value={timeRange === "custom" ? "custom" : timeRange}
                  onValueChange={handleTimeRangeChange}
                >
                  <SelectTrigger className={timeRange === "custom" ? "w-[320px]" : "w-[180px]"}>
                    {timeRange === "custom" ? (
                      <span className="truncate">
                        {formatCustomLabel()}
                      </span>
                    ) : (
                      <SelectValue />
                    )}
                  </SelectTrigger>
                  <SelectContent>
                    {TIME_RANGES.map((range) => (
                      <SelectItem key={range.value} value={range.value}>
                        {range.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-4" align="end">
              <div className="flex flex-col gap-3">
                <p className="text-sm font-medium">Custom time range</p>
                <div className="flex flex-col gap-2">
                  <label className="text-xs text-muted-foreground">
                    Start
                  </label>
                  <input
                    type="datetime-local"
                    value={customStart}
                    onChange={(e) => setCustomStart(e.target.value)}
                    className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <label className="text-xs text-muted-foreground">End</label>
                  <input
                    type="datetime-local"
                    value={customEnd}
                    onChange={(e) => setCustomEnd(e.target.value)}
                    className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  />
                </div>
                <Button
                  size="sm"
                  onClick={applyCustomRange}
                  disabled={!customStart || !customEnd}
                >
                  Apply
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>

      <div
        ref={scrollRef}
        className="flex-1 min-h-0 min-w-0 overflow-y-auto overflow-x-hidden rounded-lg border border-gray-800 bg-gray-950 font-mono text-sm"
      >
        {loading ? (
          <div className="flex items-center justify-center h-32 text-gray-500">
            <Loader2 className="h-5 w-5 animate-spin mr-2" />
            Loading events...
          </div>
        ) : error ? (
          <div className="flex items-center justify-center h-32 text-red-400">
            {error}
          </div>
        ) : flowGroups.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500">
            No events found in the selected time range.
          </div>
        ) : (
          <>
            {flowGroups.map((group) => (
              <FlowRow key={group.traceId} group={group} />
            ))}
            {loadingMore && (
              <div className="flex items-center justify-center py-3 text-gray-500">
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Loading more...
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
