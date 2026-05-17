import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  path: z.string(),
  allowed_verbs: z.array(z.string()),
  timeout: z.string(),
  sync_response: z.object({
    status: z.string(),
    headers: z.record(z.string(), z.string()),
  }),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  path: "/post",
  allowed_verbs: ["POST"],
  timeout: "5s",
  sync_response: { status: "200", headers: {} },
};

const component: FlowComponent<Config> = {
  id: "http_server",
  name: "HTTP Server",
  category: "input",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
