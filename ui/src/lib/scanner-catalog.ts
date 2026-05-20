import {
  AlignLeft,
  Archive,
  Boxes,
  Braces,
  Code2,
  FileSearch,
  FileText,
  Hash,
  Layers,
  Package,
  ScanLine,
  Sparkles,
  StopCircle,
  type LucideIcon,
} from "lucide-react";

export type ScannerGroup =
  | "RAG"
  | "Text"
  | "Structured"
  | "Binary"
  | "Composite";

export interface ScannerCatalogEntry {
  group: ScannerGroup;
  icon: LucideIcon;
  description: string;
}

const catalog: Record<string, ScannerCatalogEntry> = {
  rag_chunker: {
    group: "RAG",
    icon: Sparkles,
    description:
      "Split text into overlapping chunks for RAG indexing, with recursive, token, or markdown strategies.",
  },
  lines: {
    group: "Text",
    icon: AlignLeft,
    description: "Split the stream into one message per line.",
  },
  csv: {
    group: "Text",
    icon: FileText,
    description: "Parse comma-separated values row by row.",
  },
  re_match: {
    group: "Text",
    icon: FileSearch,
    description: "Split the stream wherever a regular expression matches.",
  },
  json_documents: {
    group: "Structured",
    icon: Braces,
    description: "Consume a stream of one or more JSON documents.",
  },
  xml_documents: {
    group: "Structured",
    icon: Code2,
    description: "Consume a stream of XML documents, optionally converted to JSON.",
  },
  avro: {
    group: "Structured",
    icon: Boxes,
    description: "Consume Avro OCF datum.",
  },
  chunker: {
    group: "Binary",
    icon: Hash,
    description: "Split the stream into fixed-size byte chunks.",
  },
  tar: {
    group: "Binary",
    icon: Archive,
    description: "Read each file in a tar archive as a separate message.",
  },
  to_the_end: {
    group: "Binary",
    icon: StopCircle,
    description: "Read the entire stream as a single message.",
  },
  skip_bom: {
    group: "Composite",
    icon: ScanLine,
    description: "Strip a leading byte order mark before delegating to a child scanner.",
  },
  decompress: {
    group: "Composite",
    icon: Layers,
    description: "Decompress the byte stream before delegating to a child scanner.",
  },
};

export function getScannerCatalogEntry(
  id: string,
): ScannerCatalogEntry | undefined {
  return catalog[id];
}

export function getScannerIcon(id: string | undefined): LucideIcon {
  if (id && catalog[id]) return catalog[id].icon;
  return Package;
}

export interface ScannerCatalogGroup<T> {
  group: ScannerGroup;
  items: T[];
}

const groupOrder: ScannerGroup[] = [
  "RAG",
  "Text",
  "Structured",
  "Binary",
  "Composite",
];

export function groupScanners<T extends { id: string }>(
  scanners: T[],
): ScannerCatalogGroup<T>[] {
  const buckets = new Map<ScannerGroup, T[]>();
  for (const s of scanners) {
    const group = catalog[s.id]?.group ?? "Binary";
    const arr = buckets.get(group);
    if (arr) arr.push(s);
    else buckets.set(group, [s]);
  }
  return groupOrder
    .filter((g) => buckets.has(g))
    .map((group) => ({
      group,
      items: (buckets.get(group) ?? []).sort((a, b) => a.id.localeCompare(b.id)),
    }));
}

export function describeScanner(id: string): string | undefined {
  return catalog[id]?.description;
}
