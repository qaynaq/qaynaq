import { lazy } from "react";
import { z } from "zod";
import type { FlowScanner } from "../types";

const configSchema = z.object({
  pattern: z.string().min(1, "Required"),
  max_buffer_size: z.number().int().min(1),
});
export type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  pattern: "",
  max_buffer_size: 65536,
};

const scanner: FlowScanner<Config> = {
  id: "re_match",
  name: "Regex Match",
  description: "Split the stream wherever a regular expression matches.",
  configSchema,
  defaultConfig,
  Editor: lazy(() => import("./editor")),
};

export default scanner;
