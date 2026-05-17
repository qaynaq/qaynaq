import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({
  schema_path: z.string(),
  schema: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = { schema_path: "", schema: "" };

const component: FlowComponent<Config> = {
  id: "json_schema",
  name: "JSON Schema",
  category: "processor",
  description:
    "Validates messages against a JSON schema. Use schema_path to point at a stored file or schema for an inline value.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
