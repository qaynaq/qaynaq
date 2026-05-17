import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const batchPolicySchema = z.object({
  enabled: z.boolean(),
  count: z.number().int().min(0),
  byte_size: z.number().int().min(0),
  period: z.string(),
  jitter: z.number().min(0),
  check: z.string(),
  processors: z.string(),
});

const configSchema = z.object({
  limit: z.number().int().min(1),
  spillover: z.boolean(),
  batch_policy: batchPolicySchema,
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  limit: 524288000,
  spillover: false,
  batch_policy: {
    enabled: false,
    count: 0,
    byte_size: 0,
    period: "",
    jitter: 0,
    check: "",
    processors: "",
  },
};

const component: FlowComponent<Config> = {
  id: "memory",
  name: "Memory",
  category: "buffer",
  description:
    "Stores consumed messages in memory and acknowledges them at the input level. During graceful shutdown, the buffer flushes remaining messages. Delivery is not guaranteed in case of crashes.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
