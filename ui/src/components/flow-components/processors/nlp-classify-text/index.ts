import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import {
  hugotModelDefaults,
  hugotModelSchema,
} from "../../shared/hugot-model-fields";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const AGGREGATION_FUNCTIONS = ["SOFTMAX", "SIGMOID"] as const;
export type AggregationFunction = (typeof AGGREGATION_FUNCTIONS)[number];

const configSchema = hugotModelSchema.extend({
  aggregation_function: z.enum(AGGREGATION_FUNCTIONS),
  multi_label: z.boolean(),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  ...hugotModelDefaults,
  aggregation_function: "SOFTMAX",
  multi_label: false,
};

const component: FlowComponent<Config> = {
  id: "nlp_classify_text",
  name: "NLP Classify Text",
  category: "processor",
  description:
    "Run a text classification model and tag the message with predicted labels.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
