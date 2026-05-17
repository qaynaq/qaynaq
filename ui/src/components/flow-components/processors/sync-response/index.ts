import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = z.object({}).passthrough();
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {};

const component: FlowComponent<Config> = {
  id: "sync_response",
  name: "Sync Response",
  category: "processor",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() =>
    import("@/components/form-primitives/empty-editor").then((m) => ({
      default: m.EmptyEditor,
    })),
  ),
};

export default component;
