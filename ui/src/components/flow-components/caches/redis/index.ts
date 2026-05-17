import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import { tlsSchema, defaultTls } from "../../shared/tls";
import { retriesSchema, defaultRetries } from "../../shared/retries";

const configSchema = z.object({
  url: z.string().min(1, "Required"),
  kind: z.enum(["simple", "cluster", "failover"]),
  master: z.string(),
  tls: tlsSchema,
  prefix: z.string(),
  default_ttl: z.string(),
  retries: retriesSchema,
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  url: "",
  kind: "simple",
  master: "",
  tls: defaultTls,
  prefix: "",
  default_ttl: "",
  retries: defaultRetries,
};

const component: FlowComponent<Config> = {
  id: "redis",
  name: "Redis",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
