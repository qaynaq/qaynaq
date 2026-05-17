import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import { tlsSchema, defaultTls } from "../../shared/tls";

const saslSchema = z.object({
  enabled: z.boolean(),
  mechanism: z.enum(["none", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"]),
  username: z.string(),
  password: z.string(),
});

const metaSchema = z.object({
  include_prefixes: z.array(z.string()),
  include_patterns: z.array(z.string()),
});

const batchingSchema = z.object({
  count: z.number().int().min(0),
  byte_size: z.number().int().min(0),
  period: z.string(),
  jitter: z.number().min(0),
  check: z.string(),
  processors: z.array(z.unknown()),
});

const configSchema = z.object({
  addresses: z.array(z.string()),
  topic: z.string(),
  key: z.string(),
  client_id: z.string(),
  max_in_flight: z.number().int().min(1),
  ack_replicas: z.boolean(),
  compression: z.enum(["none", "gzip", "snappy", "lz4", "zstd"]),
  max_message_bytes: z.number().int().min(1),
  target_version: z.string(),
  timeout: z.string(),
  metadata: metaSchema,
  tls: tlsSchema,
  sasl: saslSchema,
  batching: batchingSchema,
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  addresses: [],
  topic: "",
  key: "",
  client_id: "",
  max_in_flight: 1000,
  ack_replicas: false,
  compression: "none",
  max_message_bytes: 1000000,
  target_version: "2.0.0",
  timeout: "5s",
  metadata: { include_prefixes: [], include_patterns: [] },
  tls: defaultTls,
  sasl: { enabled: false, mechanism: "none", username: "", password: "" },
  batching: {
    count: 0,
    byte_size: 0,
    period: "",
    jitter: 0,
    check: "",
    processors: [],
  },
};

const component: FlowComponent<Config> = {
  id: "kafka",
  name: "Kafka",
  category: "output",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
