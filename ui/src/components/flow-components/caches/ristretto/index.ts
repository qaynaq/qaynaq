import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  default_ttl: z.string(),
  max_cost: z.number().int().positive(),
  num_counters: z.number().int().positive(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  default_ttl: "5m",
  max_cost: 1073741824,
  num_counters: 10000000,
};

const component: FlowComponent<Config> = {
  id: "ristretto",
  name: "Ristretto",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
