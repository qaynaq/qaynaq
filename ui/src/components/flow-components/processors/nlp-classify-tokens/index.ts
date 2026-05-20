import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import {
  hugotModelDefaults,
  hugotModelSchema,
} from "../../shared/hugot-model-fields";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const AGGREGATION_STRATEGIES = ["SIMPLE", "NONE"] as const;
export type AggregationStrategy = (typeof AGGREGATION_STRATEGIES)[number];

const configSchema = hugotModelSchema.extend({
  aggregation_strategy: z.enum(AGGREGATION_STRATEGIES),
  ignore_labels: z.array(z.string()),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  ...hugotModelDefaults,
  aggregation_strategy: "SIMPLE",
  ignore_labels: [],
};

const component: FlowComponent<Config> = {
  id: "nlp_classify_tokens",
  name: "NLP Classify Tokens",
  category: "processor",
  description:
    "Run a token classification model to extract named entities or similar token-level labels.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
