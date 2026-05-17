import { lazy } from "react";
import { z } from "zod";
import * as yaml from "js-yaml";
import type { FlowComponent } from "../../types";
import {
  rawToListItem,
  listItemToRaw,
  type ListItem,
} from "../../utils/list-items";

interface Config {
  request_map: string;
  processors: ListItem[];
  result_map: string;
}

const configSchema = z.object({
  request_map: z.string(),
  processors: z.array(z.unknown()),
  result_map: z.string(),
});

const defaultConfig: Config = {
  request_map: "",
  processors: [],
  result_map: "",
};

const component: FlowComponent<Config> = {
  id: "branch",
  name: "Branch",
  category: "processor",
  description:
    "Executes a list of processors on a copy of the message, optionally transforming the input and mapping the result back.",
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
      request_map: typeof r.request_map === "string" ? r.request_map : "",
      processors: Array.isArray(r.processors)
        ? r.processors
            .map((p) => rawToListItem("processor", p))
            .filter((x): x is ListItem => x !== null)
        : [],
      result_map: typeof r.result_map === "string" ? r.result_map : "",
    };
  },
  serialize: (c) =>
    yaml.dump(
      {
        request_map: c.request_map,
        processors: c.processors.map((item) => listItemToRaw("processor", item)),
        result_map: c.result_map,
      },
      { lineWidth: -1, noRefs: true },
    ),
  Editor: lazy(() => import("./editor")),
};

export default component;
