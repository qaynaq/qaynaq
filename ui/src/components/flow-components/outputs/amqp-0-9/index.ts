import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const exchangeDeclareSchema = z.object({
  enabled: z.boolean(),
  type: z.enum(["direct", "fanout", "topic", "x-custom"]),
  durable: z.boolean(),
});

const configSchema = z.object({
  urls: z.array(z.string()),
  exchange: z.string().min(1, "Required"),
  key: z.string(),
  type: z.string(),
  content_type: z.string(),
  persistent: z.boolean(),
  max_in_flight: z.number().int().min(1),
  exchange_declare: exchangeDeclareSchema,
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  urls: [],
  exchange: "",
  key: "",
  type: "",
  content_type: "application/octet-stream",
  persistent: false,
  max_in_flight: 64,
  exchange_declare: { enabled: false, type: "direct", durable: true },
};

const component: FlowComponent<Config> = {
  id: "amqp_0_9",
  name: "AMQP 0.9",
  category: "output",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
