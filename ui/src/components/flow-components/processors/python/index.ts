import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  script: z.string().min(1, "Required"),
  imports: z.array(z.string()),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { script: "", imports: [] };

const component: FlowComponent<Config> = {
  id: "python",
  name: "Python",
  category: "processor",
  description:
    "Executes a Python script for each message in a sandboxed WASM runtime. The message is exposed as 'this' and the result is read from 'root'.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
