import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const SQL_DRIVERS = [
  "mysql",
  "postgres",
  "clickhouse",
  "mssql",
  "sqlite",
  "oracle",
  "snowflake",
  "trino",
  "gocosmos",
  "spanner",
] as const;

const configSchema = z.object({
  driver: z.enum(SQL_DRIVERS),
  dsn: z.string().min(1, "Required"),
  query: z.string().min(1, "Required"),
  unsafe_dynamic_query: z.boolean(),
  args_mapping: z.string(),
  exec_only: z.boolean(),
  init_statement: z.string(),
  conn_max_idle_time: z.string(),
  conn_max_life_time: z.string(),
  conn_max_idle: z.number().int().min(0),
  conn_max_open: z.number().int().min(0),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  driver: "mysql",
  dsn: "",
  query: "",
  unsafe_dynamic_query: false,
  args_mapping: "",
  exec_only: false,
  init_statement: "",
  conn_max_idle_time: "",
  conn_max_life_time: "",
  conn_max_idle: 2,
  conn_max_open: 0,
};

const component: FlowComponent<Config> = {
  id: "sql_raw",
  name: "SQL Raw",
  category: "processor",
  description:
    "Runs an arbitrary SQL query against a database and replaces the message with the result.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
