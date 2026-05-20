import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import {
  hugotModelDefaults,
  hugotModelSchema,
} from "../../shared/hugot-model-fields";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = hugotModelSchema.extend({
  normalization: z.boolean(),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  ...hugotModelDefaults,
  normalization: false,
};

const component: FlowComponent<Config> = {
  id: "nlp_extract_features",
  name: "NLP Extract Features",
  category: "processor",
  description:
    "Run an ONNX feature extraction model and replace the message with its embedding.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
