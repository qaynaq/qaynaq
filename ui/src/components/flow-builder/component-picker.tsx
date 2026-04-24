import { useMemo, useState } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  describeComponent,
  getComponentIcon,
  groupComponents,
} from "@/lib/component-catalog";

type NodeKind = "input" | "processor" | "output";

export interface PickableComponent {
  id: string;
  name: string;
  component: string;
}

interface ComponentPickerProps {
  type: NodeKind;
  components: PickableComponent[];
  onSelect: (comp: PickableComponent) => void;
}

const ALL_TAB = "__all__";

export function ComponentPicker({
  type,
  components,
  onSelect,
}: ComponentPickerProps) {
  const [search, setSearch] = useState("");
  const [tab, setTab] = useState<string>(ALL_TAB);

  const groups = useMemo(
    () => groupComponents(type, components),
    [type, components],
  );

  const normalizedQuery = search.trim().toLowerCase();

  const activeItems = useMemo(() => {
    const base =
      tab === ALL_TAB
        ? components
        : groups.find((g) => g.group === tab)?.items ?? [];
    if (!normalizedQuery) return base;
    return base.filter((comp) => {
      const desc = describeComponent(type, comp.id) ?? "";
      return (
        comp.id.toLowerCase().includes(normalizedQuery) ||
        comp.name.toLowerCase().includes(normalizedQuery) ||
        comp.component.toLowerCase().includes(normalizedQuery) ||
        desc.toLowerCase().includes(normalizedQuery)
      );
    });
  }, [tab, groups, components, normalizedQuery, type]);

  const renderGrid = (items: PickableComponent[]) => {
    if (items.length === 0) {
      return (
        <p className="py-8 text-center text-sm text-muted-foreground">
          No components match "{search}".
        </p>
      );
    }

    return (
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
        {items.map((comp) => renderCard(comp))}
      </div>
    );
  };

  const renderCard = (comp: PickableComponent) => {
    const Icon = getComponentIcon(type, comp.id);
    const desc = describeComponent(type, comp.id);
    const displayName = comp.name || comp.id;
    return (
      <button
        key={comp.id}
        type="button"
        onClick={() => onSelect(comp)}
        className="group flex flex-col items-start gap-1.5 rounded-md border border-border p-3 text-left transition hover:border-foreground/40 hover:bg-accent focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <div className="flex items-center gap-2">
          <Icon className="h-4 w-4 text-muted-foreground group-hover:text-foreground" />
          <span className="text-sm font-medium">{displayName}</span>
        </div>
        {desc && (
          <p className="line-clamp-2 text-[11px] leading-snug text-muted-foreground">
            {desc}
          </p>
        )}
      </button>
    );
  };

  return (
    <div className="mt-2 flex flex-col gap-3">
      <div className="relative">
        <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          autoFocus
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search components..."
          className="pl-8"
        />
      </div>

      <Tabs value={tab} onValueChange={setTab} className="flex flex-col">
        <TabsList className="h-auto w-full flex-wrap justify-start gap-1 bg-transparent p-0">
          <TabsTrigger
            value={ALL_TAB}
            className="border border-border data-[state=active]:border-foreground/30"
          >
            All
          </TabsTrigger>
          {groups.map(({ group }) => (
            <TabsTrigger
              key={group}
              value={group}
              className="border border-border data-[state=active]:border-foreground/30"
            >
              {group}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <div className="max-h-[55vh] overflow-y-auto pr-1">
        {renderGrid(activeItems)}
      </div>
    </div>
  );
}
