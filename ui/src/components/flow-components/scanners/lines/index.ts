import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({
  custom_delimiter: z.string(),
  max_buffer_size: z.number().int().min(1),
  omit_empty: z.boolean(),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  custom_delimiter: "",
  max_buffer_size: 65536,
  omit_empty: false,
};

const scanner: FlowScanner<Config> = {
  id: "lines",
  name: "Lines",
  description: "Split the stream into one message per line.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
