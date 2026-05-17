import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  name: z.string().min(1, "Required"),
  args_mapping: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { name: "", args_mapping: "" };

const component: FlowComponent<Config> = {
  id: "command",
  name: "Command",
  category: "processor",
  description:
    "Executes a command for each message, piping the message contents to stdin and replacing the message with stdout.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
