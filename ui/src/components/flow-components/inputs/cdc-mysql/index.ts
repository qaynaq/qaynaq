import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  host: z.string().min(1, "Required"),
  port: z.number().int().min(1),
  user: z.string().min(1, "Required"),
  password: z.string().min(1, "Required"),
  server_id: z.string(),
  position_cache: z.string().min(1, "Required"),
  position_cache_key: z.string().min(1, "Required"),
  position_mode: z.enum(["gtid", "file"]),
  flavor: z.enum(["mysql", "mariadb"]),
  cache_save_interval: z.string(),
  include_tables: z.array(z.string()),
  exclude_tables: z.array(z.string()),
  use_schema_cache: z.boolean(),
  schema_cache_ttl: z.string(),
  max_batch_size: z.number().int().min(1),
  max_pending_checkpoints: z.number().int().min(1),
  retry_initial_interval: z.string(),
  retry_max_interval: z.string(),
  retry_multiplier: z.number().min(1),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  host: "",
  port: 3306,
  user: "",
  password: "",
  server_id: "1000",
  position_cache: "",
  position_cache_key: "",
  position_mode: "gtid",
  flavor: "mysql",
  cache_save_interval: "30s",
  include_tables: [],
  exclude_tables: [],
  use_schema_cache: true,
  schema_cache_ttl: "5m",
  max_batch_size: 1000,
  max_pending_checkpoints: 100,
  retry_initial_interval: "1s",
  retry_max_interval: "30s",
  retry_multiplier: 2.0,
};

const component: FlowComponent<Config> = {
  id: "cdc_mysql",
  name: "MySQL (MariaDB) CDC",
  category: "input",
  description:
    "Captures change events from a MySQL/MariaDB binlog stream. Experimental.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
