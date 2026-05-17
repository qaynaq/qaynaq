import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  mapping: z.string().min(1, "Required"),
  interval: z.string(),
  count: z.number().int().min(0),
  batch_size: z.number().int().min(1),
  auto_replay_nacks: z.boolean(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  mapping: "",
  interval: "1s",
  count: 0,
  batch_size: 1,
  auto_replay_nacks: true,
};

const component: FlowComponent<Config> = {
  id: "generate",
  name: "Generate",
  category: "input",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
