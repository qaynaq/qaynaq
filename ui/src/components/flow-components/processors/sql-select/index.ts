import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";
import { SQL_DRIVERS } from "../sql-raw";

const configSchema = z.object({
  driver: z.enum(SQL_DRIVERS),
  dsn: z.string().min(1, "Required"),
  table: z.string().min(1, "Required"),
  columns: z.array(z.string()),
  where: z.string(),
  args_mapping: z.string(),
  prefix: z.string(),
  suffix: z.string(),
  init_statement: z.string(),
  conn_max_idle_time: z.string(),
  conn_max_life_time: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  driver: "mysql",
  dsn: "",
  table: "",
  columns: [],
  where: "",
  args_mapping: "",
  prefix: "",
  suffix: "",
  init_statement: "",
  conn_max_idle_time: "",
  conn_max_life_time: "",
};

const component: FlowComponent<Config> = {
  id: "sql_select",
  name: "SQL Select",
  category: "processor",
  description:
    "Runs a select query against a database and replaces the message with the rows returned.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
