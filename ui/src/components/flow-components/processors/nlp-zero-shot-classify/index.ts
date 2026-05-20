import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import {
  hugotModelDefaults,
  hugotModelSchema,
} from "../../shared/hugot-model-fields";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const configSchema = hugotModelSchema.extend({
  labels: z.array(z.string()).min(1, "At least one label is required"),
  multi_label: z.boolean(),
  hypothesis_template: z
    .string()
    .min(1, "Required")
    .refine((s) => s.includes("{}"), {
      message: "Must contain {} where the label will be inserted.",
    }),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  ...hugotModelDefaults,
  labels: [],
  multi_label: false,
  hypothesis_template: "This example is {}.",
};

const component: FlowComponent<Config> = {
  id: "nlp_zero_shot_classify",
  name: "NLP Zero-Shot Classify",
  category: "processor",
  description:
    "Classify text into arbitrary labels at runtime, without training, using an NLI model.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
