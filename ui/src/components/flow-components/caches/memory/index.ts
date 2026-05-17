import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  default_ttl: z.string(),
  compaction_interval: z.string(),
  init_values: z.record(z.string(), z.string()),
  shards: z.number().int().min(1),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  default_ttl: "5m",
  compaction_interval: "60s",
  init_values: {},
  shards: 1,
};

const component: FlowComponent<Config> = {
  id: "memory",
  name: "Memory",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
