import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const saslSchema = z.object({
  mechanism: z.enum(["none", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"]),
  user: z.string(),
  password: z.string(),
});

const batchingSchema = z.object({
  count: z.number().int().min(0),
  byte_size: z.number().int().min(0),
  period: z.string(),
  check: z.string(),
});

const configSchema = z.object({
  addresses: z.array(z.string()),
  topics: z.array(z.string()),
  consumer_group: z.string(),
  client_id: z.string(),
  rack_id: z.string(),
  start_from_oldest: z.boolean(),
  checkpoint_limit: z.number().int().min(1),
  commit_period: z.string(),
  max_processing_period: z.string(),
  sasl: saslSchema,
  target_version: z.string(),
  batching: batchingSchema,
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  addresses: [],
  topics: [],
  consumer_group: "",
  client_id: "",
  rack_id: "",
  start_from_oldest: false,
  checkpoint_limit: 1000,
  commit_period: "1s",
  max_processing_period: "100ms",
  sasl: { mechanism: "none", user: "", password: "" },
  target_version: "2.0.0",
  batching: { count: 0, byte_size: 0, period: "", check: "" },
};

const component: FlowComponent<Config> = {
  id: "kafka",
  name: "Kafka",
  category: "input",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
