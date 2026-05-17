import { lazy } from "react";
import { z } from "zod";
import * as yaml from "js-yaml";
import type { FlowComponent } from "../../types";
import {
  rawToListItem,
  listItemToRaw,
  type ListItem,
} from "../../utils/list-items";

const configSchema = z.object({
  processors: z.array(z.unknown()),
});
type Config = { processors: ListItem[] };

const defaultConfig: Config = { processors: [] };

const component: FlowComponent<Config> = {
  id: "catch",
  name: "Catch",
  category: "processor",
  description:
    "Applies a list of child processors only when a previous processing step has failed.",
  configSchema: configSchema as unknown as z.ZodType<Config>,
  defaultConfig,
  parse: (s) => {
    if (!s?.trim()) return { processors: [] };
    let raw: unknown;
    try {
      raw = yaml.load(s);
    } catch {
      return { processors: [] };
    }
    if (!Array.isArray(raw)) return { processors: [] };
    return {
      processors: raw
        .map((r) => rawToListItem("processor", r))
        .filter((x): x is ListItem => x !== null),
    };
  },
  serialize: (c) => {
    const arr = c.processors.map((item) => listItemToRaw("processor", item));
    return yaml.dump(arr, { lineWidth: -1, noRefs: true });
  },
  toListItem: (c) => c.processors.map((item) => listItemToRaw("processor", item)),
  fromListItem: (raw) => {
    if (!Array.isArray(raw)) return { processors: [] };
    return {
      processors: raw
        .map((r) => rawToListItem("processor", r))
        .filter((x): x is ListItem => x !== null),
    };
  },
  Editor: lazy(() => import("./editor")),
};

export default component;
