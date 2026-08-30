import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Check, ChevronsUpDown, TriangleAlert, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { fetchGroups } from "@/lib/api";
import type { GroupInfo } from "@/lib/entities";
import { cn } from "@/lib/utils";

function levenshtein(a: string, b: string): number {
  const m = a.length;
  const n = b.length;
  const row = Array.from({ length: n + 1 }, (_, j) => j);
  for (let i = 1; i <= m; i++) {
    let prev = row[0];
    row[0] = i;
    for (let j = 1; j <= n; j++) {
      const tmp = row[j];
      row[j] = Math.min(
        row[j] + 1,
        row[j - 1] + 1,
        prev + (a[i - 1] === b[j - 1] ? 0 : 1),
      );
      prev = tmp;
    }
  }
  return row[n];
}

interface Props {
  value: string[];
  onChange: (next: string[]) => void;
  compact?: boolean;
}

// Picker-only by design: group names come from the registry (observed at
// login or imported in Settings), never free-typed per tool, so a typo can
// only exist in one visible place.
export function GroupMultiSelect({ value, onChange, compact }: Props) {
  const [open, setOpen] = useState(false);
  const [groups, setGroups] = useState<GroupInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    fetchGroups()
      .then(setGroups)
      .catch(() => setGroups([]))
      .finally(() => setLoaded(true));
  }, []);

  const registry = useMemo(() => new Map(groups.map((g) => [g.name, g])), [groups]);
  const observed = useMemo(
    () => groups.filter((g) => g.last_seen_at).map((g) => g.name),
    [groups],
  );

  const toggle = (name: string) => {
    onChange(
      value.includes(name) ? value.filter((v) => v !== name) : [...value, name],
    );
  };

  const warningFor = (name: string): string | null => {
    const entry = registry.get(name);
    if (entry?.last_seen_at) return null;
    const nearMiss = observed.find(
      (o) => o !== name && levenshtein(o, name) <= 2,
    );
    if (nearMiss) {
      return `No login with group "${name}" has been seen yet. Did you mean "${nearMiss}"?`;
    }
    if (!entry) {
      return `Group "${name}" is not in the registry. Nobody can match it until a login carries it.`;
    }
    return `No login with group "${name}" has been seen yet. Members won't have access until their IdP sends this group.`;
  };

  return (
    <TooltipProvider>
      <div className="space-y-1.5">
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={open}
              className={cn(
                "w-full justify-between font-normal",
                compact ? "h-9 text-sm" : "",
              )}
            >
              {value.length === 0
                ? "Everyone"
                : `${value.length} group${value.length > 1 ? "s" : ""}`}
              <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-64 p-0" align="start">
            <Command>
              <CommandInput placeholder="Search groups..." />
              <CommandList>
                <CommandEmpty>
                  {loaded && groups.length === 0 ? (
                    <span className="text-xs">
                      No groups yet. Groups appear after users sign in, or{" "}
                      <Link
                        to="/settings/groups"
                        className="underline"
                        onClick={() => setOpen(false)}
                      >
                        import them
                      </Link>
                      .
                    </span>
                  ) : (
                    "No matching group."
                  )}
                </CommandEmpty>
                <CommandGroup>
                  {groups.map((g) => (
                    <CommandItem
                      key={g.name}
                      value={g.name}
                      onSelect={() => toggle(g.name)}
                    >
                      <Check
                        className={cn(
                          "mr-2 h-4 w-4",
                          value.includes(g.name) ? "opacity-100" : "opacity-0",
                        )}
                      />
                      <span className="flex-1 truncate">{g.name}</span>
                      {!g.last_seen_at && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <TriangleAlert className="h-3.5 w-3.5 text-amber-500" />
                          </TooltipTrigger>
                          <TooltipContent>
                            Imported manually; never seen at a login yet.
                          </TooltipContent>
                        </Tooltip>
                      )}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        {value.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {value.map((name) => {
              const warning = warningFor(name);
              return (
                <Badge key={name} variant="secondary" className="gap-1 pr-1">
                  {warning && (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <TriangleAlert className="h-3 w-3 text-amber-500" />
                      </TooltipTrigger>
                      <TooltipContent className="max-w-64">{warning}</TooltipContent>
                    </Tooltip>
                  )}
                  {name}
                  <button
                    type="button"
                    className="rounded-full hover:bg-muted-foreground/20"
                    onClick={() => toggle(name)}
                    aria-label={`Remove ${name}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              );
            })}
          </div>
        )}
      </div>
    </TooltipProvider>
  );
}
