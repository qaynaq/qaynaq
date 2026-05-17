import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  path: z.string().min(1, "Required"),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { path: "" };

const component: FlowComponent<Config> = {
  id: "sqlite",
  name: "SQLite",
  category: "buffer",
  description:
    "Persists messages to a SQLite database file. Messages are not acknowledged until written to disk, and are not removed until successfully delivered. Preserves at-least-once delivery across restarts.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
