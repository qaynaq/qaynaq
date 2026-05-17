import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  directory: z.string().min(1, "Required"),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { directory: "" };

const component: FlowComponent<Config> = {
  id: "file",
  name: "File",
  category: "cache",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
