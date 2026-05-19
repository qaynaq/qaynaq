import type { ComponentType } from "react";
import type { z } from "zod";

export interface ScannerEditorProps<TConfig> {
  value: TConfig;
  onChange: (next: TConfig) => void;
  errors?: Record<string, string>;
}

export interface FlowScanner<TConfig> {
  id: string;
  name: string;
  description?: string;
  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;
  Editor: ComponentType<ScannerEditorProps<TConfig>>;
}
