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

  Editor: ComponentType<EditorProps<TConfig>>;
}
