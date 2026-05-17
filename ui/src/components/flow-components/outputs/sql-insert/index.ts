import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import { SQL_DRIVERS } from "../../processors/sql-raw";

const batchingSchema = z.object({
  count: z.number().int().min(0),
  byte_size: z.number().int().min(0),
  period: z.string(),
  jitter: z.number().min(0),
  check: z.string(),
});

const configSchema = z.object({
  driver: z.enum(SQL_DRIVERS),
  dsn: z.string().min(1, "Required"),
  table: z.string().min(1, "Required"),
  columns: z.array(z.string()),
  args_mapping: z.string().min(1, "Required"),
  suffix: z.string(),
  max_in_flight: z.number().int().min(1),
  batching: batchingSchema,
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  driver: "mysql",
  dsn: "",
  table: "",
  columns: [],
  args_mapping: "",
  suffix: "",
  max_in_flight: 64,
  batching: { count: 0, byte_size: 0, period: "", jitter: 0, check: "" },
};

const component: FlowComponent<Config> = {
  id: "sql_insert",
  name: "SQL Insert",
  category: "output",
  description: "Inserts a row into an SQL database for each message.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() =>
    import("../../processors/sql-insert/editor").then((m) => ({
      default: m.default,
    })),
  ),
};

export default component;
