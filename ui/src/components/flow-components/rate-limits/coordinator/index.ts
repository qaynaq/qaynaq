import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const INTERVAL_OPTIONS = [
  "1s",
  "5s",
  "10s",
  "30s",
  "1m",
  "5m",
  "10m",
  "30m",
  "1h",
  "2h",
  "6h",
  "12h",
  "24h",
] as const;

const configSchema = z.object({
  count: z.number().int().min(1),
  interval: z.enum(INTERVAL_OPTIONS),
  burst: z.number().int().min(0),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { count: 10, interval: "1s", burst: 0 };

const component: FlowComponent<Config> = {
  id: "coordinator",
  name: "Coordinator",
  category: "rate_limit",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
