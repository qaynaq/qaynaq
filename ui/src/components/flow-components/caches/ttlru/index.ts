import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  cap: z.number().int().positive(),
  default_ttl: z.string(),
  init_values: z.record(z.string(), z.string()),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  cap: 1024,
  default_ttl: "5m",
  init_values: {},
};

const component: FlowComponent<Config> = {
  id: "ttlru",
  name: "TTLRU",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
