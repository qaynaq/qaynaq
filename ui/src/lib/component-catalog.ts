import {
  AlertTriangle,
  ArrowRightLeft,
  Braces,
  Calendar,
  Database,
  FileText,
  Globe,
  HardDrive,
  Key,
  MessageSquare,
  Network,
  Package,
  Radio,
  Reply,
  Search,
  Server,
  Sheet,
  ShoppingBag,
  Shuffle,
  Sparkles,
  Split,
  Terminal,
  Wrench,
  Zap,
  Download,
  Workflow,
  Upload,
  PlusCircle,
  type LucideIcon,
} from "lucide-react";

export interface ComponentCatalogEntry {
  group: string;
  icon: LucideIcon;
  description: string;
}

type NodeKind = "input" | "processor" | "output";

const catalog: Record<NodeKind, Record<string, ComponentCatalogEntry>> = {
  input: {
    generate: {
      group: "Testing",
      icon: Sparkles,
      description: "Generate synthetic messages on an interval for testing.",
    },
    http_client: {
      group: "HTTP",
      icon: Globe,
      description: "Poll or fetch messages from an HTTP endpoint.",
    },
    http_server: {
      group: "HTTP",
      icon: Server,
      description: "Receive messages via an HTTP endpoint exposed by Qaynaq.",
    },
    mcp_tool: {
      group: "AI",
      icon: Wrench,
      description: "Expose this flow as an MCP tool for AI agents.",
    },
    kafka: {
      group: "Messaging",
      icon: Radio,
      description: "Consume messages from Kafka topics.",
    },
    amqp_0_9: {
      group: "Messaging",
      icon: MessageSquare,
      description: "Consume messages from an AMQP 0.9 broker (RabbitMQ).",
    },
    broker: {
      group: "Messaging",
      icon: Network,
      description: "Combine multiple inputs into one stream.",
    },
    cdc_mysql: {
      group: "Database",
      icon: Database,
      description: "Stream MySQL changes via change data capture.",
    },
    shopify: {
      group: "E-commerce",
      icon: ShoppingBag,
      description: "Pull orders, products, and other data from Shopify.",
    },
  },
  processor: {
    ai_gateway: {
      group: "AI",
      icon: Sparkles,
      description: "Route requests to LLM providers with retries and caching.",
    },
    google_calendar: {
      group: "Google",
      icon: Calendar,
      description: "Read or write events in Google Calendar.",
    },
    google_drive: {
      group: "Google",
      icon: HardDrive,
      description: "Read or write files in Google Drive.",
    },
    google_sheets: {
      group: "Google",
      icon: Sheet,
      description: "Read or write rows in a Google Sheet.",
    },
    branch: {
      group: "Flow control",
      icon: Split,
      description: "Run a sub-pipeline on a subset of fields and merge the result.",
    },
    switch: {
      group: "Flow control",
      icon: Shuffle,
      description: "Route messages to different processors based on conditions.",
    },
    catch: {
      group: "Flow control",
      icon: AlertTriangle,
      description: "Handle errors from previous processors.",
    },
    mapping: {
      group: "Transform",
      icon: ArrowRightLeft,
      description: "Transform messages with a Bloblang mapping.",
    },
    json_schema: {
      group: "Transform",
      icon: Braces,
      description: "Validate messages against a JSON Schema.",
    },
    schema_registry_decode: {
      group: "Transform",
      icon: Key,
      description: "Decode messages using a schema registry.",
    },
    sync_response: {
      group: "Response",
      icon: Reply,
      description: "Send the current message back as the HTTP sync response.",
    },
    http: {
      group: "HTTP",
      icon: Globe,
      description: "Call an HTTP endpoint and use the response.",
    },
    sql_raw: {
      group: "Database",
      icon: Terminal,
      description: "Execute a raw SQL query.",
    },
    sql_select: {
      group: "Database",
      icon: Search,
      description: "Run a SELECT query and enrich messages with the result.",
    },
    sql_insert: {
      group: "Database",
      icon: PlusCircle,
      description: "Insert messages into a SQL table.",
    },
  },
  output: {
    http_client: {
      group: "HTTP",
      icon: Globe,
      description: "Send messages to an HTTP endpoint.",
    },
    kafka: {
      group: "Messaging",
      icon: Radio,
      description: "Publish messages to Kafka topics.",
    },
    amqp_0_9: {
      group: "Messaging",
      icon: MessageSquare,
      description: "Publish messages to an AMQP 0.9 broker.",
    },
    broker: {
      group: "Messaging",
      icon: Network,
      description: "Fan out messages to multiple outputs.",
    },
    switch: {
      group: "Flow control",
      icon: Shuffle,
      description: "Route messages to different outputs based on conditions.",
    },
    sync_response: {
      group: "Response",
      icon: Reply,
      description: "Return the message as the HTTP sync response.",
    },
    sql_insert: {
      group: "Database",
      icon: PlusCircle,
      description: "Insert messages into a SQL table.",
    },
  },
};

const defaultIcon: Record<NodeKind, LucideIcon> = {
  input: Download,
  processor: Workflow,
  output: Upload,
};

export function getComponentIcon(
  type: NodeKind | undefined,
  componentId: string | undefined,
): LucideIcon {
  if (type && componentId) {
    const entry = catalog[type]?.[componentId];
    if (entry) return entry.icon;
  }
  return type ? defaultIcon[type] : Package;
}

export function getComponentCatalogEntry(
  type: NodeKind,
  componentId: string,
): ComponentCatalogEntry | undefined {
  return catalog[type]?.[componentId];
}

export interface CatalogGroup<T> {
  group: string;
  items: T[];
}

export function groupComponents<T extends { id: string }>(
  type: NodeKind,
  components: T[],
): CatalogGroup<T>[] {
  const buckets = new Map<string, T[]>();
  for (const comp of components) {
    const group = catalog[type]?.[comp.id]?.group || "Other";
    const arr = buckets.get(group);
    if (arr) arr.push(comp);
    else buckets.set(group, [comp]);
  }
  const order = [
    "HTTP",
    "Messaging",
    "Database",
    "AI",
    "Google",
    "E-commerce",
    "Transform",
    "Flow control",
    "Response",
    "Testing",
    "Other",
  ];
  return [...buckets.entries()]
    .sort(([a], [b]) => {
      const ia = order.indexOf(a);
      const ib = order.indexOf(b);
      if (ia === -1 && ib === -1) return a.localeCompare(b);
      if (ia === -1) return 1;
      if (ib === -1) return -1;
      return ia - ib;
    })
    .map(([group, items]) => ({
      group,
      items: items.sort((x, y) => x.id.localeCompare(y.id)),
    }));
}

export function describeComponent(
  type: NodeKind,
  componentId: string,
): string | undefined {
  return catalog[type]?.[componentId]?.description;
}
