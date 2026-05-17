import type { z } from "zod";
import type { ComponentType } from "react";

export type ComponentCategory =
  | "input"
  | "processor"
  | "output"
  | "cache"
  | "buffer"
  | "rate_limit";

export interface EditorProps<TConfig> {
  value: TConfig;
  onChange: (next: TConfig) => void;
  errors?: Record<string, string>;
  previewMode?: boolean;
  availableProcessors?: FlowComponent<unknown>[];
  availableInputs?: FlowComponent<unknown>[];
  availableOutputs?: FlowComponent<unknown>[];
}

export interface FlowComponent<TConfig> {
  id: string;
  name: string;
  category: ComponentCategory;
  description?: string;

  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;

  parse(yamlStr: string): TConfig;
  serialize(config: TConfig): string;

  // Optional hooks for components that appear inside processor_list,
  // output_list, input_list or output_cases slots. Default behaviour places
  // the parsed config object directly under the component id key. Override
  // these for components whose nested-list shape differs from their
  // standalone shape - e.g. a flat string-valued component like `mapping`.
  toListItem?(config: TConfig): unknown;
  fromListItem?(raw: unknown): TConfig;

  Editor: ComponentType<EditorProps<TConfig>>;
}
