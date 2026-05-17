import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  cap: z.number().int().positive(),
  init_values: z.record(z.string(), z.string()),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { cap: 1024, init_values: {} };

const component: FlowComponent<Config> = {
  id: "lru",
  name: "LRU",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
