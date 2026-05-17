import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import { retriesSchema, defaultRetries } from "../../shared/retries";

const configSchema = z.object({
  addresses: z.array(z.string()),
  prefix: z.string(),
  default_ttl: z.string(),
  retries: retriesSchema,
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  addresses: [],
  prefix: "",
  default_ttl: "300s",
  retries: { ...defaultRetries, initial_interval: "1s", max_interval: "5s", max_elapsed_time: "30s" },
};

const component: FlowComponent<Config> = {
  id: "memcached",
  name: "Memcached",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
