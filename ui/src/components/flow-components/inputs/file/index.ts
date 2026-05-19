import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  paths: z.array(z.string()),
  scanner: z.record(z.unknown()),
  delete_on_finish: z.boolean(),
  auto_replay_nacks: z.boolean(),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  paths: [],
  scanner: { lines: {} },
  delete_on_finish: false,
  auto_replay_nacks: true,
};

const component: FlowComponent<Config> = {
  id: "file",
  name: "File",
  category: "input",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
