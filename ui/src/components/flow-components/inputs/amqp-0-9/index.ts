import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const queueDeclareSchema = z.object({
  enabled: z.boolean(),
  durable: z.boolean(),
  auto_delete: z.boolean(),
});

const configSchema = z.object({
  urls: z.array(z.string()),
  queue: z.string().min(1, "Required"),
  consumer_tag: z.string(),
  auto_ack: z.boolean(),
  prefetch_count: z.number().int().min(0),
  nack_reject_patterns: z.array(z.string()),
  queue_declare: queueDeclareSchema,
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  urls: [],
  queue: "",
  consumer_tag: "",
  auto_ack: false,
  prefetch_count: 10,
  nack_reject_patterns: [],
  queue_declare: { enabled: false, durable: true, auto_delete: false },
};

const component: FlowComponent<Config> = {
  id: "amqp_0_9",
  name: "AMQP 0.9",
  category: "input",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
