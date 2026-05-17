import { lazy } from "react";
import { z } from "zod";
import * as yaml from "js-yaml";
import type { FlowComponent } from "../../types";
import {
  rawToListItem,
  listItemToRaw,
  type ListItem,
} from "../../utils/list-items";
import type { OutputCase } from "../../shared/output-cases-field";

interface Config {
  retry_until_success: boolean;
  strict_mode: boolean;
  cases: OutputCase[];
}

const configSchema = z.object({
  retry_until_success: z.boolean(),
  strict_mode: z.boolean(),
  cases: z.array(z.unknown()),
});

const defaultConfig: Config = {
  retry_until_success: false,
  strict_mode: false,
  cases: [],
};

function rawToCase(raw: unknown): OutputCase | null {
  if (typeof raw !== "object" || raw === null) return null;
  const r = raw as Record<string, unknown>;
  const output =
    r.output && typeof r.output === "object"
      ? rawToListItem("output", r.output)
      : null;
  if (!output) return null;
  return {
    check: typeof r.check === "string" ? r.check : "",
    output,
    continue: !!r.continue,
  };
}

function caseToRaw(c: OutputCase): Record<string, unknown> {
  const out: Record<string, unknown> = {
    check: c.check ?? "",
    output: listItemToRaw("output", c.output),
  };
  if (c.continue) out.continue = true;
  return out;
}

const component: FlowComponent<Config> = {
  id: "switch",
  name: "Switch",
  category: "output",
  description: "Route messages to different outputs based on their contents.",
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
      retry_until_success: !!r.retry_until_success,
      strict_mode: !!r.strict_mode,
      cases: Array.isArray(r.cases)
        ? r.cases
            .map(rawToCase)
            .filter((c): c is OutputCase => c !== null)
        : [],
    };
  },
  serialize: (c) =>
    yaml.dump(
      {
        retry_until_success: c.retry_until_success,
        strict_mode: c.strict_mode,
        cases: c.cases.map(caseToRaw),
      },
      { lineWidth: -1, noRefs: true },
    ),
  Editor: lazy(() => import("./editor")),
};

export default component;
