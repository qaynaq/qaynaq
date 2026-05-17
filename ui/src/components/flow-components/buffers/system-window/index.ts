import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  timestamp_mapping: z.string(),
  size: z.string().min(1, "Required"),
  slide: z.string(),
  offset: z.string(),
  allowed_lateness: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  timestamp_mapping: "root = now()",
  size: "",
  slide: "",
  offset: "",
  allowed_lateness: "",
};

const component: FlowComponent<Config> = {
  id: "system_window",
  name: "System Window",
  category: "buffer",
  description:
    "Groups messages into time-based windows using system clocks. Supports tumbling, sliding, and hopping windows with configurable lateness handling.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
