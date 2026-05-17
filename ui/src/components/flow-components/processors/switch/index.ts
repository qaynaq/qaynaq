import { lazy } from "react";
import { z } from "zod";
import * as yaml from "js-yaml";
import type { FlowComponent } from "../../types";
import {
  rawToListItem,
  listItemToRaw,
  type ListItem,
} from "../../utils/list-items";
import type { ProcessorCase } from "../../shared/processor-cases-field";

const configSchema = z.object({
  cases: z.array(z.unknown()),
});
type Config = { cases: ProcessorCase[] };

const defaultConfig: Config = { cases: [] };

function rawToCase(raw: unknown): ProcessorCase {
  if (typeof raw !== "object" || raw === null) {
    return { check: "", processors: [], fallthrough: false };
  }
  const r = raw as Record<string, unknown>;
  const processors = Array.isArray(r.processors)
    ? r.processors
        .map((p) => rawToListItem("processor", p))
        .filter((x): x is ListItem => x !== null)
    : [];
  return {
    check: typeof r.check === "string" ? r.check : "",
    processors,
    fallthrough: !!r.fallthrough,
  };
}

function caseToRaw(c: ProcessorCase): Record<string, unknown> {
  const out: Record<string, unknown> = {
    check: c.check ?? "",
    processors: c.processors.map((item) => listItemToRaw("processor", item)),
  };
  if (c.fallthrough) out.fallthrough = true;
  return out;
}

const component: FlowComponent<Config> = {
  id: "switch",
  name: "Switch",
  category: "processor",
  description: "Conditionally processes messages based on their contents.",
  configSchema: configSchema as unknown as z.ZodType<Config>,
  defaultConfig,
  parse: (s) => {
    if (!s?.trim()) return { cases: [] };
    let raw: unknown;
    try {
      raw = yaml.load(s);
    } catch {
      return { cases: [] };
    }
    if (!Array.isArray(raw)) return { cases: [] };
    return { cases: raw.map(rawToCase) };
  },
  serialize: (c) =>
    yaml.dump(c.cases.map(caseToRaw), { lineWidth: -1, noRefs: true }),
  toListItem: (c) => c.cases.map(caseToRaw),
  fromListItem: (raw) => {
    if (!Array.isArray(raw)) return { cases: [] };
    return { cases: raw.map(rawToCase) };
  },
  Editor: lazy(() => import("./editor")),
};

export default component;
