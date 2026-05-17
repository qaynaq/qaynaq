import { lazy } from "react";
import { z } from "zod";
import * as yaml from "js-yaml";
import type { FlowComponent } from "../../types";
import {
  rawToListItem,
  listItemToRaw,
  type ListItem,
} from "../../utils/list-items";

interface Batching {
  count: number;
  byte_size: number;
  period: string;
  jitter: number;
  check: string;
  processors: ListItem[];
}

interface Config {
  copies: number;
  inputs: ListItem[];
  batching: Batching;
}

const configSchema = z.object({
  copies: z.number().int().min(1),
  inputs: z.array(z.unknown()),
  batching: z.object({
    count: z.number().int().min(0),
    byte_size: z.number().int().min(0),
    period: z.string(),
    jitter: z.number().min(0),
    check: z.string(),
    processors: z.array(z.unknown()),
  }),
});

const defaultConfig: Config = {
  copies: 1,
  inputs: [],
  batching: {
    count: 0,
    byte_size: 0,
    period: "",
    jitter: 0,
    check: "",
    processors: [],
  },
};

function rawBatching(raw: unknown): Batching {
  if (typeof raw !== "object" || raw === null) return defaultConfig.batching;
  const r = raw as Record<string, unknown>;
  return {
    count: typeof r.count === "number" ? r.count : 0,
    byte_size: typeof r.byte_size === "number" ? r.byte_size : 0,
    period: typeof r.period === "string" ? r.period : "",
    jitter: typeof r.jitter === "number" ? r.jitter : 0,
    check: typeof r.check === "string" ? r.check : "",
    processors: Array.isArray(r.processors)
      ? r.processors
          .map((p) => rawToListItem("processor", p))
          .filter((x): x is ListItem => x !== null)
      : [],
  };
}

function batchingToRaw(b: Batching): Record<string, unknown> {
  return {
    count: b.count,
    byte_size: b.byte_size,
    period: b.period,
    jitter: b.jitter,
    check: b.check,
    processors: b.processors.map((item) => listItemToRaw("processor", item)),
  };
}

const component: FlowComponent<Config> = {
  id: "broker",
  name: "Broker",
  category: "input",
  description:
    "Combine multiple inputs into a single flow using a range of broker patterns.",
  configSchema: configSchema as unknown as z.ZodType<Config>,
  defaultConfig,
  parse: (s) => {
    if (!s?.trim()) return structuredClone(defaultConfig);
    let raw: unknown;
    try {
      raw = yaml.load(s);
    } catch {
      return structuredClone(defaultConfig);
    }
    if (typeof raw !== "object" || raw === null) {
      return structuredClone(defaultConfig);
    }
    const r = raw as Record<string, unknown>;
    return {
      copies: typeof r.copies === "number" ? r.copies : 1,
      inputs: Array.isArray(r.inputs)
        ? r.inputs
            .map((i) => rawToListItem("input", i))
            .filter((x): x is ListItem => x !== null)
        : [],
      batching: rawBatching(r.batching),
    };
  },
  serialize: (c) =>
    yaml.dump(
      {
        copies: c.copies,
        inputs: c.inputs.map((item) => listItemToRaw("input", item)),
        batching: batchingToRaw(c.batching),
      },
      { lineWidth: -1, noRefs: true },
    ),
  Editor: lazy(() => import("./editor")),
};

export default component;
